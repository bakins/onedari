package server

import (
	"encoding/json"
	"net/http"
	"path"
	"time"

	"github.com/bakins/onedari/api"
	"github.com/coreos/etcd/client"
	"golang.org/x/net/context"
)

// LabelSelector returns true if the given query matches the instance.
func LabelSelector(query map[string]string) InstanceSelectorFunc {
	return func(i *api.Instance) bool {
		if query == nil || len(query) == 0 {
			return true
		}

		// just terrible
		labels := i.Labels
		for k, v := range query {
			val, ok := labels[k]
			if !ok || v != val {
				return false
			}
		}
		return true
	}
}

// UpSelector returns true for instances that are "up".
func UpSelector(i *api.Instance) bool {
	return i.Up
}

// NodeSelector returns true only if this instance is on the given node.
// This is a bit of a hack because of the naive search implementation we use.
func NodeSelector(n *api.Node) InstanceSelectorFunc {
	return func(i *api.Instance) bool {
		return i.Node == n.ID
	}
}

// LabelMatches returns true if all the labels in query match label.
func LabelMatches(labels, query map[string]string) bool {
	if query == nil || len(query) == 0 {
		return true
	}

	// just terrible
	for k, v := range query {
		val, ok := labels[k]
		if !ok || v != val {
			return false
		}
	}
	return true
}

func QueryFromRequest(r *http.Request) (map[string]string, error) {
	if err := r.ParseForm(); err != nil {
		return nil, err
	}
	form := r.Form
	if len(form) == 0 {
		return nil, nil
	}
	query := make(map[string]string, len(form))
	for k, v := range form {
		query[k] = v[0]
	}
	return query, nil
}

func isKeyNotFound(err error) bool {
	e, ok := err.(client.Error)
	return ok && e.Code == client.ErrorCodeKeyNotFound
}

func (s *Server) etcdSet(key string, v interface{}) error {
	k := client.NewKeysAPI(s.etcd)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	_, err = k.Set(ctx, path.Join(s.prefix, key), string(data), nil)
	return err
}

func (s *Server) etcdGet(key string, v interface{}) error {
	k := client.NewKeysAPI(s.etcd)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := k.Get(ctx, path.Join(s.prefix, key), nil)
	if err != nil {
		return err
	}

	return json.Unmarshal([]byte(resp.Node.Value), v)
}

func httpError(w http.ResponseWriter, code int, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	e := json.NewEncoder(w)

	// we care if this errors?
	_ = e.Encode(&HTTPError{
		Error:   err.Error(),
		Code:    code,
		Message: http.StatusText(code),
	})
}

func ParseBody(r *http.Request, v interface{}) error {
	return json.NewDecoder(r.Body).Decode(v)
}

func JSON(w http.ResponseWriter, code int, v interface{}) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	return json.NewEncoder(w).Encode(v)
}
