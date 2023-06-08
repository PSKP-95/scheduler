package api

import (
	db "github.com/PSKP-95/schedular/db/sqlc"
	"github.com/PSKP-95/schedular/util"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

type Server struct {
	config   util.Config
	store    db.Store
	app      *fiber.App
	validate *validator.Validate
}

func NewServer(config util.Config, store db.Store) (*Server, error) {

	server := &Server{
		config:   config,
		store:    store,
		validate: validator.New(),
	}

	server.setupRouter()

	return server, nil
}

func (server *Server) setupRouter() {
	app := fiber.New()

	// add middlewares
	app.Use(logger.New())
	app.Use(recover.New())

	api := app.Group("/api")

	// add routes to router
	api.Post("/schedule", server.createSchedule)
	api.Get("/schedules", server.getSchedule)
	api.Delete("/schedule/:id", server.deleteSchedule)

	server.app = app
}

func (server *Server) Start(address string) error {
	return server.app.Listen(address)
}

// func errorResponse(err error) gin.H {
// 	return gin.H{"error": err.Error()}
// }
