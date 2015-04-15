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

func (s *Server) SaveNode() error {
	return s.etcdSet("nodes/"+s.Node.ID, s.Node)
}

func (s *Server) getLocalNode(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	JSON(w, 200, s.Node)
}

func (s *Server) listNodes(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	k := client.NewKeysAPI(s.etcd)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := k.Get(ctx, path.Join(s.prefix, "nodes"), nil)
	if err != nil {
		httpError(w, http.StatusInternalServerError, err)
		return
	}

	if resp.Node == nil || resp.Node.Nodes == nil {
		httpError(w, http.StatusInternalServerError, EmptyNodeError)
		return
	}
	nodes := make([]*api.Node, 0, len(resp.Node.Nodes))

	for _, n := range resp.Node.Nodes {
		node := &api.Node{}
		err := json.Unmarshal([]byte(n.Value), node)

		if err != nil {
			httpError(w, http.StatusInternalServerError, err)
			return
		}

		_, key := path.Split(n.Key)
		node.ID = key

		nodes = append(nodes, node)
	}

	JSON(w, 200, nodes)

}

func (s *Server) getNode(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id := ps[0].Value

	node := &api.Node{}

	if err := s.etcdGet("/nodes/"+id, node); err != nil {
		code := http.StatusInternalServerError

		if isKeyNotFound(err) {
			code = http.StatusNotFound
		}
		httpError(w, code, err)
		return
	}

	node.ID = id
	_ = JSON(w, http.StatusOK, node)
}
