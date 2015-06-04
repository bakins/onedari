package main

import (
	"log"
	"strings"

	"github.com/bakins/onedari/server"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func runServer(cmd *cobra.Command, _ []string) {
	setLogLevel()

	viper.BindPFlag("address", cmd.PersistentFlags().Lookup("address"))
	viper.BindPFlag("etcd", cmd.PersistentFlags().Lookup("etcd"))
	viper.BindPFlag("ip", cmd.PersistentFlags().Lookup("ip"))
	viper.BindPFlag("name", cmd.PersistentFlags().Lookup("name"))
	viper.BindPFlag("prefix", cmd.PersistentFlags().Lookup("prefix"))

	endpoints := make([]string, 2)
	for _, e := range strings.Split(viper.GetString("etcd"), ",") {
		endpoints = append(endpoints, strings.TrimSpace(e))
	}

	n, err := createNode()
	if err != nil {
		log.Fatal(err)
	}

	s, err := server.New(
		n,
		server.Address(viper.GetString("address")),
		server.EtcdEndpoints(endpoints),
		server.Prefix(viper.GetString("prefix")),
	)

	if err != nil {
		log.Fatal(err)
	}

	if err = s.Run(); err != nil {
		log.Fatal(err)
	}
}

func serverCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "server",
		Short: "Run http server",
		Run:   runServer,
	}

	cmd.PersistentFlags().StringP("address", "a", server.DefaultAddress, "listen address")
	cmd.PersistentFlags().StringP("etcd", "e", strings.Join(server.DefaultEndpoints, ","),
		"comma seperated list of etcd endpoints")
	cmd.PersistentFlags().StringP("prefix", "p", server.DefaultPrefix, "etcd prefix")
	cmd.PersistentFlags().StringP("name", "n", "", "node name. Default is hostname.")
	cmd.PersistentFlags().StringP("ip", "", "", "node ip. default is detected.")

	return cmd
}
