# BIG-IoT-Gateway

Minimal big-iot gateway using BIG-IOT Golang SDK and Thingful Pipe

## How to Build

```
Please use 'make <target>' where <target> is one of
build-darwin                   Build a darwin x64 executable
build                          Build a linux x64 executable
clean                          Remove all generated artefacts
release                        Create target container with the release version of the app
run                            Run binary on docker container
```

## Usage
```
BIG-IoT gateway

Usage:
  big-iot-gw [command]

Available Commands:
  help        Help about any command
  start       Start BIG-IoT Gateway

Flags:
      --config string                  Config file (default is ./config.yaml)
  -h, --help                           help for big-iot-gw
      --marketPlaceURI string          Main URI for BIG-IoT Market Place (default "https://market.big-iot.org")
      --offeringActiveLengthSec int    Offering Active Length Sec (default 300)
      --offeringCheckIntervalSec int   Offering Check Interval in secs (default 10)
      --offeringEndpoint string        Offering End Point
      --pipeAccessToken string         Pipes access token
      --providerID string              Provider ID for BIG-IoT MarketPlace

Use "big-iot-gw [command] --help" for more information about a command.
```

The app can be configured using these methods ordered by their precedence:
1. Env Vars 
2. Flags
3. Configuration file

Example for `offeringEndpoint` var with a file with the next content:
```
offeringEndpoint: http://url1
```
Running:
`OFFERINGENDPOINT=http://url2 big-iot-gateway start --config config.yaml`

Will override the value http://url1 contained in the config file and it will change the value to http://url2

Running:
`big-iot-gateway start --config config.yaml --offeringEndpoint=http://url3`

Will override the value http://url1 contained in the config file and it will change the value to http://url3

Running:
`big-iot-gateway start --offeringEndpoint=http://url4`

Will run **without consider the config file at all** and  must be required to specified all the flags or env vars explicitly. 



## 
