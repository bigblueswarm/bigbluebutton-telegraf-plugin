// Package bigbluebutton provides gather functionality
package bigbluebutton

import "encoding/xml"

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
	XMLName        xml.Name `xml:"recording"`
	RecordID       string   `xml:"recordID"`
	Published      bool     `xml:"published"`
	Metadata       Metadata `xml:"metadata"`
	ParsedMetadata map[string]string
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
	Metadata              Metadata `xml:"metadata"`
	ParsedMetadata        map[string]string
}

// HealthCheck is a api health check response
type HealthCheck struct {
	XMLName    xml.Name `xml:"response"`
	ReturnCode string   `xml:"returncode"`
	Version    string   `xml:"version"`
}

// MeetingsMetric is a meetings metric struct that contains all meetings metric value
type MeetingsMetric struct {
	Meetings             uint64
	Participants         uint64
	ListenerParticipants uint64
	VoiceParticipants    uint64
	VideoParticipants    uint64
	ActiveRecordings     uint64
}

// NewMeetingMetric initialize a new MeetingMetric struct
func NewMeetingMetric() *MeetingsMetric {
	return &MeetingsMetric{
		Meetings:             uint64(0),
		Participants:         uint64(0),
		ListenerParticipants: uint64(0),
		VoiceParticipants:    uint64(0),
		VideoParticipants:    uint64(0),
		ActiveRecordings:     uint64(0),
	}
}

// NewMetadataMeetingMetric initialize a new map containg string as key and RecordingMetric as value
func NewMetadataMeetingMetric() *map[string]MeetingsMetric {
	return &map[string]MeetingsMetric{}
}

// RecordingsMetric is a recordings metric struct that contains all recording metrics value
type RecordingsMetric struct {
	Recordings          uint64
	PublishedRecordings uint64
}

// NewRecordingMetric initialize a new RecordingMetric struct
func NewRecordingMetric() *RecordingsMetric {
	return &RecordingsMetric{
		Recordings:          uint64(0),
		PublishedRecordings: uint64(0),
	}
}

// NewMetadataRecordingMetric initialize a new map containg string as key and RecordingMetric as value
func NewMetadataRecordingMetric() *map[string]RecordingsMetric {
	return &map[string]RecordingsMetric{}
}

// APIStatusMetric is a api status metric struct that contains the api status metric value
type APIStatusMetric struct {
	Online uint64
}

// NewAPIStatusMetric initialize a new APIStatusMetric struct
func NewAPIStatusMetric() *APIStatusMetric {
	return &APIStatusMetric{
		Online: 0,
	}
}
