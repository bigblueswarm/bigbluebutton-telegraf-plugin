# BigBlueButton Input Plugin

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

    ## Server score
	#[inputs.bigbluebutton.scores]
	#  meeting_created = 0
	#  user_joined = 0
	#  user_listen = 0
	#  user_voice_enabled = 0
	#  user_video_enabled = 0
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
    - server score
- bigbluebutton_recordings:
  - fields:
    - recordings_count
    - published_recordings_count

## Example output
``` 
bigbluebutton_meetings active_recording=0i,listener_count=1i,participant_count=2i,video_count=0i,voice_participant_count=0i,score=9i,active_meetings=2i 1617611008787972024
bigbluebutton_recordings recordings_count=0i,published_recordings_count=0i 1617611008800460253
```

## Installation
* Download the latest release from [release page](https://github.com/SLedunois/bigbluebutton-telegraf-plugin/releases)
* Configure telegraf to call it using execd
 ```toml
[[inputs.execd]]
  command = ["/path/to/bbb-telegraf", "-config /path/to/bbb-telegraf/config","-poll_interval 10s"]
  signal = "none"
 ```

Alternatively, you can build your own binary using:
```bash
git clone git@github.com:SLedunois/bigbluebutton-telegraf-plugin.git
go build -o bbb-telegraf cmd/main.go
```