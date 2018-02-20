package main

import (
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/thingful/big-iot-gateway/pkg/log"
)

var (
	cfgFile   string
	offerFile string
	offers    *viper.Viper
	awsCreds  *AwsCreds
)

var RootCmd = &cobra.Command{
	Use:   "big-iot-gw",
	Short: "BIG-IoT gateway",
}

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		log.Log("error", err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "Config file (default is ./config.yaml)")
	RootCmd.PersistentFlags().StringVar(&offerFile, "offerFile", "", "Config file with offerings (default is ./offerings.yaml)")

	// for remote offers config
	awsCreds = &AwsCreds{}
	RootCmd.PersistentFlags().StringVar(&awsCreds.accessKey, "aws_key", "", "AWS Access key id")
	RootCmd.PersistentFlags().StringVar(&awsCreds.secret, "aws_secret", "", "AWS Secret access key")
	RootCmd.PersistentFlags().StringVar(&awsCreds.region, "aws_region", "", "AWS region")

	RootCmd.PersistentFlags().String("marketPlaceURI", "https://market.big-iot.org", "Main URI for BIG-IoT Market Place")
	RootCmd.PersistentFlags().String("providerID", "", "Provider ID for BIG-IoT MarketPlace")
	RootCmd.PersistentFlags().String("providerSecret", "", "Provider Secret for BIG-IoT MarketPlace")
	RootCmd.PersistentFlags().Int("offeringActiveLengthSec", 300, "Offering Active Length Sec")
	RootCmd.PersistentFlags().Int("offeringCheckIntervalSec", 600, "Offering Check Interval in secs")
	RootCmd.PersistentFlags().String("offeringEndpoint", "", "Offering End Point")
	RootCmd.PersistentFlags().String("pipeAccessToken", "", "Pipes access token")
	RootCmd.PersistentFlags().String("mapsKey", "", "API Key for Geocoding locations via Google Maps API")
	RootCmd.PersistentFlags().Int("HTTPPort", 0, "HTTP Port where will be running the service")
	RootCmd.PersistentFlags().String("HTTPHost", "localhost", "HTTP Hostname where will be running the service")
	RootCmd.PersistentFlags().Bool("debug", false, "enable debug")
	RootCmd.PersistentFlags().Bool("noauth", false, "disable auth")

	viper.BindPFlag("marketPlaceURI", RootCmd.PersistentFlags().Lookup("marketPlaceURI"))
	viper.BindPFlag("providerID", RootCmd.PersistentFlags().Lookup("providerID"))
	viper.BindPFlag("providerSecret", RootCmd.PersistentFlags().Lookup("providerSecret"))
	viper.BindPFlag("offeringActiveLengthSec", RootCmd.PersistentFlags().Lookup("offeringActiveLengthSec"))
	viper.BindPFlag("offeringCheckIntervalSec", RootCmd.PersistentFlags().Lookup("offeringCheckIntervalSec"))
	viper.BindPFlag("offeringEndpoint", RootCmd.PersistentFlags().Lookup("offeringEndpoint"))
	viper.BindPFlag("pipeAccessToken", RootCmd.PersistentFlags().Lookup("pipeAccessToken"))
	viper.BindPFlag("mapsKey", RootCmd.PersistentFlags().Lookup("mapsKey"))
	viper.BindPFlag("HTTPPort", RootCmd.PersistentFlags().Lookup("HTTPPort"))
	viper.BindPFlag("HTTPHost", RootCmd.PersistentFlags().Lookup("HTTPHost"))
	viper.BindPFlag("debug", RootCmd.PersistentFlags().Lookup("debug"))
	viper.BindPFlag("noauth", RootCmd.PersistentFlags().Lookup("noauth"))
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	// System Config File
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		viper.SetConfigName("config.yaml")

	}
	// Get Env
	viper.AutomaticEnv()

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		log.Log("msg", "Using config file:"+viper.ConfigFileUsed())
	}

	// Offerings file
	offers = viper.New()
	if err := checkOfferFile(offerFile, offers); err != nil {
		log.Fatal(err)
	}
}

func bindViper(flags *pflag.FlagSet, names ...string) error {
	for _, name := range names {
		err := viper.BindPFlag(name, flags.Lookup(name))
		if err != nil {
			return fmt.Errorf("Error trying to bind flags: %s", err.Error())
		}
	}
	return nil
}

func checkOfferFile(offFileName string, offConfig *viper.Viper) error {

	if offFileName == "" {
		log.Log("msg", "no offer file, trying default offers.json file")
		offConfig.SetConfigName("offers")
		offConfig.AddConfigPath("./")
		return offConfig.ReadInConfig()
	}

	s, err := url.Parse(offFileName)
	if err != nil {
		log.Log("error", err)
		return err
	}

	switch s.Scheme {
	case "s3":
		r, err := getS3Offers(s.Host, strings.Replace(s.Path, "/", "", -1))
		if err != nil {
			return err
		}

		ext := strings.Replace(filepath.Ext(s.Path), ".", "", -1)
		offConfig.SetConfigType(ext)
		log.Log("msg", "getting offers from aws s3")

		return offConfig.ReadConfig(r)

	case "file":
		p, err := getFilePath(s.String())
		if err != nil {
			return err
		}

		offConfig.SetConfigFile(p)
		log.Log("msg", "using local file")

		return offConfig.ReadInConfig()

	default:
		return errors.New("unknown offers file")
	}

}

type AwsCreds struct {
	accessKey string
	secret    string
	region    string
}

func getS3Offers(bucket, file string) (io.Reader, error) {

	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(awsCreds.region),
	})
	if err != nil {
		return nil, err
	}

	svc := s3.New(sess)

	res, err := svc.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(file),
	})
	if err != nil {
		return nil, err
	}

	return io.Reader(res.Body), nil
}

func getFilePath(path string) (string, error) {
	p := strings.Split(path, "//")
	if len(p) < 2 {
		return "", errors.New("file path string is not well formed")
	}
	return p[1], nil
}
