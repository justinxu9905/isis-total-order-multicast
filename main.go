package main

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"isis-total-order-multicast/client"
	"isis-total-order-multicast/config"
	log "isis-total-order-multicast/logger"
	"isis-total-order-multicast/server"
)

func main() {
	viper.SetConfigFile("isis.yaml")

	// if a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		log.Printf("Using config file: %v", viper.ConfigFileUsed())
	}

	cmd := &cobra.Command{
		Run: func(c *cobra.Command, args []string) {
			id := viper.GetString("peer_id")

			// node config
			nc := config.NodeConfig{
				Id: id,
				Addr: viper.GetString(id+".ip"),
				Port: viper.GetInt(id+".port"),
				StorageFilename: viper.GetString(id+".filename"),
			}

			// start isis server
			svr := server.NewServer(nc)
			go svr.Start()

			// start isis client
			cli := client.NewClient(svr)
			cli.Start()
		},
	}

	cmd.PersistentFlags().StringP("peer_id", "p", "", "peer id")
	viper.BindPFlag("peer_id", cmd.PersistentFlags().Lookup("peer_id"))

	cmd.Execute()
}
