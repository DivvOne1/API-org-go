package app

import (
	"net/http"
	"time"

	"org-structure-api/internal/config"
	"org-structure-api/internal/db"
	"org-structure-api/internal/handler"
	"org-structure-api/internal/repository"
	"org-structure-api/internal/service"
)

type App struct {
	server *http.Server
}

func New(cfg config.Config) (*App, error) {
	dbConn, err := db.NewPostgres(cfg.DatabaseURL)
	if err != nil {
		return nil, err
	}

	repo := repository.NewDepartmentRepository(dbConn)
	svc := service.NewDepartmentService(repo)
	httpHandler := handler.NewHandler(svc)

	server := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           httpHandler.Routes(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	return &App{server: server}, nil
}

func (a *App) Run() error {
	return a.server.ListenAndServe()
}
