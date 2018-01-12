package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/thingful/bigiot"
	goji "goji.io"
	"goji.io/pat"

	"github.com/thingful/big-iot-gateway/utils"
)

const (
	marketplaceURI           = "https://market.big-iot.org"
	providerID               = "thingful_test_org-thingful_test_provider"
	providerSecret           = "3cKoYLd-RdyaB5EZZov7Sg=="
	offeringActiveLengthSec  = 300
	offeringCheckIntervalSec = 10
	offeringEndpoint         = "https://ec2-35-157-149-71.eu-central-1.compute.amazonaws.com:8888/bigiot/access/airqualitydata"
	pipeAccessToken          = "f94a62e6-455f-4a5e-8f7a-36f000cace4d"

	ngrokForward   = true
	forwardAddress = "http://8d43fe3f.ngrok.io"
	defaultHost    = "localhost"
	defaultPort    = "8080"
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

func main() {
	addCommonOutputToOfferings(offerings)

	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	host := os.Getenv("HOST")
	if host == "" {
		host = defaultHost
	}

	//M1 authenticate , register provider
	provider, err := authenticateProvider()
	if err != nil {
		panic(err)
	}
	fmt.Println("M1 authenticateProvider success")

	// M2 register
	for _, o := range offerings {
		off := makeOffering(o, host, port)
		// spew.Dump(off)
		_, err = provider.RegisterOffering(context.Background(), off)
		if err != nil {
			panic(err) // handle error properly
		}
		go offeringCheck(o, provider, host, port)
	}
	fmt.Println("M2 register offering completed")

	mux := goji.NewMux()
	mux.HandleFunc(pat.Get("/offering/:offeringID"), access)

	log.Printf("Starting server listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))

}

func access(w http.ResponseWriter, r *http.Request) {
	offeringID := pat.Param(r, "offeringID")
	fmt.Printf("incoming request for: %s\n", offeringID)
	index := utils.GetOfferingIndex(offeringID, offerings)
	if index == -1 { // we check if the path is valid, if not return 404
		w.WriteHeader(404)
		return
	}

	// then we try to call pipe
	pipeURL := offerings[index].PipeURL
	pipeJSON, err := utils.MakePipeRequest(pipeURL, pipeAccessToken)
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

}

func addCommonOutputToOfferings(o []utils.OfferingConfig) {
	for i := range o {
		o[i].Outputs = append(o[i].Outputs, commonOutputs...)
	}

}

// the first register could also happen here
func offeringCheck(offering utils.OfferingConfig, provider *bigiot.Provider, host string, port string) {

	for range time.Tick(time.Second * offeringCheckIntervalSec) {
		fmt.Printf("now we check for offering:%s\n", offering.Name)
		bytes, err := utils.MakePipeRequest(offering.PipeURL+"?limit=1", pipeAccessToken)
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

func authenticateProvider() (*bigiot.Provider, error) {
	provider, err := bigiot.NewProvider(
		providerID,
		providerSecret,
		bigiot.WithMarketplace(marketplaceURI),
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

	if ngrokForward {
		addOfferingInput.Endpoints[0].URI = fmt.Sprintf("%s/offering/%s", forwardAddress, strings.ToLower(o.ID))
	}

	return addOfferingInput
}

func unregisterOnExit(provider *bigiot.Provider, offering *bigiot.Offering) {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		fmt.Println("exit")
		deleteOfferingInput := &bigiot.DeleteOffering{
			ID: offering.ID,
		}

		err := provider.DeleteOffering(context.Background(), deleteOfferingInput)
		if err != nil {
			panic(err) // handle error properly
		}
		fmt.Println("unregister complete")
		os.Exit(0)
	}()
}
