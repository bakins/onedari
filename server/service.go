package server

import (
	"encoding/json"
	"net/http"
	"path"
	"time"

	"github.com/bakins/onedari/api"
	"github.com/coreos/etcd/client"
	"github.com/julienschmidt/httprouter"
	"golang.org/x/net/context"
)

func (s *Server) createService(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	v := &api.Service{}
	if err := ParseBody(r, v); err != nil {
		httpError(w, http.StatusBadRequest, err)
		return
	}

	// should labels be required?
	if len(v.Labels) == 0 {
		httpError(w, http.StatusExpectationFailed, MissingLabelError)
		return
	}

	// should query be required?
	if len(v.Query) == 0 {
		httpError(w, http.StatusExpectationFailed, MissingQueryError)
		return
	}

	// TODO: make sure ID is something valid
	v.ID = ps[0].Value
	v.Instances = nil

	if err := s.etcdSet("services/"+v.ID, v); err != nil {
		httpError(w, http.StatusInternalServerError, err)
		return
	}

	// how to handle error?? logger interface?
	_ = JSON(w, http.StatusCreated, s)
}

func (s *Server) listServices(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	query, err := QueryFromRequest(r)
	if err != nil {
		httpError(w, http.StatusBadRequest, err)
		return
	}

	k := client.NewKeysAPI(s.etcd)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := k.Get(ctx, path.Join(s.prefix, "services"), &client.GetOptions{Recursive: true})
	if err != nil {
		httpError(w, http.StatusInternalServerError, err)
		return
	}

	if resp.Node == nil || resp.Node.Nodes == nil {
		httpError(w, http.StatusInternalServerError, EmptyNodeError)
		return
	}
	services := make([]*api.Service, 0, len(resp.Node.Nodes))

	// brute force O(n) search
NODES:
	for _, n := range resp.Node.Nodes {
		v := &api.Service{}
		err := json.Unmarshal([]byte(n.Value), v)

		// should a single error be fatal??
		if err != nil {
			httpError(w, http.StatusInternalServerError, err)
			return
		}

		_, key := path.Split(n.Key)
		v.ID = key

		if !LabelMatches(v.Labels, query) {
			continue NODES
		}
		services = append(services, v)
	}
	_ = JSON(w, http.StatusOK, services)
}

func (s *Server) getService(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id := ps[0].Value

	v := &api.Service{}

	if err := s.etcdGet("services/"+id, v); err != nil {
		code := http.StatusInternalServerError

		if isKeyNotFound(err) {
			code = http.StatusNotFound
		}
		httpError(w, code, err)
		return
	}

	v.ID = id

	var err error
	v.Instances, err = s.ListInstances(LabelSelector(v.Query), UpSelector)
	if err != nil {
		httpError(w, http.StatusInternalServerError, err)
		return
	}

	_ = JSON(w, http.StatusOK, v)
}
