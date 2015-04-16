package dns

import (
	"fmt"
	"strings"

	"github.com/bakins/onedari/api"
	d "github.com/miekg/dns"
)

func (s *Server) ServiceQueryA(name string, w d.ResponseWriter, r *d.Msg) {
	service := &api.Service{}

	if err := s.DoHTTP("/v0/services/"+name, service); err != nil {
		fmt.Println(err)
		s.nameError(w, r, err)
	}

	m := &d.Msg{}
	m.SetReply(r)

	question := r.Question[0]

	header := d.RR_Header{
		Name:   question.Name,
		Rrtype: question.Qtype,
		Class:  question.Qclass,
		Ttl:    s.ttl,
	}

	m.Answer = make([]d.RR, 0, len(service.Instances))

	for _, instance := range service.Instances {
		if instance.Address == nil {
			continue
		}
		answer := &d.A{
			Hdr: header,
			A:   instance.Address,
		}

		m.Answer = append(m.Answer, answer)
	}

	// what if we have no instances?? should we return a dns error

	_ = w.WriteMsg(m)
}

func (s *Server) ServiceQuerySRV(name string, w d.ResponseWriter, r *d.Msg) {
	service := &api.Service{}

	if err := s.DoHTTP("/v0/services/"+name, service); err != nil {
		fmt.Println(err)
		s.nameError(w, r, err)
	}

	m := &d.Msg{}
	m.SetReply(r)

	question := r.Question[0]

	header := d.RR_Header{
		Name:   question.Name,
		Rrtype: question.Qtype,
		Class:  question.Qclass,
		Ttl:    s.ttl,
	}

	m.Answer = make([]d.RR, 0, len(service.Instances))
	m.Extra = make([]d.RR, 0, len(service.Instances))

	for _, instance := range service.Instances {
		if instance.Address == nil || instance.Node == "" {
			continue
		}

		target := strings.ToLower(strings.Join([]string{instance.Node, "nodes", s.domain}, "."))

		answer := &d.SRV{
			Hdr:    header,
			Port:   instance.Port,
			Target: target,
		}

		m.Answer = append(m.Answer, answer)

		extra := &d.A{
			Hdr: d.RR_Header{
				Name:   target,
				Rrtype: d.TypeA,
				Class:  question.Qclass,
				Ttl:    s.ttl,
			},
			A: instance.Address,
		}

		m.Extra = append(m.Extra, extra)

	}

	// what if we have no instances?? should we return a dns error
	_ = w.WriteMsg(m)

}