package main

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/thingful/big-iot-gateway/gw"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start BIG-IoT Gateway",
	Run: func(cmd *cobra.Command, args []string) {

		config := gw.NewConfig()
		err := config.Load(viper.AllSettings())
		if err != nil {
			panic(err)
		}

		offerings := gw.OfferConf{}
		if err = offers.Unmarshal(&offerings); err != nil {
			panic(err)
		}

		if err = gw.Start(config, offerings.Offers); err != nil {
			panic(err)
		}
	},
}

func init() {
	RootCmd.AddCommand(startCmd)
}
