package api

import (
	"github.com/PSKP-95/scheduler/config"
	db "github.com/PSKP-95/scheduler/db/sqlc"
	"github.com/PSKP-95/scheduler/hooks"
	"github.com/PSKP-95/scheduler/mlog"
	"github.com/PSKP-95/scheduler/worker"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

type Server struct {
	config   config.ServerConfig
	store    db.Store
	app      *fiber.App
	validate *validator.Validate
	executor *hooks.Executor
	worker   *worker.Worker
	Logger   *mlog.Log
}

func NewServer(config config.ServerConfig, store db.Store, executor *hooks.Executor, worker *worker.Worker, logger *mlog.Log) (*Server, error) {
	server := &Server{
		config:   config,
		store:    store,
		validate: validator.New(),
		executor: executor,
		worker:   worker,
		Logger:   logger,
	}

	server.setupRouter()

	return server, nil
}

func (server *Server) setupRouter() {
	app := fiber.New(fiber.Config{
		ServerHeader: "Fiber",
	})

	// add middlewares
	app.Use(logger.New())
	app.Use(recover.New())

	api := app.Group("/api")

	// add routes to router
	api.Post("/schedule", server.createSchedule)
	api.Get("/schedules", server.listSchedules)
	api.Get("/schedule/:id", server.getSchedule)
	api.Get("/hooks", server.getHooks)
	api.Delete("/schedule/:id", server.deleteSchedule)
	api.Put("/schedule/:id", server.editSchedule)
	api.Get("/schedule/:id/trigger", server.triggerSchedule)
	api.Get("/schedule/:id/history", server.getScheduleHistory)

	app.Static("/", "./ui")

	server.app = app
}

func (server *Server) Start(address string) error {
	return server.app.Listen(address)
}

func (s *Server) Shutdown() error {
	return s.app.Shutdown()
}
