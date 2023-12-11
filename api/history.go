package api

import (
	"math"
	"net/http"

	db "github.com/PSKP-95/scheduler/db/sqlc"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type ScheduleHistoryResponse struct {
	Page    Page                `json:"page"`
	History []db.ListHistoryRow `json:"history"`
}

func (s *Server) getScheduleHistory(ctx *fiber.Ctx) error {
	id := ctx.Params("id")
	if id == "" {
		return ctx.Status(http.StatusBadRequest).JSON(
			&fiber.Map{"message": "provide parameter id"})
	}

	page := int32(ctx.QueryInt("page", 1))
	size := int32(ctx.QueryInt("size", 10))

	history, err := s.store.ListHistory(ctx.Context(), db.ListHistoryParams{
		Schedule: uuid.MustParse(id),
		Limit:    size,
		Offset:   size * (page - 1),
	})

	if err != nil {
		return ctx.Status(http.StatusInternalServerError).JSON(&fiber.Map{"message": err.Error()})
	}

	scheduleHistoryResponse := ScheduleHistoryResponse{
		History: history,
		Page: Page{
			Number:        page,
			Size:          size,
			TotalPages:    int32(math.Ceil(float64(history[0].TotalRecords) / float64(size))),
			TotalElements: int32(history[0].TotalRecords),
		},
	}

	return ctx.Status(http.StatusOK).JSON(scheduleHistoryResponse)
}
