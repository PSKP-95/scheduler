package api

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
)

func (s *Server) getHooks(ctx *fiber.Ctx) error {
	keys := make([]string, 0, len(s.executor.GetHooks()))
	for k := range s.executor.GetHooks() {
		keys = append(keys, k)
	}

	return ctx.Status(http.StatusOK).JSON(keys)
}
