package api

import "net"

type (
	// Service is a group of instances.
	Service struct {
		ID        string            `json:"id"`
		Labels    map[string]string `json:"labels"`
		Query     map[string]string `json:"query"`
		Instances []*Instance       `json:"instances,omitempty"`
	}

	// Instance is a single running instance of an app.
	Instance struct {
		ID       string            `json:"id"` // Default is node+app
		Node     string            `json:"node"`
		Labels   map[string]string `json:"labels"`
		Address  net.IP            `json:"ip"`
		Port     uint16            `json:"port"`
		Up       bool              `json:"up"`
		Priority uint16            `json:"priority"`
		Weight   uint16            `json:"weight"`
	}

	// Node is a "server."
	Node struct {
		ID      string `json:"id"`
		Address net.IP `json:"ip"` // base ip usually
	}
)

// NewInstance creates a new, blank instance.
func NewInstance() *Instance {
	return &Instance{
		Labels: make(map[string]string),
	}
}
