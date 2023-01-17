// Package bigbluebutton provides gather functionality
package bigbluebutton

import (
	"crypto/sha1"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"

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

// Gather retrieve and publish metrics using the telegraf.Accumulator
func (b *BigBlueButton) Gather(acc telegraf.Accumulator) error {
	m, err := b.getMeetings()
	if err != nil {
		return err
	}

	r, err := b.getRecordings()
	if err != nil {
		return err
	}

	h, err := b.getHealCheck()
	if err != nil {
		return err
	}

	rec := NewRecordFrom(m.Meetings.Values, r.Recordings.Values, *h)
	acc.AddFields("bigbluebutton", toStringMapInterface(rec.ToMap()), make(map[string]string))

	if b.shouldGatherByMetadata() {
		recs := b.GetMetadataRecords(m, r, h)
		for k, v := range recs {
			acc.AddFields(fmt.Sprintf("bigbluebutton:%s", k), toStringMapInterface(v.ToMap()), make(map[string]string))
		}
	}

	return nil
}

// GetMetadataRecords parse responses and returns a map for record
func (b *BigBlueButton) GetMetadataRecords(mr *MeetingsResponse, rr *RecordingsResponse, hr *HealthCheck) map[string]*Record {
	type storage struct {
		meetings   []Meeting
		recordings []Recording
	}

	store := map[string]*storage{}
	res := map[string]*Record{}

	createStorageIfNotExists := func(key string) {
		if _, ok := store[key]; !ok {
			store[key] = &storage{
				meetings:   []Meeting{},
				recordings: []Recording{},
			}
		}
	}

	for _, md := range b.GatherByMetadata {
		for _, m := range mr.Meetings.Values {
			m.ParseMetadata()
			if !m.ContainsMetadata(md) {
				continue
			}

			val := m.GetMetadata(md)
			createStorageIfNotExists(val)

			s := store[val]
			s.meetings = append(s.meetings, m)
		}

		for _, r := range rr.Recordings.Values {
			r.ParseMetadata()
			if !r.ContainsMetadata(md) {
				continue
			}

			val := r.GetMetadata(md)
			createStorageIfNotExists(val)

			s := store[val]
			s.recordings = append(s.recordings, r)
		}

	}

	for key, val := range store {
		res[key] = NewRecordFrom(val.meetings, val.recordings, *hr)
	}

	return res
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

func (b *BigBlueButton) getMeetings() (*MeetingsResponse, error) {
	body, err := b.api(b.getMeetingsURL)
	if err != nil {
		return nil, err
	}

	var response MeetingsResponse
	err = xml.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

func (b *BigBlueButton) getRecordings() (*RecordingsResponse, error) {
	body, err := b.api(b.getRecordingsURL)
	if err != nil {
		return nil, err
	}

	var response RecordingsResponse
	err = xml.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

func (b *BigBlueButton) getHealCheck() (*HealthCheck, error) {
	body, err := b.api(b.getHealthCheckURL())
	if err != nil {
		return nil, err
	}

	var response HealthCheck
	err = xml.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

func (b *BigBlueButton) shouldGatherByMetadata() bool {
	return len(b.GatherByMetadata) > 0
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
