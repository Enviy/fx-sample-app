package postgres

const (
	CREATED = "CREATED"
	MERGED  = "MERGED"
)

type RecordParam struct {
	FindingID     string   `json:"finding_id"     yaml:"finding_id"`
	DetectionName string   `json:"detection_name" yaml:"detection_name"`
	CollisionSlug string   `json:"collision_slug" yaml:"collision_slug"`
	FirstEvent    int64    `json:"first_event"    yaml:"first_event"`
	LastEvent     int64    `json:"last_event"     yaml:"last_event"`
	RawEvents     []string `json:"raw_events"     yaml:"raw_events"`
}
