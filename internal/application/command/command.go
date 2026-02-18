package command

import "github.com/petri-board-arena/internal/domain/arena"

type Command struct {
	Name   string
	Config arena.Config
	// opcional: quem criou (para auditoria)
	CreatedBy *arena.PlayerID
}

type Result struct {
	ArenaID arena.ID
}
