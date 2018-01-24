package main

import (
	"os"

	"github.com/thingful/big-iot-gateway/pkg/log"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/thingful/big-iot-gateway/gw"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start BIG-IoT Gateway",
	Run: func(cmd *cobra.Command, args []string) {
		log.Log("msg", "start called")

		// A goroutine listening for a ctrl+c signal
		go func() {
			<-exitChan
			log.Log("msg", "Exiting...")
			// Clean something?
			os.Exit(1)
		}()

		log.Log(viper.AllSettings())

		config := gw.NewConfig()
		err := config.Load(viper.AllSettings())
		if err != nil {
			panic(err)
		}

		if err := gw.Start(config); err != nil {
			panic(err)
		}
	},
}

func init() {
	RootCmd.AddCommand(startCmd)
}
