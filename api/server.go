package api

import (
	"github.com/PSKP-95/scheduler/config"
	db "github.com/PSKP-95/scheduler/db/sqlc"
	"github.com/PSKP-95/scheduler/hooks"
	"github.com/PSKP-95/scheduler/worker"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/rs/zerolog/log"
)

type Server struct {
	config   config.ServerConfig
	store    db.Store
	app      *fiber.App
	validate *validator.Validate
	executor *hooks.Executor
	worker   *worker.Worker
}

func NewServer(config config.ServerConfig, store db.Store, executor *hooks.Executor, worker *worker.Worker) (*Server, error) {
	server := &Server{
		config:   config,
		store:    store,
		validate: validator.New(),
		executor: executor,
		worker:   worker,
	}

	server.setupRouter()

	return server, nil
}

func (s *Server) setupRouter() {
	app := fiber.New(fiber.Config{
		ServerHeader: "Fiber",
	})

	// add middlewares
	app.Use(logger.New())
	app.Use(recover.New())
	app.Use(compress.New(compress.Config{
		Level: compress.LevelBestSpeed,
	}))

	api := app.Group("/api")

	// add routes to router
	api.Post("/schedule", s.createSchedule)
	api.Get("/schedules", s.listSchedules)
	api.Get("/schedule/:id", s.getSchedule)
	api.Get("/hooks", s.getHooks)
	api.Delete("/schedule/:id", s.deleteSchedule)
	api.Put("/schedule/:id", s.editSchedule)
	api.Get("/schedule/:id/trigger", s.triggerSchedule)
	api.Get("/schedule/:id/history", s.getScheduleHistory)

	app.Static("/", "./ui")

	s.app = app
}

func (s *Server) Start(address string) error {
	return s.app.Listen(address)
}

func (s *Server) Shutdown() error {
	log.Info().Msg("Graceful shutdown of server.")
	return s.app.Shutdown()
}
