package api

import (
	"net/http"

	db "github.com/PSKP-95/schedular/db/sqlc"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func (server *Server) getScheduleHistory(ctx *fiber.Ctx) error {
	id := ctx.Params("id")
	if id == "" {
		ctx.Status(http.StatusBadRequest).JSON(
			&fiber.Map{"message": "provide parameter id"})
		return nil
	}

	page := int32(ctx.QueryInt("page", 1))
	size := int32(ctx.QueryInt("size", 10))

	history, err := server.store.ListHistory(ctx.Context(), db.ListHistoryParams{
		Schedule: uuid.MustParse(id),
		Limit:    size,
		Offset:   size * (page - 1),
	})

	if err != nil {
		ctx.Status(http.StatusInternalServerError).JSON(&fiber.Map{"message": err.Error()})
		return err
	}

	ctx.Status(http.StatusOK).JSON(history)

	return nil
}
