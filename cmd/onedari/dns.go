package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/bakins/onedari/dns"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func runDNS(cmd *cobra.Command, args []string) {
	setLogLevel()

	viper.BindPFlag("api", cmd.PersistentFlags().Lookup("api"))
	viper.BindPFlag("ttl", cmd.PersistentFlags().Lookup("ttl"))
	viper.BindPFlag("domain", cmd.PersistentFlags().Lookup("domain"))

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

func dnsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dns",
		Short: "Run DNS server",
		Run:   runDNS,
	}

	cmd.PersistentFlags().StringP("api", "a", dns.DefaultEndpoint, "API endpoint")
	cmd.PersistentFlags().Uint32P("ttl", "t", dns.DefaultTTL, "DNS ttl")
	cmd.PersistentFlags().StringP("domain", "d", dns.DefaultDomain, "DNS domain")

	return cmd
}
