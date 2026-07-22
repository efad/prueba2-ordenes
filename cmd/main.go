package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/efad/prueba2-ordenes/internal/delivery/graphql"
	"github.com/efad/prueba2-ordenes/internal/delivery/graphql/middleware"
	"github.com/efad/prueba2-ordenes/internal/domain"
	"github.com/efad/prueba2-ordenes/internal/repository/postgres"
	"github.com/efad/prueba2-ordenes/internal/seed"
	jwtservice "github.com/efad/prueba2-ordenes/internal/service/jwt"
	"github.com/efad/prueba2-ordenes/internal/usecase"
	"github.com/vektah/gqlparser/v2/ast"
)

type healthResponse struct {
	Status string `json:"status"`
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	ctx := context.Background()
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL es obligatorio")
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET es obligatorio")
	}

	jwtExpiration := 24 * time.Hour
	if value := os.Getenv("JWT_EXPIRATION"); value != "" {
		parsed, err := time.ParseDuration(value)
		if err != nil {
			log.Fatalf("JWT_EXPIRATION invalido: %v", err)
		}
		jwtExpiration = parsed
	}

	if err := postgres.RunMigrations(databaseURL); err != nil {
		log.Fatalf("error ejecutando migraciones: %v", err)
	}

	db, err := postgres.NewDB(ctx, databaseURL)
	if err != nil {
		log.Fatalf("error conectando postgres: %v", err)
	}
	defer db.Close()

	if err := seed.Products(ctx, db); err != nil {
		log.Fatalf("error ejecutando seed: %v", err)
	}

	userRepo := postgres.NewUserRepository(db)
	productRepo := postgres.NewProductRepository(db)
	orderRepo := postgres.NewOrderRepository(db)
	txManager := postgres.NewTransactionManager(db)
	tokenService, err := jwtservice.NewService(jwtSecret, jwtExpiration)
	if err != nil {
		log.Fatalf("error inicializando jwt: %v", err)
	}

	authUC := usecase.NewAuthUseCase(userRepo, tokenService)
	productUC := usecase.NewProductUseCase(productRepo)
	orderUC := usecase.NewOrderUseCase(orderRepo, productRepo, txManager)
	resolver := &graphql.Resolver{
		AuthUC:    authUC,
		ProductUC: productUC,
		OrderUC:   orderUC,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler)
	mux.Handle("/", playground.Handler("GraphQL playground", "/query"))
	mux.Handle("/query", newGraphQLHandler(resolver, tokenService, userRepo, productRepo))

	addr := ":" + port
	log.Printf("servidor escuchando en http://localhost%s", addr)
	log.Printf("health check: http://localhost%s/health", addr)
	log.Printf("graphql playground: http://localhost%s/", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("error al iniciar servidor: %v", err)
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		http.Error(w, "metodo no permitido", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(healthResponse{Status: "ok"})
}

func newGraphQLHandler(
	resolver *graphql.Resolver,
	tokenService domain.TokenService,
	userRepo domain.UserRepository,
	productRepo domain.ProductRepository,
) http.Handler {
	srv := handler.New(graphql.NewExecutableSchema(graphql.Config{
		Resolvers: resolver,
	}))

	srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.POST{})
	srv.SetQueryCache(lru.New[*ast.QueryDocument](1000))
	srv.Use(extension.Introspection{})
	srv.Use(extension.AutomaticPersistedQuery{
		Cache: lru.New[string](100),
	})
	srv.AroundOperations(middleware.Auth(tokenService))
	srv.AroundOperations(graphql.DataLoaderMiddleware(userRepo, productRepo))
	srv.SetErrorPresenter(graphql.ErrorPresenter)

	return srv
}
