package bigbluebutton

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

var emptyState = false

func getXMLResponse(requestURI string) ([]byte, int) {
	apiName := strings.Split(strings.TrimPrefix(requestURI, "/bigbluebutton/api/"), "?")[0]
	if apiName == "/bigbluebutton/api" {
		apiName = "healthcheck"
	}

	xmlFile := fmt.Sprintf("./testdata/%s.xml", apiName)

	if emptyState && apiName != "healthcheck" {
		xmlFile = fmt.Sprintf("%s.empty_state", xmlFile)
	}

	code := 200
	_, err := os.Stat(xmlFile)
	if err != nil {
		return nil, 404
	}

	b, _ := ioutil.ReadFile(xmlFile)
	return b, code
}

// return mocked HTTP server
func getHTTPServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, code := getXMLResponse(r.RequestURI)
		w.WriteHeader(code)
		if code == 200 {
			w.Header().Set("Content-Type", "text/xml")
			w.Write(body)
		} else {
			w.Write([]byte(""))
		}
	}))
}

func getPlugin(url string, gatherByMetatdata []string) BigBlueButton {
	return BigBlueButton{
		URL:              url,
		SecretKey:        "OxShRR1sT8FrJZq",
		GatherByMetadata: gatherByMetatdata,
	}
}

func gather(t *testing.T, url string, gatherByMetatdata []string) *testutil.Accumulator {
	plugin := getPlugin(url, gatherByMetatdata)
	plugin.Init()
	acc := &testutil.Accumulator{}
	plugin.Gather(acc)

	require.Empty(t, acc.Errors)

	return acc
}

func getExpectedEmptyValues() map[string]uint64 {
	record := map[string]uint64{
		"meetings":              0,
		"participants":          0,
		"listener_participants": 0,
		"voice_participants":    0,
		"video_participants":    0,
		"active_recordings":     0,
		"recordings":            0,
		"published_recordings":  0,
		"online":                1,
	}

	return record
}

func getExpectedValues() map[string]uint64 {
	record := map[string]uint64{
		"meetings":              2,
		"participants":          15,
		"listener_participants": 12,
		"voice_participants":    4,
		"video_participants":    1,
		"active_recordings":     1,
		"recordings":            2,
		"published_recordings":  1,
		"online":                1,
	}

	return record
}

func TestBigBlueButton(t *testing.T) {
	emptyState = false
	s := getHTTPServer()
	defer s.Close()

	acc := gather(t, s.URL, []string{})
	record := getExpectedValues()
	tags := make(map[string]string)

	expected := []telegraf.Metric{
		testutil.MustMetric("bigbluebutton", tags, toStringMapInterface(record), time.Unix(0, 0)),
	}

	acc.Wait(len(expected))

	testutil.RequireMetricsEqual(t, expected, acc.GetTelegrafMetrics(), testutil.IgnoreTime())
}

func TestBigBlueButtonEmptyState(t *testing.T) {
	emptyState = true
	s := getHTTPServer()
	defer s.Close()

	acc := gather(t, s.URL, []string{})
	record := getExpectedEmptyValues()
	tags := make(map[string]string)

	expected := []telegraf.Metric{
		testutil.MustMetric("bigbluebutton", tags, toStringMapInterface(record), time.Unix(0, 0)),
	}

	acc.Wait(len(expected))

	testutil.RequireMetricsEqual(t, expected, acc.GetTelegrafMetrics(), testutil.IgnoreTime())
}

func TestBigBlueButtonGatherByMetadata(t *testing.T) {
	emptyState = false
	s := getHTTPServer()
	defer s.Close()

	metadata := "tenant"
	tenant := "localhost"

	acc := gather(t, s.URL, []string{metadata})

	tenantRecord := map[string]uint64{
		"meetings":              1,
		"participants":          5,
		"listener_participants": 3,
		"voice_participants":    3,
		"video_participants":    1,
		"active_recordings":     0,
		"recordings":            1,
		"published_recordings":  1,
		"online":                1,
	}

	record := getExpectedValues()
	tags := map[string]string{
		"tenant": tenant,
	}

	expected := []telegraf.Metric{
		testutil.MustMetric("bigbluebutton", map[string]string{}, toStringMapInterface(record), time.Unix(0, 0)),
		testutil.MustMetric(metadata, tags, toStringMapInterface(tenantRecord), time.Unix(0, 0)),
	}

	acc.Wait(len(expected))
	testutil.RequireMetricsEqual(t, expected, acc.GetTelegrafMetrics(), testutil.IgnoreTime())
}
