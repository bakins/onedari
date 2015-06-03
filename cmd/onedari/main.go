package main

import (
	log "github.com/Sirupsen/logrus"
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

	root.AddCommand(
		serverCommand(),
		announceCommand(),
		dnsCommand(),
	)
	_ = root.Execute()
}

func setLogLevel() {
	level, err := log.ParseLevel(viper.GetString("log-level"))
	if err != nil {
		log.Fatalf("failed to set log level: %s", err)
	}
	log.SetLevel(level)
}
