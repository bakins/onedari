package dns

import (
	"encoding/json"
	"fmt"

	d "github.com/miekg/dns"
)

func (s *Server) sendError(w d.ResponseWriter, req *d.Msg, err error, code int) {
	m := &d.Msg{}
	m.SetRcode(req, code)
	_ = w.WriteMsg(m)

	// TODO: log error? add a logger to options?
}

func (s *Server) DoHTTP(uri string, v interface{}) error {
	resp, err := s.client.Get(s.endpoint + uri)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}
	defer resp.Body.Close()

	return json.NewDecoder(resp.Body).Decode(v)
}
