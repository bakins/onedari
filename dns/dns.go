package dns

import (
	"fmt"
	"net/http"
	"reflect"
	"runtime"
	"strings"
	"time"

	d "github.com/miekg/dns"
)

type (
	Server struct {
		address  string
		endpoint string
		client   *http.Client
		domain   string
		ttl      uint32
		server   *d.Server
	}

	OptionFunc func(*Server) error
)

const (
	// DefaultEndpoint is the default URL for onedari API.
	DefaultEndpoint = "http://127.0.0.1:63412"
	// DefaultAddress is the default listen address for the DNS server.
	DefaultAddress = "127.0.0.1:15353"
	// DefaultDomain is the default DNS domain.
	DefaultDomain = "onedari.local."
	// DefaultTTL is the default DNS ttl.
	DefaultTTL = 0
)

const (
	//DNS Query types
	UnknownQueryType = iota
	NodeQueryType
	ServiceQueryType
)

// Endpoint sets the API endpoint.
func Endpoint(endpoint string) OptionFunc {
	return func(s *Server) error {
		s.endpoint = endpoint
		return nil
	}
}

// HTTPClient sets the http client to use.
func HTTPClient(client *http.Client) OptionFunc {
	return func(s *Server) error {
		s.client = client
		return nil
	}
}

// Domain sets the DNS domain.
func Domain(d string) OptionFunc {
	return func(s *Server) error {

		//s.domain = strings.ToLower(strings.Join(d, ".") + ".")
		s.domain = d
		return nil
	}
}

// TTL sets the DNS ttl in seconds
func TTL(ttl uint32) OptionFunc {
	return func(s *Server) error {
		s.ttl = ttl
		return nil
	}
}

// New creates a new DNS Server.
func New(options ...OptionFunc) (*Server, error) {
	s := &Server{
		address:  DefaultAddress,
		endpoint: DefaultEndpoint,
		domain:   DefaultDomain,
		ttl:      DefaultTTL,
		client: &http.Client{
			Timeout: time.Duration(5 * time.Second),
		},
	}

	for _, option := range options {
		if err := option(s); err != nil {
			f := runtime.FuncForPC(reflect.ValueOf(option).Pointer()).Name()
			return nil, fmt.Errorf("failure in function %s: %s", f, err)
		}
	}

	return s, nil
}

// Run starts the server.  It does not return, generally.
func (s *Server) Run() error {
	d.Handle(s.domain, s)

	s.server = &d.Server{
		Addr:         s.address,
		Net:          "udp",
		ReadTimeout:  10 * time.Second, // configurable??
		WriteTimeout: 10 * time.Second, // configurable??
	}

	return s.server.ListenAndServe()
}

// getQueryType gets query type.
func getQueryType(name string) (int, string, error) {
	parts := strings.Split(name, ".")
	// pop blank field
	parts = parts[:len(parts)-1]
	if len(parts) != 2 {
		return UnknownQueryType, "", fmt.Errorf("incorrect length of name: %s", name)
	}
	switch parts[1] {
	case "services":
		return ServiceQueryType, parts[0], nil
	case "nodes":
		return NodeQueryType, parts[0], nil
	default:
		return UnknownQueryType, "", fmt.Errorf("unknown sub-domain: %s", parts[1])
	}
}

// ServeDNS implements the dns.Server interface.
func (s *Server) ServeDNS(w d.ResponseWriter, r *d.Msg) {
	// get just the query in lowercase
	query := strings.TrimSuffix(strings.ToLower(r.Question[0].Name), s.domain)

	queryType, name, err := getQueryType(query)

	if err != nil {
		s.sendError(w, r, err, d.RcodeNameError)
		return
	}

	qType := r.Question[0].Qtype

	// this is a bit clumsy
	switch qType {
	case d.TypeA:
		switch queryType {
		case ServiceQueryType:
			s.ServiceQueryA(name, w, r)
			return
		case NodeQueryType:
			s.NodeQuery(name, w, r)
			return
		}
	case d.TypeSRV:
		switch queryType {
		case ServiceQueryType:
			s.ServiceQuerySRV(name, w, r)
			return
		default:
			s.sendError(w, r, fmt.Errorf("invalid query type for SRV: %s", query), d.RcodeNameError)
			return
		}
	default:
		// unknown query type
		s.sendError(w, r, fmt.Errorf("unhandled query type: %s", qType), d.RcodeNameError)
		return
	}

	// we shouldn't make it to here
	s.sendError(w, r, fmt.Errorf("unhandled query: %s", query), d.RcodeNameError)
}
