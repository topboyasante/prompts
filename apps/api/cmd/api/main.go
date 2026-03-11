package main

import (
	"fmt"
	"log"
	"os"

	"github.com/topboyasante/prompts/internal/auth"
	"github.com/topboyasante/prompts/internal/config"
	"github.com/topboyasante/prompts/internal/db"
	"github.com/topboyasante/prompts/internal/identities"
	"github.com/topboyasante/prompts/internal/prompts"
	"github.com/topboyasante/prompts/internal/server"
	"github.com/topboyasante/prompts/internal/storage"
	"github.com/topboyasante/prompts/internal/users"
	"github.com/topboyasante/prompts/internal/versions"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fatalf("load config: %v", err)
	}
	log.Printf("api: config loaded (env=%s, port=%s)", cfg.Env, cfg.Port)

	gdb, err := db.Open(cfg.DatabaseURL)
	if err != nil {
		fatalf("open database connection: %v", err)
	}
	log.Printf("api: database connection established")

	if err := db.RunMigrations(cfg.DatabaseURL, "migrations"); err != nil {
		fatalf("run database migrations: %v", err)
	}
	log.Printf("api: migrations completed")

	usersRepo := users.NewGORMRepository(gdb)
	identitiesRepo := identities.NewGORMRepository(gdb)
	promptsRepo := prompts.NewGORMRepository(gdb)
	versionsRepo := versions.NewGORMRepository(gdb)

	authService := auth.NewAuthService(cfg, usersRepo, identitiesRepo)
	authHandler := auth.NewHandler(authService, cfg)
	authMiddleware := auth.NewMiddleware(cfg.JWTSecret)

	storageClient, err := storage.NewR2Client(cfg)
	if err != nil {
		fatalf("initialize storage client: %v", err)
	}

	promptsHandler := prompts.NewHandler(promptsRepo)
	versionsHandler := versions.NewHandler(versionsRepo, promptsRepo, storageClient)

	r := server.New()
	v1 := r.Group("/v1")
	{
		v1.GET("/auth/:provider/login", authHandler.Login)
		v1.GET("/auth/:provider/callback", authHandler.Callback)

		authorized := v1.Group("/")
		authorized.Use(authMiddleware.Authenticate())
		{
			authorized.POST("/prompts", promptsHandler.Create)
			authorized.POST("/prompts/:id/versions", versionsHandler.Upload)
		}

		v1.GET("/prompts", promptsHandler.Search)
		v1.GET("/prompts/:owner/:name", promptsHandler.Get)
		v1.GET("/prompts/:owner/:name/versions", versionsHandler.List)
		v1.GET("/prompts/:owner/:name/versions/:version/download", versionsHandler.Download)
	}

	if err := r.Run(":" + cfg.Port); err != nil {
		fatalf("start api server: %v", err)
	}
}

func fatalf(format string, args ...any) {
	_, _ = fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
