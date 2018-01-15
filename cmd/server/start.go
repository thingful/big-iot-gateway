package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"goji.io"
	"goji.io/pat"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/thingful/big-iot-gateway/pkg/utils"
	"github.com/thingful/bigiot"
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

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start BIG-IoT Gateway",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("start called")
		fmt.Println("offeringEndpoint:", viper.Get("offeringEndpoint"))
		// A goroutine listening for a ctrl+c signal
		go func() {
			<-exitChan
			fmt.Println("Exiting...")
			// Clean something?
			os.Exit(1)
		}()
		// For refactor
		addCommonOutputToOfferings(offerings)

		provider, err := authenticateProvider()
		if err != nil {
			panic(err)
		}
		for _, o := range offerings {
			off := makeOffering(o, "localhost", "8080")
			_, err = provider.RegisterOffering(context.Background(), off)
			if err != nil {
				panic(err)
			}
			go offeringCheck(o, provider, "localhost", "8080")
		}
		mux := goji.NewMux()

		mux.HandleFunc(pat.Get("/offering/:offeringID"), func(w http.ResponseWriter, r *http.Request) {
			offeringID := pat.Param(r, "offeringID")
			fmt.Printf("incoming request for: %s\n", offeringID)
			index := utils.GetOfferingIndex(offeringID, offerings)
			if index == -1 { // we check if the path is valid, if not return 404
				w.WriteHeader(404)
				return
			}

			// then we try to call pipe
			pipeURL := offerings[index].PipeURL
			pipeJSON, err := utils.MakePipeRequest(pipeURL, viper.GetString("pipeAccessToken"))
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
			io.WriteString(w, string(bigiotJSON))
		})

		log.Fatal(http.ListenAndServe(":8080", mux))
	},
}

func init() {
	RootCmd.AddCommand(startCmd)
}

func addCommonOutputToOfferings(o []utils.OfferingConfig) {
	for i := range o {
		o[i].Outputs = append(o[i].Outputs, commonOutputs...)
	}

}

func authenticateProvider() (*bigiot.Provider, error) {
	provider, err := bigiot.NewProvider(
		viper.GetString("providerID"),
		viper.GetString("providerSecret"),
		bigiot.WithMarketplace(viper.GetString("marketPlaceURI")),
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

func makeOffering(o utils.OfferingConfig, host string, port string) *bigiot.OfferingDescription {
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
			ExpirationTime: time.Now().Add(viper.GetDuration("offeringActiveLengthSec") * time.Second), // need to set this
		},
	}
	for _, output := range o.Outputs {
		d := bigiot.DataField{
			Name:   output.BigiotName,
			RdfURI: output.BigiotRDF,
		}
		addOfferingInput.OutputData = append(addOfferingInput.OutputData, d)
	}
	/*
		if ngrokForward {
			addOfferingInput.Endpoints[0].URI = fmt.Sprintf("%s/offering/%s", forwardAddress, strings.ToLower(o.ID))
		}
	*/

	return addOfferingInput
}

// the first register could also happen here
func offeringCheck(offering utils.OfferingConfig, provider *bigiot.Provider, host string, port string) {

	for range time.Tick(time.Second * viper.GetDuration("offeringCheckIntervalSec")) {
		fmt.Printf("now we check for offering:%s\n", offering.Name)
		bytes, err := utils.MakePipeRequest(offering.PipeURL+"?limit=1", viper.GetString("pipeAccessToken"))
		if err != nil {
			panic(err)
		}

		// we unmarshal the response, check number of result
		var m interface{}
		err = json.Unmarshal(bytes, &m)
		if err != nil {
			panic(err)
		}

		j := m.([]interface{}) //type case to slice first
		if len(j) == 1 {

			fmt.Printf("pipe for offering: %s return 1 result, re-registering offering:", offering.Name)

			off := makeOffering(offering, host, port)
			// spew.Dump(off)
			_, err = provider.RegisterOffering(context.Background(), off)
			if err != nil {
				panic(err)
			}

			fmt.Printf(" COMPLETED\n")

		} else {
			// delete offering from marketplace
			fmt.Printf("pipe for offering: %s return 0 result, deleting offering\n", offering.Name)
			deleteOfferingInput := &bigiot.DeleteOffering{
				ID: offering.ID,
			}
			err := provider.DeleteOffering(context.Background(), deleteOfferingInput)
			if err != nil {
				panic(err)
			}
			fmt.Printf(" COMPLETED\n")
		}
	}
}
