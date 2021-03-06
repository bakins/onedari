package dns

import (
	"fmt"

	"github.com/bakins/onedari/api"
	d "github.com/miekg/dns"
)

func (s *Server) NodeQuery(name string, w d.ResponseWriter, r *d.Msg) {
	node := &api.Node{}

	if err := s.DoHTTP("/v0/nodes/"+name, node); err != nil {
		// need to check if it is not found
		s.sendError(w, r, err, d.RcodeNameError)
		return
	}
	// sanity check
	if node.Address == nil || node.ID == "" {
		s.sendError(w, r, fmt.Errorf("invalid node: %s, %s", node.Address, node.ID), d.RcodeServerFailure)
		return
	}

	m := &d.Msg{}
	m.SetReply(r)

	question := r.Question[0]
	m.Answer = []d.RR{
		&d.A{
			Hdr: d.RR_Header{
				Name:   question.Name,
				Rrtype: question.Qtype,
				Class:  question.Qclass,
				Ttl:    s.ttl,
			},
			A: node.Address,
		},
	}

	_ = w.WriteMsg(m)
}
