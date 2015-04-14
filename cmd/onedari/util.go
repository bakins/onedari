package main

import (
	"fmt"
	"net"
	"os"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/bakins/onedari/api"
	"github.com/spf13/viper"
)

func getName() (string, error) {
	name := viper.GetString("name")
	if name == "" {
		var err error
		name, err = os.Hostname()
		if err != nil {
			return "", fmt.Errorf("failed to get hostname: %s", err)
		}
	}

	name = strings.ToLower(name)

	parts := strings.Split(name, ".")
	if len(parts) > 1 {
		log.Warnf("truncating hostname %s to %s", name, parts[0])
		name = parts[0]
	}
	return name, nil

}

func getIP() (net.IP, error) {
	addr := viper.GetString("ip")
	if addr == "" {
		addrs, err := net.InterfaceAddrs()
		if err != nil {
			return nil, fmt.Errorf("failed to get interface addresses: %s", err)
		}

		for _, a := range addrs {
			ip, _, err := net.ParseCIDR(a.String())
			if err != nil {
				// log error?
				continue
			}
			// XXX: ipv4 only now
			if ip.To4() == nil {
				continue
			}
			if ip.IsGlobalUnicast() {
				addr = ip.String()
				break
			}
		}
	}

	if addr == "" {
		return nil, fmt.Errorf("failed to get address")
	}

	ip := net.ParseIP(addr)
	if ip == nil {
		return nil, fmt.Errorf("failed to parse address: %s", addr)
	}

	// XXX: ipv4 only now
	if ip.To4() == nil {
		return nil, fmt.Errorf("not an ipv4 address: %s", ip.String())
	}
	return ip, nil
}

func createNode() (*api.Node, error) {
	n := &api.Node{}
	var err error
	n.ID, err = getName()
	if err != nil {
		return nil, err
	}
	n.Address, err = getIP()
	if err != nil {
		return nil, err
	}
	return n, nil
}
