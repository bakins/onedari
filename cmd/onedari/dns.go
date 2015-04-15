package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/bakins/onedari/dns"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func runDNS(cmd *cobra.Command, args []string) {
	setLogLevel()

	if len(args) > 0 {
		log.Fatal("extra command line arguments")
	}

	s, err := dns.New(
		dns.Endpoint(viper.GetString("api")),
		dns.TTL(uint32(viper.GetInt("ttl"))),
		dns.Domain(viper.GetString("domain")),
	)

	if err != nil {
		log.Fatal(err)
	}

	if err := s.Run(); err != nil {
		log.Fatal(err)
	}
}
