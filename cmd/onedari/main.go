package main

import (
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/bakins/onedari/announce"
	"github.com/bakins/onedari/dns"
	"github.com/bakins/onedari/server"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func help(cmd *cobra.Command, _ []string) {
	_ = cmd.Help()
}

func main() {
	viper.SetEnvPrefix("onedari")
	viper.AutomaticEnv()

	root := &cobra.Command{
		Use:  "onedari",
		Long: "onedari is a simple service discovery framework for etcd",
		Run:  help,
	}

	root.PersistentFlags().StringP("log-level", "l", "warning", "log level")
	viper.BindPFlag("log-level", root.PersistentFlags().Lookup("log-level"))

	cmdServer := &cobra.Command{
		Use:   "server",
		Short: "Run http server",
		Run:   runServer,
	}

	cmdServer.PersistentFlags().StringP("address", "a", server.DefaultAddress, "listen address")
	viper.BindPFlag("address", cmdServer.PersistentFlags().Lookup("address"))

	cmdServer.PersistentFlags().StringP("etcd", "e", strings.Join(server.DefaultEndpoints, ","), "comma seperated list of etcd endpoints")
	viper.BindPFlag("etcd", cmdServer.PersistentFlags().Lookup("etcd"))

	cmdServer.PersistentFlags().StringP("prefix", "p", server.DefaultPrefix, "etcd prefix")
	viper.BindPFlag("prefix", cmdServer.PersistentFlags().Lookup("prefix"))

	cmdServer.PersistentFlags().StringP("name", "n", "", "node name. Default is hostname.")
	viper.BindPFlag("name", cmdServer.PersistentFlags().Lookup("name"))

	cmdServer.PersistentFlags().StringP("ip", "", "", "node ip. default is detected.")
	viper.BindPFlag("ip", cmdServer.PersistentFlags().Lookup("ip"))

	cmdAnnounce := &cobra.Command{
		Use:   "announce",
		Short: "Run service announcement",
		Run:   runAnnounce,
	}

	cmdAnnounce.PersistentFlags().StringP("api", "a", announce.DefaultEndpoint, "API endpoint")
	viper.BindPFlag("api", cmdAnnounce.PersistentFlags().Lookup("api"))

	cmdAnnounce.PersistentFlags().DurationP("ttl", "t", 0, "ttl")
	viper.BindPFlag("ttl", cmdAnnounce.PersistentFlags().Lookup("ttl"))

	cmdAnnounce.PersistentFlags().DurationP("interval", "i", 60*time.Second, "announce interval")
	viper.BindPFlag("interval", cmdAnnounce.PersistentFlags().Lookup("interval"))

	cmdAnnounce.PersistentFlags().StringP("check", "c", "", "app/service check")
	viper.BindPFlag("check", cmdAnnounce.PersistentFlags().Lookup("check"))

	cmdAnnounce.PersistentFlags().Uint16P("weight", "w", 100, "weight")
	viper.BindPFlag("weight", cmdAnnounce.PersistentFlags().Lookup("weight"))

	cmdAnnounce.PersistentFlags().Uint16P("priority", "p", 100, "priority")
	viper.BindPFlag("priority", cmdAnnounce.PersistentFlags().Lookup("priority"))

	cmdDNS := &cobra.Command{
		Use:   "dns",
		Short: "Run DNS server",
		Run:   runDNS,
	}

	cmdDNS.PersistentFlags().StringP("api", "a", dns.DefaultEndpoint, "API endpoint")
	viper.BindPFlag("api", cmdDNS.PersistentFlags().Lookup("api"))

	cmdDNS.PersistentFlags().Uint32P("ttl", "t", dns.DefaultTTL, "DNS ttl")
	viper.BindPFlag("ttl", cmdDNS.PersistentFlags().Lookup("ttl"))

	cmdDNS.PersistentFlags().StringP("domain", "d", dns.DefaultDomain, "DNS domain")
	viper.BindPFlag("domain", cmdDNS.PersistentFlags().Lookup("domain"))

	root.AddCommand(
		cmdServer,
		cmdAnnounce,
		cmdDNS,
	)
	_ = root.Execute()
}

func runServer(cmd *cobra.Command, _ []string) {
	setLogLevel()

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

func setLogLevel() {
	log.SetFormatter(&log.TextFormatter{DisableColors: true})

	level, err := log.ParseLevel(viper.GetString("log-level"))
	if err != nil {
		log.Fatalf("failed to set log level: %s", err)
	}
	log.SetLevel(level)
}
