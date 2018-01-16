package gw

import (
	"errors"
	"time"

	"github.com/spf13/cast"
)

// Config contains the required configuration for the gateway
type Config struct {
	MarketPlaceURI           string
	ProviderID               string
	ProviderSecret           string
	OfferingActiveLengthSec  time.Duration
	OfferingCheckIntervalSec time.Duration
	OfferingEndPoint         string
	PipeAccessToken          string
}

// NewConfig return a new Config
func NewConfig() Config {
	return Config{}
}

// Load function load configuration from a map and can be used to
// read from viper.AllSettings() function
// if some needed setting doesn't exist it returns an error
func (c *Config) Load(conf map[string]interface{}) error {
	if val, ok := conf["marketPlaceURI"]; ok {
		c.MarketPlaceURI = cast.ToString(val)
	} else {
		return errors.New("marketPlaceURI is not set")
	}
	if val, ok := conf["providerID"]; ok {
		c.ProviderID = cast.ToString(val)
	} else {
		return errors.New("providerID is not set")
	}
	if val, ok := conf["providerSecret"]; ok {
		c.ProviderSecret = cast.ToString(val)
	} else {
		return errors.New("providerSecret is not set")
	}
	if val, ok := conf["offeringActiveLengthSec"]; ok {
		c.OfferingActiveLengthSec = cast.ToDuration(val)
	} else {
		return errors.New("offeringActiveLengthSec is not set")
	}
	if val, ok := conf["offeringCheckIntervalSec"]; ok {
		c.OfferingCheckIntervalSec = cast.ToDuration(val)
	} else {
		return errors.New("offeringCheckIntervalSec is not set")
	}
	if val, ok := conf["offeringEndpoint"]; ok {
		c.OfferingEndPoint = cast.ToString(val)
	} else {
		return errors.New("offeringEndpoint is not set")
	}
	if val, ok := conf["pipeAccessToken"]; ok {
		c.PipeAccessToken = cast.ToString(val)
	} else {
		return errors.New("pipeAccessToken is not set")
	}
	return nil
}
