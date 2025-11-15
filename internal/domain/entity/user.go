package entity

import "time"

type User struct {
	Id        string    `db:"id"`
	Name      string    `db:"name"`
	IsActive  bool      `db:"is_active"`
	Team      string    `db:"team_id"`
	CreatedAt time.Time `db:"created_at"`
}
