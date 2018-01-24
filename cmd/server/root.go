package main

import (
	"os"
	"os/signal"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/thingful/big-iot-gateway/pkg/log"
)

var (
	cfgFile  string
	exitChan chan os.Signal // used for trap ctrl+c signal
)

var RootCmd = &cobra.Command{
	Use:   "big-iot-gw",
	Short: "BIG-IoT gateway",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Trap interrupt signal
		exitChan = make(chan os.Signal, 1)
		signal.Notify(exitChan, os.Interrupt)
	},
}

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		log.Log("msg", err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "Config file (default is ./config.yaml)")
	RootCmd.PersistentFlags().String("marketPlaceURI", "https://market.big-iot.org", "Main URI for BIG-IoT Market Place")
	RootCmd.PersistentFlags().String("providerID", "", "Provider ID for BIG-IoT MarketPlace")
	RootCmd.PersistentFlags().String("providerSecret", "", "Provider Secret for BIG-IoT MarketPlace")
	RootCmd.PersistentFlags().Int("offeringActiveLengthSec", 300, "Offering Active Length Sec")
	RootCmd.PersistentFlags().Int("offeringCheckIntervalSec", 10, "Offering Check Interval in secs")
	RootCmd.PersistentFlags().String("offeringEndpoint", "", "Offering End Point")
	RootCmd.PersistentFlags().String("pipeAccessToken", "", "Pipes access token")
	RootCmd.PersistentFlags().Int("HTTPPort", 8080, "HTTP Port where will be running the service")
	RootCmd.PersistentFlags().String("HTTPHost", "localhost", "HTTP Hostname where will be running the service")
	RootCmd.PersistentFlags().Bool("debug", false, "enable debug")
	/*
		bindViper(RootCmd.Flags(),
			"marketPlaceURI",
			"providerID",
			"providerSecret",
			"offeringActiveLengthSec",
			"offeringCheckIntervalSec",
			"offeringEndpoint",
			"pipeAccessToken",
		)
	*/

	viper.BindPFlag("marketPlaceURI", RootCmd.PersistentFlags().Lookup("marketPlaceURI"))
	viper.BindPFlag("providerID", RootCmd.PersistentFlags().Lookup("providerID"))
	viper.BindPFlag("providerSecret", RootCmd.PersistentFlags().Lookup("providerSecret"))
	viper.BindPFlag("offeringActiveLengthSec", RootCmd.PersistentFlags().Lookup("offeringActiveLengthSec"))
	viper.BindPFlag("offeringCheckIntervalSec", RootCmd.PersistentFlags().Lookup("offeringCheckIntervalSec"))
	viper.BindPFlag("offeringEndpoint", RootCmd.PersistentFlags().Lookup("offeringEndpoint"))
	viper.BindPFlag("pipeAccessToken", RootCmd.PersistentFlags().Lookup("pipeAccessToken"))
	viper.BindPFlag("HTTPPort", RootCmd.PersistentFlags().Lookup("HTTPPort"))
	viper.BindPFlag("HTTPHost", RootCmd.PersistentFlags().Lookup("HTTPHost"))
	viper.BindPFlag("debug", RootCmd.PersistentFlags().Lookup("debug"))

}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		//viper.AddConfigPath("/etc")
		viper.SetConfigName("config.yaml")

	}

	viper.AutomaticEnv()

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		log.Log("Using config file:", viper.ConfigFileUsed())
	}
}

func bindViper(flags *pflag.FlagSet, names ...string) {
	for _, name := range names {
		err := viper.BindPFlag(name, flags.Lookup(name))
		if err != nil {
			panic("Error trying to bind:" + name + " " + err.Error())
		}
	}
}
