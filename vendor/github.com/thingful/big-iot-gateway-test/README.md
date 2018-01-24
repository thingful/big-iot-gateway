# big-iot-gateway-test

Minimal big-iot gateway using BIG-IOT Golang SDK and Thingful Pipe

## How to use it
* `make help`
To show the available commands help
* `make install`
To install all the required dependencies to build the app
* `make build`
To build the executable app under folder build/
* `make release`
To create a target container with the release version of the app
* `make run`
To run the app into a local container and will register [this offering](https://market.big-iot.org/offering/thingful_test_org-thingful_test_provider-torinoWeather) to big-iot marketplace and serve the endpoint at `http://localhost:8080/offering/torinoweather `

## logic

Single gateway, single endpoint but take different argument? eg; `/offering/torinotraffic`

### upon startup
* authenticate itself
* look for the config, which will have a bunch of offering settings
* register each offerings on the marketplace
    * this will make each offering valid for XXX mins
* start single endpoint, each offering is separated by path

### On request
* get the `offeringID` from path, validate the offering name
* then call associated pipe, and process the result differently depends on the offering config
* if there is problem with offering, delete it from marketplace

### On interval
* try accessing each offering's pipe, if OK re-register to extend the activation period for XXX mins
    * this health check result can be limit to 1
* if not OK delete the offering from the marketplace

## note:
* that means free account on Heroku doesn't work
