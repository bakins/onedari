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

// XXX: maybe store instances with node as a directory?
// is common case searching via services or via node?
// also, if we want "External" instances, we

type InstanceSelectorFunc func(*api.Instance) bool

// ListInstances fetches all instances optionally using the query as a selector.
func (s *Server) ListInstances(selectors ...InstanceSelectorFunc) ([]*api.Instance, error) {
	k := client.NewKeysAPI(s.etcd)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := k.Get(ctx, path.Join(s.prefix, "instances"), &client.GetOptions{Recursive: true})
	if err != nil {
		return nil, err
	}

	if resp.Node == nil || resp.Node.Nodes == nil {
		return nil, EmptyNodeError
	}
	instances := make([]*api.Instance, 0, len(resp.Node.Nodes))

NODES:
	for _, n := range resp.Node.Nodes {
		i := &api.Instance{}
		err := json.Unmarshal([]byte(n.Value), i)

		// should a single error be fatal??
		if err != nil {
			return nil, err
		}

		_, i.ID = path.Split(n.Key)

		for _, f := range selectors {
			if !f(i) {
				continue NODES
			}
		}

		instances = append(instances, i)
	}
	return instances, nil
}

func (s *Server) createInstanceNode(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	i := api.NewInstance()
	if err := ParseBody(r, i); err != nil {
		httpError(w, http.StatusBadRequest, err)
		return
	}

	i.Node = s.Node.ID

	app := ps[0].Value
	i.Labels["app"] = app
	i.ID = i.Node + "-" + app

	if i.Address == nil {
		i.Address = s.Node.Address
	}
	if err := s.etcdSet("instances/"+i.ID, i); err != nil {
		httpError(w, http.StatusInternalServerError, err)
		return
	}

	// how to handle error?? logger interface?
	_ = JSON(w, http.StatusCreated, i)
}

func (s *Server) createInstance(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	i := &api.Instance{}
	if err := ParseBody(r, i); err != nil {
		httpError(w, http.StatusBadRequest, err)
		return
	}

	// should labels be required?
	if len(i.Labels) == 0 {
		httpError(w, http.StatusExpectationFailed, MissingLabelError)
		return
	}

	if i.Address == nil && i.Node == "" {
		httpError(w, http.StatusExpectationFailed, InvalidInstanceError)
		return
	}

	// TODO: make sure ID is something valid
	i.ID = ps[0].Value

	if err := s.etcdSet("instances/"+i.ID, i); err != nil {
		httpError(w, http.StatusInternalServerError, err)
		return
	}

	// how to handle error?? logger interface?
	_ = JSON(w, http.StatusCreated, i)
}

func (s *Server) listInstancesNode(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	query, err := QueryFromRequest(r)
	if err != nil {
		httpError(w, http.StatusBadRequest, err)
		return
	}

	instances, err := s.ListInstances(NodeSelector(s.Node), LabelSelector(query))
	if err != nil {
		httpError(w, http.StatusInternalServerError, err)
		return
	}
	_ = JSON(w, http.StatusOK, instances)
}

func (s *Server) listInstances(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	query, err := QueryFromRequest(r)
	if err != nil {
		httpError(w, http.StatusBadRequest, err)
		return
	}

	instances, err := s.ListInstances(LabelSelector(query))

	if err != nil {
		httpError(w, http.StatusInternalServerError, err)
		return
	}
	_ = JSON(w, http.StatusOK, instances)
}

func (s *Server) getInstance(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id := ps[0].Value

	i := &api.Instance{}

	if err := s.etcdGet("instances/"+id, i); err != nil {
		code := http.StatusInternalServerError

		if isKeyNotFound(err) {
			code = http.StatusNotFound
		}
		httpError(w, code, err)
		return
	}

	i.ID = id

	// how to handle error??
	_ = JSON(w, http.StatusOK, i)
}
