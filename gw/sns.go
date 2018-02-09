package gw

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/thingful/big-iot-gateway/pkg/log"
)

const (
	subConfirmType   = "SubscriptionConfirmation"
	notificationType = "Notification"
)

func subscribeHandler(w http.ResponseWriter, r *http.Request) {
	var f interface{}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Fprint(w, err)
		log.Log("error", err)
		return
	}
	err = json.Unmarshal(body, &f)
	if err != nil {
		fmt.Fprint(w, err)
		log.Log("error", err)
		return
	}
	data := f.(map[string]interface{})
	log.Log("debug", data["Type"].(string))

	if data["Type"].(string) == subConfirmType {
		subscribeURL := data["SubscribeURL"].(string)
		go confirmSubscription(subscribeURL)
	} else {
		log.Log("debug", data["Message"].(string))
	}
	fmt.Fprintf(w, "Success")
}

func confirmSubscription(subscribeURL string) {
	resp, err := http.Get(subscribeURL)
	if err != nil {
		log.Log("error", err)
	} else {
		log.Log("msg", fmt.Sprintf("Subscription Confirmed:%d", resp.StatusCode))
	}
}
