package dns

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/bakins/onedari/api"
	d "github.com/miekg/dns"
)

const (
	defaultMetadataInt  = 100
	maxServiceResponses = 3 // UDP max number of responses. TODO: listen on tcp as well.
)

func (s *Server) ServiceQueryA(name string, w d.ResponseWriter, r *d.Msg) {
	service := &api.Service{}

	if err := s.DoHTTP("/v0/services/"+name, service); err != nil {
		fmt.Println(err)
		// need to check if it is not found
		s.sendError(w, r, err, d.RcodeNameError)
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

	// we are udp only. need to answer on tcp as well for large responses?
	if len(m.Answer) > maxServiceResponses {
		// should we return a random number
		m.Answer = m.Answer[:maxServiceResponses]
		//m.Truncated = true
	}
	// what if we have no instances?? should we return a dns error

	_ = w.WriteMsg(m)
}

func (s *Server) ServiceQuerySRV(name string, w d.ResponseWriter, r *d.Msg) {
	service := &api.Service{}

	if err := s.DoHTTP("/v0/services/"+name, service); err != nil {
		fmt.Println(err)
		s.sendError(w, r, err, d.RcodeServerFailure)
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
			Hdr:      header,
			Port:     instance.Port,
			Target:   target,
			Weight:   getMetadataInt(instance, "weight"),
			Priority: getMetadataInt(instance, "priority"),
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

func getMetadataInt(instance *api.Instance, f string) uint16 {
	if instance.Metadata == nil {
		return defaultMetadataInt
	}
	v, ok := instance.Metadata[f]
	if !ok {
		return defaultMetadataInt
	}

	// should we care about error here? log it?
	if i, err := strconv.ParseUint(v, 10, 0); err == nil {
		return uint16(i)
	}

	return defaultMetadataInt
}
