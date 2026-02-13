package domain

type Room struct {
	ID        string
	HostID    string
	TrackURL  string
	Timestamp int64
	Status    string // play/pause
	Users     []string
}
