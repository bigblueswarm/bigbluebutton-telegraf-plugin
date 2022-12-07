// Package bigbluebutton provides gather functionality
package bigbluebutton

import (
	"bytes"
	"crypto/sha1"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/common/proxy"
	"github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
)

// BigBlueButton is the global configuration object
type BigBlueButton struct {
	URL              string   `toml:"url"`
	PathPrefix       string   `toml:"path_prefix"`
	SecretKey        string   `toml:"secret_key"`
	Username         string   `toml:"username"`
	Password         string   `toml:"password"`
	GatherByMetadata []string `toml:"gather_by_metadata"`
	getMeetingsURL   string
	getRecordingsURL string
	healthCheckURL   string

	tls.ClientConfig
	proxy.HTTPProxy
	client *http.Client
}

var defaultPathPrefix = "/bigbluebutton"

var sampleConfig = `
	## Required BigBlueButton server url
	url = "http://localhost:8090"

	## BigBlueButton path prefix. Default is "/bigbluebutton"
	# path_prefix = "/bigbluebutton"

	## Required BigBlueButton secret key
	secret_key = ""

	## Gather metrics by metadata
	# Using this option, gathering data will also insert metrics grouped by metadata configuration
	# gather_by_metadata = []

	## Optional HTTP Basic Auth Credentials
	# username = "username"
	# password = "pa$$word

	## Optional HTTP Proxy support
	# http_proxy_url = ""

	## Optional TLS Config
	# tls_ca = "/etc/telegraf/ca.pem"
	# tls_cert = "/etc/telegraf/cert.pem"
	# tls_key = "/etc/telegraf/key.pem"

	## Use TLS but skip chain & host verification
	# insecure_skip_verify = false
`

// Init initialize the BigBlueButton struct with precalculated data
func (b *BigBlueButton) Init() error {
	if b.SecretKey == "" {
		return fmt.Errorf("BigBlueButton secret key is required")
	}

	if b.PathPrefix == "" {
		b.PathPrefix = defaultPathPrefix
	}

	b.getMeetingsURL = b.getURL("getMeetings")
	b.getRecordingsURL = b.getURL("getRecordings")
	b.healthCheckURL = b.getHealthCheckURL()

	tlsCfg, err := b.ClientConfig.TLSConfig()
	if err != nil {
		return err
	}

	proxy, err := b.HTTPProxy.Proxy()
	if err != nil {
		return err
	}

	transport := &http.Transport{
		TLSClientConfig: tlsCfg,
		Proxy:           proxy,
	}

	b.client = &http.Client{
		Transport: transport,
	}

	return nil
}

// SampleConfig provides a sample config object
func (b *BigBlueButton) SampleConfig() string {
	return sampleConfig
}

// Description provides a simple description sentence that explain the plugin
func (b *BigBlueButton) Description() string {
	return "Gather BigBlueButton web conferencing server metrics"
}

// Gather gather data from the BigBlueButton server end send them into the telegraf accumulator
func (b *BigBlueButton) Gather(acc telegraf.Accumulator) error {
	if err := b.gatherMeetings(acc); err != nil {
		return err
	}

	if err := b.gatherRecordings(acc); err != nil {
		return err
	}

	return b.gatherAPIStatus(acc)
}

// BigBlueButton uses an authentication based on a SHA1 checksum processed from api call name and server secret key
func (b *BigBlueButton) checksum(apiCallName string) []byte {
	hash := sha1.New()
	hash.Write([]byte(fmt.Sprintf("%s%s", apiCallName, b.SecretKey)))
	return hash.Sum(nil)
}

func (b *BigBlueButton) getURL(apiCallName string) string {
	endpoint := fmt.Sprintf("%s/api/%s", b.PathPrefix, apiCallName)
	return fmt.Sprintf("%s%s?checksum=%x", b.URL, endpoint, b.checksum(apiCallName))
}

func (b *BigBlueButton) getHealthCheckURL() string {
	endpoint := fmt.Sprintf("%s/api", b.PathPrefix)
	return fmt.Sprintf("%s%s", b.URL, endpoint)
}

// Call BBB server api
func (b *BigBlueButton) api(url string) ([]byte, error) {
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	if b.Username != "" || b.Password != "" {
		request.SetBasicAuth(b.Username, b.Password)
	}

	resp, err := b.client.Do(request)

	if err != nil || resp.StatusCode != 200 {
		return nil, fmt.Errorf("error getting bbb metrics: %s status %d", err, resp.StatusCode)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return nil, err
	}

	return body, nil
}

func (b *BigBlueButton) gatherAPIStatus(acc telegraf.Accumulator) error {
	record := map[string]uint64{
		"online": 0,
	}

	body, err := b.api(b.healthCheckURL)
	if err != nil {
		acc.AddFields("bigbluebutton_api", toStringMapInterface(record), make(map[string]string))
		return err
	}

	var response HealthCheck
	err = xml.Unmarshal(body, &response)
	if err != nil {
		acc.AddFields("bigbluebutton_api", toStringMapInterface(record), make(map[string]string))
		return err
	}

	if response.ReturnCode == "SUCCESS" {
		record["online"] = 1
	}

	acc.AddFields("bigbluebutton_api", toStringMapInterface(record), make(map[string]string))
	return nil
}

func (b *BigBlueButton) gatherMeetings(acc telegraf.Accumulator) error {
	body, err := b.api(b.getMeetingsURL)
	if err != nil {
		return err
	}

	var response MeetingsResponse
	err = xml.Unmarshal(body, &response)
	if err != nil {
		return err
	}

	record := emptyMeetingsMap()
	mRecords := map[string]map[string]uint64{}

	if response.MessageKey == "noMeetings" {
		acc.AddFields("bigbluebutton_meetings", toStringMapInterface(record), make(map[string]string))
		return nil
	}

	for i := 0; i < len(response.Meetings.Values); i++ {
		meeting := response.Meetings.Values[i]
		meeting.ParsedMetadata = xmlToMap(bytes.NewReader(meeting.Metadata.Inner))
		record["active_meetings"]++
		record["participant_count"] += meeting.ParticipantCount
		record["listener_count"] += meeting.ListenerCount
		record["voice_participant_count"] += meeting.VoiceParticipantCount
		record["video_count"] += meeting.VideoCount
		if meeting.Recording {
			record["active_recording"]++
		}

		if b.shouldGatherByMetadata() {
			b.gatherMeetingsByMetadata(&mRecords, meeting)
		}
	}

	acc.AddFields("bigbluebutton_meetings", toStringMapInterface(record), make(map[string]string))
	addMetadataRecordingsToAcc(acc, mRecords)
	return nil
}

func (b *BigBlueButton) gatherRecordingsByMetadata(records *map[string]map[string]uint64, recording Recording) {
	rec := (*records)
	for i := 0; i < len(b.GatherByMetadata); i++ {
		name := b.GatherByMetadata[i]

		if val, ok := recording.ParsedMetadata[name]; ok { // Check if metadata name found in parsed metadata
			key := fmt.Sprintf("%s:bigbluebutton_recordings", val)
			if _, ok := rec[key]; !ok { // If val not found in storage then initialize storage
				rec[key] = emptyRecordingsMap()
			}

			rec[key]["recordings_count"]++
			if recording.Published {
				rec[key]["published_recordings_count"]++
			}
		}
	}
}

func (b *BigBlueButton) gatherMeetingsByMetadata(records *map[string]map[string]uint64, meeting Meeting) {
	rec := (*records)
	for i := 0; i < len(b.GatherByMetadata); i++ {
		name := b.GatherByMetadata[i]

		if val, ok := meeting.ParsedMetadata[name]; ok { // Check if metadata name found in parsed metadata
			key := fmt.Sprintf("%s:bigbluebutton_meetings", val)
			if _, ok := rec[key]; !ok { // If val not found in storage then initialize storage
				rec[key] = emptyMeetingsMap()
			}

			rec[key]["active_meetings"]++
			rec[key]["participant_count"] += meeting.ParticipantCount
			rec[key]["listener_count"] += meeting.ListenerCount
			rec[key]["voice_participant_count"] += meeting.VoiceParticipantCount
			rec[key]["video_count"] += meeting.VideoCount
			if meeting.Recording {
				rec[key]["active_recording"]++
			}
		}
	}
}

func (b *BigBlueButton) shouldGatherByMetadata() bool {
	return len(b.GatherByMetadata) > 0
}

func addMetadataRecordingsToAcc(acc telegraf.Accumulator, records map[string]map[string]uint64) {
	for key, val := range records {
		acc.AddFields(key, toStringMapInterface(val), make(map[string]string), time.Unix(0, 0))
	}
}

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

func emptyMeetingsMap() map[string]uint64 {
	return map[string]uint64{
		"active_meetings":         0,
		"active_recording":        0,
		"listener_count":          0,
		"participant_count":       0,
		"video_count":             0,
		"voice_participant_count": 0,
	}
}

func emptyRecordingsMap() map[string]uint64 {
	return map[string]uint64{
		"recordings_count":           0,
		"published_recordings_count": 0,
	}
}

func (b *BigBlueButton) gatherRecordings(acc telegraf.Accumulator) error {
	body, err := b.api(b.getRecordingsURL)
	if err != nil {
		return err
	}

	var response RecordingsResponse
	err = xml.Unmarshal(body, &response)
	if err != nil {
		return err
	}

	record := emptyRecordingsMap()
	mRecords := map[string]map[string]uint64{}

	if response.MessageKey == "noRecordings" {
		acc.AddFields("bigbluebutton_recordings", toStringMapInterface(record), make(map[string]string))
		return nil
	}

	for i := 0; i < len(response.Recordings.Values); i++ {
		recording := response.Recordings.Values[i]
		recording.ParsedMetadata = xmlToMap(bytes.NewReader(recording.Metadata.Inner))
		record["recordings_count"]++
		if recording.Published {
			record["published_recordings_count"]++
		}

		if b.shouldGatherByMetadata() {
			b.gatherRecordingsByMetadata(&mRecords, recording)
		}
	}

	acc.AddFields("bigbluebutton_recordings", toStringMapInterface(record), make(map[string]string))
	addMetadataRecordingsToAcc(acc, mRecords)
	return nil
}

func toStringMapInterface(in map[string]uint64) map[string]interface{} {
	m := make(map[string]interface{}, len(in))
	for k, v := range in {
		m[k] = v
	}
	return m
}

func init() {
	inputs.Add("bigbluebutton", func() telegraf.Input {
		return &BigBlueButton{}
	})
}
