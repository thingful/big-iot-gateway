package main

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/thingful/big-iot-gateway/gw"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start BIG-IoT Gateway",
	RunE: func(cmd *cobra.Command, args []string) error {

		config := gw.NewConfig()
		err := config.Load(viper.AllSettings())
		if err != nil {
			return err
		}

		offerings := gw.OfferConf{}
		if err = offers.Unmarshal(&offerings); err != nil {
			return err
		}

		if err = gw.Start(config, offerings.Offers); err != nil {
			return err
		}
		return nil
	},
}

func init() {
	RootCmd.AddCommand(startCmd)
}
