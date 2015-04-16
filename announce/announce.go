package announce

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"runtime"
	"time"

	"github.com/bakins/onedari/api"
)

const (
	// Defaults for Announce.
	DefaultEndpoint = "http://127.0.0.1:63412"
)

type (
	OptionFunc func(*Announce) error

	Announce struct {
		app      string
		endpoint string
		client   *http.Client
	}
)

// Endpoint sets the API endpoint.
func Endpoint(endpoint string) OptionFunc {
	return func(a *Announce) error {
		a.endpoint = endpoint
		return nil
	}
}

// HTTPClient sets the http client to use
func HTTPClient(client *http.Client) OptionFunc {
	return func(a *Announce) error {
		a.client = client
		return nil
	}
}

// New creates a new Announce.
func New(app string, options ...OptionFunc) (*Announce, error) {
	a := &Announce{
		app:      app,
		endpoint: DefaultEndpoint,
		client: &http.Client{
			Timeout: time.Duration(5 * time.Second),
		},
	}

	for _, option := range options {
		if err := option(a); err != nil {
			f := runtime.FuncForPC(reflect.ValueOf(option).Pointer()).Name()
			return nil, fmt.Errorf("failure in function %s: %s", f, err)
		}
	}

	return a, nil
}

// Announce registers a single instance. Set TTL to disable
func (a *Announce) Announce(i *api.Instance, ttl time.Duration) error {
	data, err := json.Marshal(i)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("PUT", a.endpoint+"/v0/node/instances/"+a.app, bytes.NewBuffer(data))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := a.client.Do(req)

	if err != nil {
		return err
	}

	// currently we do not care about body, though it may have a useful error in it
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	return nil
}
