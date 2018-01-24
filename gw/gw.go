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
	"github.com/thingful/bigiot"
	goji "goji.io"
	"goji.io/pat"
)

var (
	// output that every offerings will have
	commonOutputs = []utils.OfferingOutput{
		utils.OfferingOutput{
			BigiotName: "latitude",
			BigiotRDF:  "http://schema.org/latitude",
			PipeTerm:   "latitude",
		},
		utils.OfferingOutput{
			BigiotName: "longitude",
			BigiotRDF:  "http://schema.org/longitude",
			PipeTerm:   "longitude",
		},
		utils.OfferingOutput{
			BigiotName: "attribution",
			BigiotRDF:  "http://xxx/yyy/zzz",
			PipeTerm:   "provider.name",
		},
	}

	offerings = []utils.OfferingConfig{ // this is where we define the offerings
		utils.OfferingConfig{
			ID:          "torinoWeather",
			Name:        "Torino Weather",
			City:        "Torino",
			PipeURL:     "https://thingful-pipes.herokuapp.com/api/run/1b9cfeb3-c741-4673-ac5e-49c5ec3f7753",
			Category:    "http://schema.org/environmental",
			Datalicense: "CCBySAV4URL",
			Outputs: []utils.OfferingOutput{
				utils.OfferingOutput{
					BigiotName: "airTemperatureValue",
					BigiotRDF:  "schema:airTemperatureValue",
					PipeTerm:   "Air Temperature, Weather Temperature, Ambient Temperature",
				},
				utils.OfferingOutput{
					BigiotName: "airHumidityValue",
					BigiotRDF:  "schema:airHumidityValue",
					PipeTerm:   "Humidity",
				},
			},
		},
	}
)

// Start starts gw service
func Start(config Config) error {
	addCommonOutputToOfferings(offerings)

	if config.Debug {
		log.Log("Debug", "", "Settings", viper.AllSettings())
	}

	provider, err := authenticateProvider(config.ProviderID, config.ProviderSecret, config.MarketPlaceURI)
	if err != nil {
		return err
	}

	for _, o := range offerings {
		off := makeOffering(o, config.HTTPHost, cast.ToString(config.HTTPPort), config.OfferingActiveLengthSec)
		_, err = provider.RegisterOffering(context.Background(), off)
		if err != nil {
			log.Log("msg", "Error Registering Offering:", err)
		}

		go func() {
			err := offeringCheck(o, provider, "localhost", "8081", config.PipeAccessToken, config.OfferingCheckIntervalSec)
			log.Log("debug", "", "Error checking Offering:", err)
		}()
	}
	mux := goji.NewMux()

	mux.HandleFunc(pat.Get("/offering/:offeringID"), func(w http.ResponseWriter, r *http.Request) {
		offeringID := pat.Param(r, "offeringID")
		log.Log("msg", "incoming request for: ", offeringID)
		index := utils.GetOfferingIndex(offeringID, offerings)
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
		bigiotJSON, err := utils.ConvertJSON(pipeJSON, offerings[index])
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

func addCommonOutputToOfferings(o []utils.OfferingConfig) {
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

func makeOffering(o utils.OfferingConfig, host string, port string, offeringActiveLengthSec time.Duration) *bigiot.OfferingDescription {
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
	offering utils.OfferingConfig,
	provider *bigiot.Provider,
	host string,
	port string,
	pipeAccessToken string,
	offeringCheckIntervalSec time.Duration) error {

	for range time.Tick(time.Second * offeringCheckIntervalSec) {
		log.Log("debug", "", "now we check for offering:", offering.Name, "pipeURL", offering.PipeURL)
		log.Log("pipeAccessToken", pipeAccessToken)
		bytes, err := utils.MakePipeRequest(offering.PipeURL+"?limit=1", pipeAccessToken)
		if err != nil {
			return err
		}
		log.Log("bytes", bytes)
		// we unmarshal the response, check number of result
		var m interface{}
		err = json.Unmarshal(bytes, &m)
		if err != nil {
			return err
		}

		j := m.([]interface{}) //type case to slice first
		if len(j) == 1 {

			log.Log("msg", "pipe for offering: ", offering.Name, " return 1 result, re-registering offering:")

			off := makeOffering(offering, host, port, offeringCheckIntervalSec)
			// spew.Dump(off)
			_, err = provider.RegisterOffering(context.Background(), off)
			if err != nil {
				return err
			}

			log.Log("msg", " COMPLETED\n")

		} else {
			// delete offering from marketplace
			log.Log("msg", "pipe for offering: %s return 0 result, deleting offering\n", offering.Name)
			deleteOfferingInput := &bigiot.DeleteOffering{
				ID: offering.ID,
			}
			err := provider.DeleteOffering(context.Background(), deleteOfferingInput)
			if err != nil {
				return err
			}
			log.Log("msg", " COMPLETED\n")
		}
	}
	return nil
}
