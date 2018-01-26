package gw

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/spf13/cast"

	"net/http"
	"strings"
	"time"

	"github.com/spf13/viper"
	"github.com/thingful/big-iot-gateway-test/utils"
	"github.com/thingful/big-iot-gateway/pkg/log"
	"github.com/thingful/big-iot-gateway/pkg/middleware"
	"github.com/thingful/bigiot"
	goji "goji.io"
	"goji.io/pat"
)

var (
	// output that every offerings will have
	commonOutputs = []Output{
		Output{
			BigiotName: "latitude",
			BigiotRDF:  "http://schema.org/latitude",
			PipeTerm:   "latitude",
		},
		Output{
			BigiotName: "longitude",
			BigiotRDF:  "http://schema.org/longitude",
			PipeTerm:   "longitude",
		},
		Output{
			BigiotName: "attribution",
			BigiotRDF:  "http://xxx/yyy/zzz",
			PipeTerm:   "provider.name",
		},
	}
)

// Start starts gw service
func Start(config Config, offerings []Offer) error {

	HTTPPort := cast.ToString(config.HTTPPort)

	addCommonOutputToOfferings(offerings)

	if config.Debug {
		log.Log("Debug", "", "Settings", viper.AllSettings())
	}

	provider, err := authenticateProvider(config.ProviderID, config.ProviderSecret, config.MarketPlaceURI)
	if err != nil {
		return err
	}

	for _, o := range offerings {
		off := makeOffering(o, config.HTTPHost, HTTPPort, config.OfferingActiveLengthSec)
		_, err = provider.RegisterOffering(context.Background(), off)
		if err != nil {
			log.Log("msg", "Error Registering Offering:", err)
		}

		go func() {
			err := offeringCheck(o, provider, config.HTTPHost, HTTPPort, config.PipeAccessToken, config.OfferingCheckIntervalSec)
			log.Log("debug", "", "Error checking Offering:", err)
		}()
	}

	auth, err := middleware.NewAuth(provider)
	if err != nil {
		return (err)
	}

	mux := goji.NewMux()

	mux.HandleFunc(pat.Get("/pulse"), pulse)
	mux.Use(auth.Handler)

	mux.HandleFunc(pat.Get("/offering/:offeringID"), func(w http.ResponseWriter, r *http.Request) {
		offeringID := pat.Param(r, "offeringID")
		log.Log("msg", "incoming request for: ", offeringID)
		index := getOfferingIndex(offeringID, offerings)
		if index == -1 { // we check if the path is valid, if not return 404
			w.WriteHeader(404)
			return
		}

		// then we try to call pipe
		pipeURL := offerings[index].PipeURL
		pipeJSON, err := utils.MakePipeRequest(pipeURL, config.PipeAccessToken)
		if err != nil {
			w.WriteHeader(500)
			return
		}

		// now we reformat our json to their json
		bigiotJSON, err := ConvertJSON(pipeJSON, offerings[index])
		if err != nil {
			w.WriteHeader(500)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if _, err := io.WriteString(w, string(bigiotJSON)); err != nil {
			log.Log(err)
		}
	})

	log.Fatal("Fatal", http.ListenAndServe(fmt.Sprintf(":%d", config.HTTPPort), mux))

	return nil
}

func addCommonOutputToOfferings(o []Offer) {
	for i := range o {
		o[i].Outputs = append(o[i].Outputs, commonOutputs...)
	}

}

func authenticateProvider(id, secret, uri string) (*bigiot.Provider, error) {
	provider, err := bigiot.NewProvider(
		id,
		secret,
		bigiot.WithMarketplace(uri),
	)
	if err != nil {
		return nil, err
	}

	err = provider.Authenticate()
	if err != nil {
		return nil, err
	}

	return provider, err
}

func makeOffering(o Offer, host string, port string, offeringActiveLengthSec time.Duration) *bigiot.OfferingDescription {
	addOfferingInput := &bigiot.OfferingDescription{
		LocalID: o.ID,
		Name:    o.Name,
		RdfURI:  o.Category,
		InputData: []bigiot.DataField{
			{
				Name:   "latitude",
				RdfURI: "http://schema.org/latitude",
			},
			{
				Name:   "longitude",
				RdfURI: "http://schema.org/longitude",
			},
			{
				Name:   "geoRadius",
				RdfURI: "http://schema.org/geoRadius",
			},
		},
		Endpoints: []bigiot.Endpoint{
			{
				URI:                 fmt.Sprintf("http://%s:%s/offering/%s", host, port, strings.ToLower(o.ID)),
				EndpointType:        bigiot.HTTPGet,
				AccessInterfaceType: bigiot.BIGIoTLib,
			},
		},
		License: bigiot.OpenDataLicense,
		Extent: bigiot.Address{
			City: o.City,
		},
		Price: bigiot.Price{
			Money: bigiot.Money{
				Amount:   0,
				Currency: bigiot.EUR,
			},
			PricingModel: bigiot.Free,
		},
		Activation: bigiot.Activation{
			Status:         true,
			ExpirationTime: time.Now().Add(offeringActiveLengthSec * time.Second), // need to set this
		},
	}
	for _, output := range o.Outputs {
		d := bigiot.DataField{
			Name:   output.BigiotName,
			RdfURI: output.BigiotRDF,
		}
		addOfferingInput.OutputData = append(addOfferingInput.OutputData, d)
	}

	return addOfferingInput
}

// the first register could also happen here
func offeringCheck(
	offering Offer,
	provider *bigiot.Provider,
	host string,
	port string,
	pipeAccessToken string,
	offeringCheckIntervalSec time.Duration) error {

	ticker := time.NewTicker(time.Second * offeringCheckIntervalSec)
	for range ticker.C {
		//log.Log("debug", "", "now we check for offering:", offering.Name, "pipeURL", offering.PipeURL)
		//log.Log("pipeAccessToken", pipeAccessToken)
		bytes, err := utils.MakePipeRequest(offering.PipeURL+"?limit=1", pipeAccessToken)
		if err != nil {
			return err
		}

		// we unmarshal the response, check number of result
		var m interface{}
		err = json.Unmarshal(bytes, &m)
		if err != nil {
			return err
		}

		j := m.([]interface{}) //type case to slice first
		if len(j) == 1 {
			//Debug
			//log.Log("msg", "pipe for offering: ", offering.Name, " return 1 result, re-registering offering:")
			off := makeOffering(offering, host, port, offeringCheckIntervalSec)
			_, err = provider.RegisterOffering(context.Background(), off)
			if err != nil {
				return err
			}
			//Debug
			//log.Log("msg", " COMPLETED")

		} else {
			// delete offering from marketplace
			log.Log("msg", offering.Name+" returns 0 result, deleting offering", offering.Name)

			deleteOfferingInput := &bigiot.DeleteOffering{
				ID: offering.ID,
			}
			err := provider.DeleteOffering(context.Background(), deleteOfferingInput)
			if err != nil {
				return err
			}
			//log.Log("msg", " COMPLETED")
		}
	}
	return nil
}

func pulse(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "ok")
}

func getOfferingIndex(id string, offerings []Offer) int {
	offeringIndex := -1
	for i, offering := range offerings {
		if id == strings.ToLower(offering.ID) {
			offeringIndex = i
			break
		}
	}
	return offeringIndex
}

// ConvertJSON takes pipe json and change to big-iot json depends on offerinConfig provide
func ConvertJSON(pipeJson []byte, offering Offer) ([]byte, error) {

	output := []map[string]interface{}{}

	var m interface{}
	err := json.Unmarshal(pipeJson, &m)
	if err != nil {
		return nil, err
	}

	j := m.([]interface{}) //type case to slice first

	for _, member := range j {
		pipeData := member.(map[string]interface{}) // then for each member, cast to map string interface

		bigiotData := map[string]interface{}{} // make temporary var

		for _, output := range offering.Outputs {
			if val, ok := pipeData[output.PipeTerm]; ok { // find if the key exist, if it does assign it
				bigiotData[output.BigiotName] = val
			} else { // if it doesn't exist, assing default value
				bigiotData[output.BigiotName] = ""
			}
		}

		output = append(output, bigiotData)
	}

	s, err := json.Marshal(output)
	if err != nil {
		return nil, err
	}

	return s, nil
}
