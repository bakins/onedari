package dns

import (
	"encoding/json"
	"fmt"

	d "github.com/miekg/dns"
)

func (s *Server) nameError(w d.ResponseWriter, req *d.Msg, err error) {
	m := &d.Msg{}
	m.SetReply(req)
	m.SetRcode(req, d.RcodeNameError)
	_ = w.WriteMsg(m)

	// TODO: log error?
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
