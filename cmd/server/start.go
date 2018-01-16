package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/thingful/big-iot-gateway/gw"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start BIG-IoT Gateway",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("start called")
		fmt.Println("offeringEndpoint:", viper.Get("offeringEndpoint"))
		// A goroutine listening for a ctrl+c signal
		go func() {
			<-exitChan
			fmt.Println("Exiting...")
			// Clean something?
			os.Exit(1)
		}()
		config := gw.NewConfig()
		err := config.Load(viper.AllSettings())
		if err != nil {
			panic(err)
		}

		err = gw.Start(config)
		if err != nil {
			panic(err)
		}
	},
}

func init() {
	RootCmd.AddCommand(startCmd)
}
