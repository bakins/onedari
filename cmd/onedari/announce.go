package main

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/bakins/onedari/announce"
	"github.com/bakins/onedari/api"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type (
	announcement struct {
		announce *announce.Announce
		instance *api.Instance
		ttl      time.Duration
		check    string
	}
)

func runAnnounce(cmd *cobra.Command, args []string) {
	setLogLevel()

	flags := cmd.PersistentFlags()

	viper.BindPFlag("api", flags.Lookup("api"))
	viper.BindPFlag("check", flags.Lookup("check"))
	viper.BindPFlag("interval", flags.Lookup("interval"))
	viper.BindPFlag("ip", flags.Lookup("ip"))
	viper.BindPFlag("priority", flags.Lookup("priority"))
	viper.BindPFlag("ttl", flags.Lookup("ttl"))
	viper.BindPFlag("weight", flags.Lookup("weight"))
	viper.BindPFlag("port", flags.Lookup("port"))

	if len(args) < 1 {
		log.Fatal("need an app name")
	}

	n, err := createNode()
	if err != nil {
		log.Fatal(err)
	}

	app := args[0]

	ttl := time.Duration(uint32(viper.GetInt("ttl"))) * time.Second
	interval := time.Duration(uint32(viper.GetInt("interval"))) * time.Second

	if ttl != time.Duration(0) && ttl < interval {
		log.Fatal("announce ttl must be greater than interval")
	}

	a, err := announce.New(
		app,
		announce.Endpoint(viper.GetString("api")),
	)
	if err != nil {
		log.Fatal(err)
	}

	i := api.NewInstance()

	i.Port = uint16(viper.GetInt("port"))
	i.Address = n.Address
	i.Metadata["weight"] = fmt.Sprintf("%d", viper.GetInt("weight"))
	i.Metadata["priority"] = fmt.Sprintf("%d", viper.GetInt("priority"))

	_, args = args[0], args[1:]
	for _, arg := range args {
		parts := strings.SplitN(arg, "=", 2)
		if len(parts) != 2 {
			log.Warningf("ignoring invalid label: %s", arg)
			continue
		}
		i.Labels[parts[0]] = parts[1]
	}

	check := viper.GetString("check")
	if check == "" {
		i.Up = true
	}

	v := &announcement{
		announce: a,
		ttl:      ttl,
		check:    check,
		instance: i,
	}

	v.doAnnounce()
	for _ = range time.Tick(interval) {
		v.doAnnounce()
	}
}

func (a *announcement) doAnnounce() {
	if a.check != "" {
		// should we wrap in a timeout?
		c := exec.Command("/bin/sh", "-c", a.check)
		output, err := c.CombinedOutput()
		if err != nil {
			// we should probably do rise/fall style checks
			// only mark as up after x successful and down after y
			log.Printf("check failed '%s' : %s : '%s'", a.check, err, output)
			return
		}
	}

	if err := a.announce.Announce(a.instance, a.ttl); err != nil {
		log.Error(err)
	}
}

func announceCommand() *cobra.Command {

	cmd := &cobra.Command{
		Use:   "announce",
		Short: "Run service announcement",
		Run:   runAnnounce,
	}

	cmd.PersistentFlags().StringP("api", "a", announce.DefaultEndpoint, "API endpoint")
	cmd.PersistentFlags().StringP("check", "c", "", "app/service check")
	cmd.PersistentFlags().StringP("ip", "", "", "node ip. default is detected.")
	cmd.PersistentFlags().Uint16P("priority", "p", 100, "priority")
	cmd.PersistentFlags().Uint16("port", 0, "instance port")
	cmd.PersistentFlags().Uint16P("weight", "w", 100, "weight")
	cmd.PersistentFlags().Uint32P("interval", "i", 60, "announce interval")
	cmd.PersistentFlags().Uint32P("ttl", "t", 0, "ttl")

	return cmd
}
