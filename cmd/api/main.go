package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	_ "github.com/lib/pq"

	"github.com/petri-board-arena/graph"

	"github.com/petri-board-arena/internal/runtime/banner"
	"github.com/petri-board-arena/internal/runtime/buildinfo"

	createarena "github.com/petri-board-arena/internal/application/command"
	"github.com/petri-board-arena/internal/application/port/repository"
	"github.com/petri-board-arena/internal/infrastructure/adapter"
	pg "github.com/petri-board-arena/internal/infrastructure/persistence/postgres"
	pgwrite "github.com/petri-board-arena/internal/infrastructure/persistence/postgres/write"
)

func main() {
	banner.Print(banner.Info{
		AppName:       "petri-arena",
		Env:           os.Getenv("APP_ENV"),
		Port:          os.Getenv("PORT"),
		WriteDBURL:    os.Getenv("WRITE_DATABASE_URL"),
		ReadDBURL:     os.Getenv("READ_DATABASE_URL"),
		GitCommit:     buildinfo.GitCommit,
		BuildTime:     buildinfo.BuildTime,
		MigrationsDir: "migrations",
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// ✅ CQRS: write = Postgres
	writeDSN := os.Getenv("WRITE_DATABASE_URL")
	if writeDSN == "" {
		log.Fatal("WRITE_DATABASE_URL not set")
	}

	// ✅ CQRS: read = Redis
	readDSN := os.Getenv("READ_DATABASE_URL")
	if readDSN == "" {
		log.Fatal("READ_DATABASE_URL not set")
	}
	_ = readDSN

	db, err := sql.Open("postgres", writeDSN)
	if err != nil {
		log.Fatalf("open write db: %v", err)
	}
	defer db.Close()

	if err := db.PingContext(context.Background()); err != nil {
		log.Fatalf("ping write db: %v", err)
	}

	// Infra (write side)
	uow := pg.NewUnitOfWork(db)
	var writeRepo repository.ArenaWriteRepository = pgwrite.NewArenaRepo(db)

	clock := adapter.RealClock{}
	ids := adapter.UUIDGen{}
	pub := adapter.NopArenaPublisher{}

	// Application handler (command side)
	createArenaHandler := createarena.NewHandler(uow, writeRepo, ids, clock, pub)

	// GraphQL resolver (composition root)
	resolver := graph.NewResolver(graph.ResolverDeps{
		CreateArenaHandler: createArenaHandler,
	})

	srv := handler.NewDefaultServer(graph.NewExecutableSchema(graph.Config{Resolvers: resolver}))

	http.Handle("/", playground.Handler("GraphQL", "/query"))
	http.Handle("/query", srv)

	log.Printf("GraphQL running at http://localhost:%s/", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
