package server

import (
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"runtime"

	"github.com/bakins/onedari/api"
	"github.com/coreos/etcd/client"
	"github.com/gorilla/handlers"
	"github.com/julienschmidt/httprouter"
)

const (
	DefaultAddress = "127.0.0.1:63412"
	DefaultPrefix  = "/akins.org/onedari"
)

var (
	DefaultEndpoints     = []string{"http://127.0.0.1:2379", "http://127.0.0.1:4001"}
	MissingLabelError    = errors.New("must have at least one label")
	MissingQueryError    = errors.New("must have at least one label query")
	InvalidIDError       = errors.New("invalid ID")
	EmptyNodeError       = errors.New("empty node")
	InvalidInstanceError = errors.New("invalid instance")
)

type (
	Server struct {
		address   string
		endpoints []string
		etcd      client.Client
		prefix    string
		Node      *api.Node
	}

	OptionFunc func(*Server) error

	HTTPError struct {
		Error   string `json:"error"`
		Code    int    `json:"code"`
		Message string `json:"message"`
	}
)

// See http://dave.cheney.net/2014/10/17/functional-options-for-friendly-apis for inspiration for options

func Address(addr string) OptionFunc {
	return func(s *Server) error {
		// TODO: verify that address looks valid?
		s.address = addr
		return nil
	}
}

func EtcdEndpoints(endpoints []string) OptionFunc {
	return func(s *Server) error {
		s.endpoints = endpoints
		return nil
	}
}

func Prefix(p string) OptionFunc {
	return func(s *Server) error {
		s.prefix = p
		return nil
	}
}

func New(node *api.Node, options ...OptionFunc) (*Server, error) {
	s := &Server{
		address:   DefaultAddress,
		endpoints: DefaultEndpoints,
		prefix:    DefaultPrefix,
		Node:      node,
	}

	for _, option := range options {
		if err := option(s); err != nil {
			f := runtime.FuncForPC(reflect.ValueOf(option).Pointer()).Name()
			return nil, fmt.Errorf("failure in function %s: %s", f, err)
		}
	}

	cfg := client.Config{
		Endpoints: s.endpoints,
		Transport: client.DefaultTransport,
	}

	var err error
	s.etcd, err = client.New(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create etcd client: %s", err)
	}

	return s, nil
}

func (s *Server) Run() error {

	if err := s.SaveNode(); err != nil {
		return err
	}

	r := httprouter.New()

	r.PUT("/v0/node/instances/:app", s.createInstanceNode)
	r.GET("/v0/node/instances", s.listInstancesNode)

	r.GET("/v0/node", s.getLocalNode)

	r.GET("/v0/nodes", s.listNodes)
	r.GET("/v0/nodes/:id", s.getNode)

	r.PUT("/v0/instances/:id", s.createInstance)
	r.GET("/v0/instances/:id", s.getInstance)
	r.GET("/v0/instances", s.listInstances)
	// TODO: add patch

	r.PUT("/v0/services/:id", s.createService)
	r.GET("/v0/services/:id", s.getService)
	r.GET("/v0/services", s.listServices)
	// TODO: add patch

	return http.ListenAndServe(s.address, handlers.CompressHandler(r))

}
