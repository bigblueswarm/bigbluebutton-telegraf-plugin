// Package bigbluebutton provides gather functionality
package bigbluebutton

import (
	"encoding/xml"
	"io"
)

func xmlToMap(r io.Reader) map[string]string {
	m := make(map[string]string)
	values := make([]string, 0)
	p := xml.NewDecoder(r)
	for token, err := p.Token(); err == nil; token, err = p.Token() {
		switch t := token.(type) {
		case xml.CharData:
			values = append(values, string([]byte(t)))
		case xml.EndElement:
			m[t.Name.Local] = values[len(values)-1]
			values = values[:]
		}
	}

	return m
}
