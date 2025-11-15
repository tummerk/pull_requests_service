package entity

import "time"

type Team struct {
	Name      string    `db:"name"`
	CreatedAt time.Time `db:"created_at"`
}
