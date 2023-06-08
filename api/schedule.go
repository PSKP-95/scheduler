package api

import (
	"net/http"
	"time"

	db "github.com/PSKP-95/schedular/db/sqlc"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

type createscheduleRequest struct {
	Cron   string    `json:"cron" validate:"required,cron"`
	Hook   string    `json:"hook" validate:"required"`
	Owner  string    `json:"owner"`
	Active bool      `json:"active" validate:"required"`
	Till   time.Time `json:"till" validate:"required"`
}

func (server *Server) createSchedule(ctx *fiber.Ctx) error {
	scheduleReq := createscheduleRequest{}

	err := ctx.BodyParser(&scheduleReq)
	if err != nil {
		ctx.Status(http.StatusUnprocessableEntity).JSON(
			&fiber.Map{"message": "request failed. " + err.Error()})
		return err
	}

	err = server.validate.Struct(scheduleReq)

	if err != nil {
		ctx.Status(http.StatusBadRequest).JSON(&fiber.Map{"message": err.Error()})
		return err
	}

	scheduleParams := db.CreateScheduleParams{
		ID:           uuid.New(),
		Cron:         scheduleReq.Cron,
		Hook:         scheduleReq.Hook,
		Owner:        scheduleReq.Owner,
		Active:       scheduleReq.Active,
		LastModified: time.Now(),
		Till:         scheduleReq.Till,
	}

	schedule, err := server.store.CreateSchedule(ctx.Context(), scheduleParams)

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code.Name() {
			case "foreign_key_violation", "unique_violation":
				ctx.Status(http.StatusBadRequest).JSON(&fiber.Map{"message": err.Error()})
				return err
			}
		}
		ctx.Status(http.StatusInternalServerError).JSON(&fiber.Map{"message": err.Error()})
		return err
	}

	ctx.Status(http.StatusCreated).JSON(schedule)
	return nil
}

func (server *Server) getSchedule(ctx *fiber.Ctx) error {
	return nil
}

func (server *Server) deleteSchedule(ctx *fiber.Ctx) error {
	id := ctx.Params("id")
	if id == "" {
		ctx.Status(http.StatusBadRequest).JSON(
			&fiber.Map{"message": "provide parameter id"})
		return nil
	}

	err := server.store.DeleteSchedule(ctx.Context(), uuid.MustParse(id))

	if err != nil {
		// if pqErr, ok := err.(*pq.Error); ok {
		// 	switch pqErr.Code.Name() {
		// 	case "foreign_key_violation", "unique_violation":
		// 		ctx.Status(http.StatusBadRequest).JSON(&fiber.Map{"message": err.Error()})
		// 		return err
		// 	}
		// }
		ctx.Status(http.StatusInternalServerError).JSON(&fiber.Map{"message": err.Error()})
		return err
	}

	return nil
}
