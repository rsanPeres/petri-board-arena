package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	_ "github.com/lib/pq"

	"github.com/petri-board-arena/graph"

	createarena "github.com/petri-board-arena/internal/application/command"
	"github.com/petri-board-arena/internal/application/port/repository"

	"github.com/petri-board-arena/internal/infrastructure/adapter"
	pg "github.com/petri-board-arena/internal/infrastructure/persistence/postgres"
	pgwrite "github.com/petri-board-arena/internal/infrastructure/persistence/postgres/write"
)

const defaultPort = "8080"

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL not set")
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("ping db: %v", err)
	}

	uow := pg.NewUnitOfWork(db)
	var writeRepo repository.ArenaWriteRepository = pgwrite.NewArenaRepo(db)

	clock := adapter.RealClock{}
	ids := adapter.UUIDGen{}
	pub := adapter.NopArenaPublisher{}

	createArenaHandler := createarena.NewHandler(uow, writeRepo, ids, clock, pub)

	resolvers := graph.NewResolver(graph.ResolverDeps{
		CreateArenaHandler: createArenaHandler,
	})

	schema := graph.NewExecutableSchema(graph.Config{Resolvers: resolvers})
	srv := handler.NewDefaultServer(schema)

	srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.POST{})

	srv.Use(extension.Introspection{})

	http.Handle("/", playground.Handler("GraphQL playground", "/query"))
	http.Handle("/query", srv)

	log.Printf("connect to http://localhost:%s/ for GraphQL playground", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
