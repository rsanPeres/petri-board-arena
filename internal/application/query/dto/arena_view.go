package dto

import "time"

type ArenaView struct {
	ID        string
	Name      string
	Status    string
	Tick      int64
	CreatedAt time.Time
	StartedAt *time.Time
}
