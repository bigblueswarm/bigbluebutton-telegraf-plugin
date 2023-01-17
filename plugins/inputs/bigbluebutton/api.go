// Package bigbluebutton provides gather functionality
package bigbluebutton

import (
	"bytes"
	"encoding/xml"
)

// MetadataStruct is a generic object that contains a Metadata and ParsedMetada
type MetadataStruct struct {
	Metadata       Metadata `xml:"metadata"`
	ParsedMetadata map[string]string
}

// ParseMetadata parse the Metadata xml into a map[string]string
func (m *MetadataStruct) ParseMetadata() {
	m.ParsedMetadata = xmlToMap(bytes.NewReader(m.Metadata.Inner))
}

// ContainsMetadata check if the struct contains the metadata
func (m *MetadataStruct) ContainsMetadata(md string) bool {
	_, ok := m.ParsedMetadata[md]
	return ok
}

// GetMetadata returns the metadata value
func (m *MetadataStruct) GetMetadata(md string) string {
	return m.ParsedMetadata[md]
}

// MeetingsResponse is BigBlueButton XML global getMeetings api reponse type
type MeetingsResponse struct {
	XMLName    xml.Name `xml:"response"`
	ReturnCode string   `xml:"returncode"`
	MessageKey string   `xml:"messageKey"`
	Meetings   Meetings `xml:"meetings"`
}

// RecordingsResponse is BigBlueButton XML global getRecordings api response type
type RecordingsResponse struct {
	XMLName    xml.Name   `xml:"response"`
	ReturnCode string     `xml:"returncode"`
	MessageKey string     `xml:"messageKey"`
	Recordings Recordings `xml:"recordings"`
}

// Recordings is BigBlueButton XML recordings section
type Recordings struct {
	XMLName xml.Name    `xml:"recordings"`
	Values  []Recording `xml:"recording"`
}

// Recording is recording response containt information like state, record identifier, ...
type Recording struct {
	XMLName   xml.Name `xml:"recording"`
	RecordID  string   `xml:"recordID"`
	Published bool     `xml:"published"`
	MetadataStruct
}

// Meetings is BigBlueButton XML meetings section
type Meetings struct {
	XMLName xml.Name  `xml:"meetings"`
	Values  []Meeting `xml:"meeting"`
}

type Metadata struct {
	Inner []byte `xml:",innerxml"`
}

// Meeting is a meeting response containing information like name, id, created time, created date, ...
type Meeting struct {
	XMLName               xml.Name `xml:"meeting"`
	ParticipantCount      uint64   `xml:"participantCount"`
	ListenerCount         uint64   `xml:"listenerCount"`
	VoiceParticipantCount uint64   `xml:"voiceParticipantCount"`
	VideoCount            uint64   `xml:"videoCount"`
	Recording             bool     `xml:"recording"`
	MetadataStruct
}

// HealthCheck is a api health check response
type HealthCheck struct {
	XMLName    xml.Name `xml:"response"`
	ReturnCode string   `xml:"returncode"`
	Version    string   `xml:"version"`
}
