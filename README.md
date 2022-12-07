# BigBlueButton Input Plugin

[![Codacy Badge](https://app.codacy.com/project/badge/Grade/0ffb957fe6074e93b06b6b52106a4659)](https://www.codacy.com/gh/bigblueswarm/bigbluebutton-telegraf-plugin/dashboard?utm_source=github.com&amp;utm_medium=referral&amp;utm_content=bigblueswarm/bigbluebutton-telegraf-plugin&amp;utm_campaign=Badge_Grade)
[![Codacy Badge](https://app.codacy.com/project/badge/Coverage/0ffb957fe6074e93b06b6b52106a4659)](https://www.codacy.com/gh/bigblueswarm/bigbluebutton-telegraf-plugin/dashboard?utm_source=github.com&utm_medium=referral&utm_content=bigblueswarm/bigbluebutton-telegraf-plugin&utm_campaign=Badge_Coverage)
[![Code linting](https://github.com/bigblueswarm/bbsctl/actions/workflows/lint.yml/badge.svg)](https://github.com/bigblueswarm/bigbluebutton-telegraf-plugin/actions/workflows/lint.yml)
[![Unit tests](https://github.com/bigblueswarm/bbsctl/actions/workflows/unit_test.yml/badge.svg)](https://github.com/bigblueswarm/bigbluebutton-telegraf-plugin/actions/workflows/unit_test.yml)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/bigblueswarm/bigbluebutton-telegraf-plugin)
![GitHub](https://img.shields.io/github/license/bigblueswarm/bigbluebutton-telegraf-plugin)

The BigBlueButton Input Plugin gathers metrics from [BigBlueButton](https://bigbluebutton.org/) server. It uses [BigBlueButton API](https://docs.bigbluebutton.org/dev/api.html) `getMeetings` and `getRecordings` endpoints to query the data.

## Configuration

```toml
[[inputs.bigbluebutton]]
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
```

## Metrics

- bigbluebutton_meetings:
  - fields:
    - active_meetings
    - participant_count
    - listener_count
    - voice_participant_count
    - video_count
    - active_recording
- bigbluebutton_recordings:
  - fields:
    - recordings_count
    - published_recordings_count
- bigbluebutton_api:
  - fields:
	- online

Using the `gather_by_metadata`, plugin will add meetings and recordings metrics grouped by meetings provided metadata like the following:
```
localhost:8090:bigbluebutton_meetings active_recording=0i,listener_count=0i,participant_count=0i,video_count=0i,voice_participant_count=0i,active_meetings=1i 0
```

For example, using the following configuration:
```toml
## Gather metrics by metadata
# Using this option, gathering data will also insert metrics grouped by metadata configuration
gather_by_metadata = ["tenant"]
```
With a meeting:
```xml
<meeting>
	<meetingName>Meeting 2</meetingName>
	<meetingID>b0a78452-2266-4a0a-abae-8a016db8fccd</meetingID>
	<internalMeetingID>6e2f5787a62c9c3e13ee557c847decded4a53d59-1613138647914</internalMeetingID>
	<createTime>1613138647914</createTime>
	<createDate>Fri Feb 12 15:04:07 CET 2021</createDate>
	<voiceBridge>75042</voiceBridge>
	<dialNumber>613-555-1234</dialNumber>
	<attendeePW>e313fc20-2247-48dd-884a-b1cb48c7919c</attendeePW>
	<moderatorPW>be89c431-00d9-4e38-a2f9-c9a54c9873a3</moderatorPW>
	<running>true</running>
	<duration>0</duration>
	<hasUserJoined>true</hasUserJoined>
	<recording>false</recording>
	<hasBeenForciblyEnded>false</hasBeenForciblyEnded>
	<startTime>1613138647937</startTime>
	<endTime>0</endTime>
	<participantCount>5</participantCount>
	<listenerCount>3</listenerCount>
	<voiceParticipantCount>3</voiceParticipantCount>
	<videoCount>1</videoCount>
	<maxUsers>0</maxUsers>
	<moderatorCount>1</moderatorCount>
	<attendees>
	</attendees>
	<metadata>
		<tenant>localhost</tenant>
	</metadata>
	<isBreakout>false</isBreakout>
</meeting>
```
will generate the following metric:
```
localhost:8090:bigbluebutton_meetings active_recording=0i,listener_count=3i,participant_count=5i,video_count=1i,voice_participant_count=3i,active_meetings=1i 1617611008787972024
```

## Example output
```sh
bigbluebutton_meetings active_recording=0i,listener_count=1i,participant_count=2i,video_count=0i,voice_participant_count=0i,active_meetings=2i 1617611008787972024
localhost:8090:bigbluebutton_meetings active_recording=0i,listener_count=0i,participant_count=0i,video_count=0i,voice_participant_count=0i,active_meetings=1i 1617611008787972024
bigbluebutton_recordings recordings_count=0i,published_recordings_count=0i 1617611008800460253
bigbluebutton_api online=1i 1617611008800460842
```

## Installation
- Download the latest release from [release page](https://github.com/SLedunois/bigbluebutton-telegraf-plugin/releases)
- Configure telegraf to call it using execd
 ```toml
[[inputs.execd]]
  command = ["/path/to/bbb-telegraf", "-config", "/path/to/bbb-telegraf/config", "-poll_interval", "10s"]
  signal = "none"
 ```

Alternatively, you can build your own binary using:
```bash
git clone git@github.com:SLedunois/bigbluebutton-telegraf-plugin.git
go build -o bbb-telegraf cmd/main.go
```