package gw

import (
	"errors"
	"os"
	"time"

	"github.com/spf13/cast"
)

// Config contains the required configuration for the gateway
type Config struct {
	MarketPlaceURI           string        // Needed to manage offers
	ProviderID               string        // Needed to login into Marketplace
	ProviderSecret           string        // Needed to login into Marketplace
	OfferingActiveLengthSec  time.Duration // timeout
	OfferingCheckIntervalSec time.Duration // Offering Check interval
	OfferingEndPoint         string
	PipeAccessToken          string // Token to access pipes
	HTTPPort                 int    // GW port
	HTTPHost                 string // GW Host
	Debug                    bool   // Debug Flag
	NoAuth                   bool   // disable auth flag
}

// NewConfig return a new Config
func NewConfig() Config {
	return Config{}
}

// Load function load configuration from a map and can be used to
// read from viper.AllSettings() function
// if some needed setting doesn't exist it returns an error
func (c *Config) Load(conf map[string]interface{}) error {
	if val, ok := conf["marketplaceuri"]; ok {
		c.MarketPlaceURI = cast.ToString(val)
	} else {
		return errors.New("marketPlaceURI is not set")
	}
	if val, ok := conf["providerid"]; ok {
		c.ProviderID = cast.ToString(val)
	} else {
		return errors.New("providerID is not set")
	}
	if val, ok := conf["providersecret"]; ok {
		c.ProviderSecret = cast.ToString(val)
	} else {
		return errors.New("providerSecret is not set")
	}
	if val, ok := conf["offeringactivelengthsec"]; ok {
		c.OfferingActiveLengthSec = cast.ToDuration(val)
	} else {
		return errors.New("offeringactivelengthsec is not set")
	}
	if val, ok := conf["offeringcheckintervalsec"]; ok {
		c.OfferingCheckIntervalSec = cast.ToDuration(val)
	} else {
		return errors.New("offeringcheckintervalsec is not set")
	}
	if val, ok := conf["offeringendpoint"]; ok {
		c.OfferingEndPoint = cast.ToString(val)
	} else {
		return errors.New("offeringserver is not set")
	}
	if val, ok := conf["pipeaccesstoken"]; ok {
		c.PipeAccessToken = cast.ToString(val)
	} else {
		return errors.New("pipeAccessToken is not set")
	}
	if val, ok := conf["httpport"]; ok && cast.ToInt(val) != 0 {
		c.HTTPPort = cast.ToInt(val)
	} else {
		// if it's running in heroku
		c.HTTPPort = cast.ToInt(os.Getenv("PORT"))
		//fmt.Println("PORT=", cast.ToInt(os.Getenv("PORT")))
	}
	if c.HTTPPort == 0 {
		return errors.New("httpport/PORT is not set")
	}

	if val, ok := conf["httphost"]; ok {
		c.HTTPHost = cast.ToString(val)
	} else {
		return errors.New("httphost is not set")
	}
	if val, ok := conf["debug"]; ok {
		c.Debug = cast.ToBool(val)
	}
	if val, ok := conf["noauth"]; ok {
		c.NoAuth = cast.ToBool(val)
	}
	return nil
}
