package api

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
)

func (server *Server) getHooks(ctx *fiber.Ctx) error {
	keys := make([]string, 0, len(server.executor.GetHooks()))
	for k := range server.executor.GetHooks() {
		keys = append(keys, k)
	}

	return ctx.Status(http.StatusOK).JSON(keys)
}
