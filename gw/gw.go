package gw

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/signal"

	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/spf13/viper"
	"github.com/thingful/big-iot-gateway/pkg/log"
	"github.com/thingful/big-iot-gateway/pkg/middleware"
	"github.com/thingful/big-iot-gateway/pkg/pipes"
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
			BigiotRDF:  "http://schema.org/attribution",
			PipeTerm:   "provider.name",
		},
	}
)

// Start starts gw service
func Start(config Config, offers []Offer) error {
	stop := make(chan os.Signal)
	signal.Notify(stop, os.Interrupt)

	addCommonOutputToOfferings(offers)

	if config.Debug {
		log.Log("Settings", viper.AllSettings())
	}

	provider, err := authenticateProvider(config.ProviderID, config.ProviderSecret, config.MarketPlaceURI)
	if err != nil {
		return err
	}

	offeringEndpoint, err := url.Parse(config.OfferingEndPoint)
	if err != nil {
		return err
	}

	offerings := []*bigiot.Offering{}

	for _, o := range offers {
		offeringDescription := makeOfferingInput(o, offeringEndpoint.Host, config.OfferingActiveLengthSec)
		offering, err := provider.RegisterOffering(context.Background(), offeringDescription)
		if err != nil {
			log.Log("msg", "Error Registering Offering:", err)
		}

		offerings = append(offerings, offering)

		go func() {
			err := offeringCheck(o, provider, offeringEndpoint.Host, config.PipeAccessToken, config.OfferingCheckIntervalSec)
			log.Log("msg", "Error checking Offering:", "error", err)
		}()
	}

	mux := goji.NewMux()

	mux.HandleFunc(pat.Get("/pulse"), pulse)

	if !config.NoAuth {
		log.Log("msg", "ading auth mddleware")
		auth, err := middleware.NewAuth(provider)
		if err != nil {
			return err
		}
		mux.Use(auth.Handler)
	} else {
		log.Log("msg", "no auth")
	}

	mux.HandleFunc(pat.Get("/offering/:offeringID"), func(w http.ResponseWriter, r *http.Request) {
		offeringID := pat.Param(r, "offeringID")
		log.Log("offeringID", offeringID, "msg", "incoming request")
		index := getOfferingIndex(offeringID, offers)
		if index == -1 { // we check if the path is valid, if not return 404
			w.WriteHeader(404)
			return
		}

		log.Log("pipeURL", offers[index].PipeURL, "token", config.PipeAccessToken)

		// then we try to call pipe
		pipeURL := offers[index].PipeURL
		pipeJSON, err := pipes.MakeRequest(pipeURL, config.PipeAccessToken)
		if err != nil {
			log.Log("error", err)
			w.WriteHeader(500)
			return
		}

		// now we reformat our json to their json
		bigiotJSON, err := ConvertJSON(pipeJSON, offers[index])
		if err != nil {
			log.Log("error", err)
			w.WriteHeader(500)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if _, err := io.WriteString(w, string(bigiotJSON)); err != nil {
			log.Log("error", err)
		}
	})

	srv := &http.Server{Addr: fmt.Sprintf(":%d", config.HTTPPort), Handler: mux}

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Log("port", config.HTTPPort, "msg", "listening")
		}
	}()

	<-stop
	log.Log("msg", "shutting down, removing offerings")

	// range over offerings and remove them all from marketplace
	for _, o := range offerings {
		deleteOffering := &bigiot.DeleteOffering{
			ID: o.ID,
		}

		err = provider.DeleteOffering(context.Background(), deleteOffering)
		if err != nil {
			return err
		}
	}

	srv.Shutdown(context.Background())

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

func makeOfferingInput(o Offer, host string, offeringActiveLengthSec time.Duration) *bigiot.OfferingDescription {
	var pricingModel bigiot.PricingModel

	if o.Price > 0 {
		pricingModel = bigiot.PerAccess
	} else {
		pricingModel = bigiot.Free
	}

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
				URI:                 fmt.Sprintf("http://%s/offering/%s", host, strings.ToLower(o.ID)),
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
				Amount:   o.Price,
				Currency: bigiot.EUR,
			},
			PricingModel: pricingModel,
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
	pipeAccessToken string,
	offeringCheckIntervalSec time.Duration) error {

	ticker := time.NewTicker(time.Second * offeringCheckIntervalSec)
	for range ticker.C {
		//log.Log("debug", "", "now we check for offering:", offering.Name, "pipeURL", offering.PipeURL)
		//log.Log("pipeAccessToken", pipeAccessToken)
		bytes, err := pipes.MakeRequest(offering.PipeURL+"?limit=1", pipeAccessToken)
		if err != nil {
			return err
		}

		fmt.Println(string(bytes))
		// we unmarshal the response, check number of result
		var m interface{}
		err = json.Unmarshal(bytes, &m)
		if err != nil {
			return err
		}

		j := m.([]interface{}) //type case to slice first
		if len(j) == 1 {
			offeringDescription := makeOfferingInput(offering, host, offeringCheckIntervalSec)
			_, err := provider.RegisterOffering(context.Background(), offeringDescription)
			if err != nil {
				return err
			}

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
	log.Log("index", offeringIndex)
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
