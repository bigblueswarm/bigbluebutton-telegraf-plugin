// Package bigbluebutton provides gather functionality
package bigbluebutton

// Record is a telegraf acc record object
type Record struct {
	Meetings             uint64
	Participants         uint64
	ListenerParticipants uint64
	VoiceParticipants    uint64
	VideoParticipants    uint64
	ActiveRecordings     uint64
	Recordings           uint64
	PublishedRecordings  uint64
	Online               uint64
}

// NewRecord initialize a new Record struct
func NewRecord() *Record {
	return &Record{
		Meetings:             uint64(0),
		Participants:         uint64(0),
		ListenerParticipants: uint64(0),
		VoiceParticipants:    uint64(0),
		VideoParticipants:    uint64(0),
		ActiveRecordings:     uint64(0),
		Recordings:           uint64(0),
		PublishedRecordings:  uint64(0),
		Online:               uint64(0),
	}
}

// NewRecordFrom initialize a new Record and fill it with computed valued from meetings, recordings and health check
func NewRecordFrom(m []Meeting, r []Recording, h HealthCheck) *Record {
	rec := NewRecord()
	rec.ComputeMeetingMetrics(m)
	rec.ComputeRecordingMetrics(r)
	rec.ComputeOnlineMetric(h)

	return rec
}

// ToMap returns the record as a valid map[string]uint64
func (rec *Record) ToMap() map[string]uint64 {
	return map[string]uint64{
		"meetings":              rec.Meetings,
		"participants":          rec.Participants,
		"listener_participants": rec.ListenerParticipants,
		"voice_participants":    rec.VoiceParticipants,
		"video_participants":    rec.VideoParticipants,
		"active_recordings":     rec.ActiveRecordings,
		"recordings":            rec.Recordings,
		"published_recordings":  rec.PublishedRecordings,
		"online":                rec.Online,
	}
}

// ComputeMeetingMetrics perform a computation and update the record from the meeting values
func (rec *Record) ComputeMeetingMetrics(ms []Meeting) {
	if len(ms) == 0 {
		return
	}

	rec.Meetings = uint64(len(ms))
	for _, m := range ms {
		rec.Participants += m.ParticipantCount
		rec.ListenerParticipants += m.ListenerCount
		rec.VoiceParticipants += m.VoiceParticipantCount
		rec.VideoParticipants += m.VideoCount
		if m.Recording {
			rec.ActiveRecordings++
		}
	}
}

// ComputeRecordingMetrics perform a computation and update the record from the meeting values
func (rec *Record) ComputeRecordingMetrics(rs []Recording) {
	if len(rs) == 0 {
		return
	}

	rec.Recordings = uint64(len(rs))
	for _, r := range rs {
		if r.Published {
			rec.PublishedRecordings++
		}
	}

}

// ComputeOnlineMetric perform a computation and update the record from the meeting values
func (rec *Record) ComputeOnlineMetric(h HealthCheck) {
	if h.ReturnCode == "SUCCESS" {
		rec.Online = 1
	}
}
