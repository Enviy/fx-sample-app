package postgres

import (
	"database/sql"

	"github.com/lib/pq"
)

// Person corresponds to the person table.
type Person struct {
	ID        string `json:"id,omitempty"         db:"id"`
	FirstName string `json:"first_name,omitempty" db:"first_name"`
	LastName  string `json:"last_name,omitempty"  db:"last_name"`
	Email     string `json:"email,omitempty"      db:"email"`
	Timestamp int64  `json:"timestamp,omitempty"  db:"timestamp"`
}

// Place corresponds to the place table.
type Place struct {
	ID        string         `json:"id,omitempty"        db:"id"`
	Country   string         `json:"country,omitempty"   db:"country"`
	City      sql.NullString `json:"city,omitempty"      db:"city"`
	Comments  pq.StringArray `json:"comments,omitempty"  db:"comments"`
	TelCode   int            `json:"telcode,omitempty"   db:"telcode"`
	Timestamp int64          `json:"timestamp,omitempty" db:"timestamp"`
}
