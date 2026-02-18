package graph

import (
	createarena "github.com/petri-board-arena/internal/application/command"
)

// ResolverDeps: dependências injetadas (composition root)
type ResolverDeps struct {
	CreateArenaHandler *createarena.Handler
}

// Resolver: raiz do gqlgen
type Resolver struct {
	CreateArenaHandler *createarena.Handler
}

func NewResolver(deps ResolverDeps) *Resolver {
	return &Resolver{
		CreateArenaHandler: deps.CreateArenaHandler,
	}
}
