package api

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"

	db "github.com/PSKP-95/scheduler/db/sqlc"
	"github.com/PSKP-95/scheduler/hooks"
	"github.com/PSKP-95/scheduler/util"
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
		return nil
	}

	err = server.validate.Struct(scheduleReq)

	if err != nil {
		ctx.Status(http.StatusBadRequest).JSON(&fiber.Map{"message": err.Error()})
		return nil
	}

	// validate cron expression
	nextOccurence, err := util.CalculateNextOccurence(scheduleReq.Cron)
	if err != nil {
		ctx.Status(http.StatusBadRequest).JSON(&fiber.Map{"message": err.Error()})
		return nil
	}

	scheduleParams := db.CreateScheduleParams{
		ID:     uuid.New(),
		Cron:   scheduleReq.Cron,
		Hook:   scheduleReq.Hook,
		Owner:  scheduleReq.Owner,
		Active: scheduleReq.Active,
		Till:   scheduleReq.Till,
	}

	schedule, err := server.store.CreateSchedule(ctx.Context(), scheduleParams)

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code.Name() {
			case "foreign_key_violation", "unique_violation":
				ctx.Status(http.StatusBadRequest).JSON(&fiber.Map{"message": err.Error()})
				return nil
			}
		}
		ctx.Status(http.StatusInternalServerError).JSON(&fiber.Map{"message": err.Error()})
		return nil
	}

	occurenceParams := db.CreateOccurenceParams{
		Schedule: schedule.ID,
		Manual:   false,
		Status:   db.StatusPending,
		Occurence: sql.NullTime{
			Time:  nextOccurence,
			Valid: true,
		},
	}

	_, err = server.store.CreateOccurence(ctx.Context(), occurenceParams)

	if err != nil {
		ctx.Status(http.StatusBadRequest).JSON(&fiber.Map{"message": err.Error()})
		return nil
	}

	ctx.Status(http.StatusCreated).JSON(schedule)
	return nil
}

func (server *Server) getSchedule(ctx *fiber.Ctx) error {
	id := ctx.Params("id")
	if id == "" {
		ctx.Status(http.StatusBadRequest).JSON(
			&fiber.Map{"message": "provide parameter id"})
		return nil
	}

	suuid, err := uuid.Parse(id)

	if err != nil {
		ctx.Status(http.StatusBadRequest).JSON(&fiber.Map{"message": fmt.Sprintf("invalid uuid %s. %s", id, err.Error())})
		return nil
	}

	schedule, err := server.store.GetSchedule(ctx.Context(), suuid)

	if err != nil {
		if err == sql.ErrNoRows {
			ctx.Status(http.StatusNotFound).JSON(&fiber.Map{"message": fmt.Sprintf("schedule with id %s not found. %s", id, err.Error())})
			return nil
		}

		ctx.Status(http.StatusInternalServerError).JSON(&fiber.Map{"message": err.Error()})
		return nil
	}

	ctx.Status(http.StatusOK).JSON(schedule)

	return nil
}

func (server *Server) deleteSchedule(ctx *fiber.Ctx) error {
	id := ctx.Params("id")
	if id == "" {
		ctx.Status(http.StatusBadRequest).JSON(
			&fiber.Map{"message": "provide parameter id"})
		return nil
	}

	suuid, err := uuid.Parse(id)

	if err != nil {
		ctx.Status(http.StatusBadRequest).JSON(&fiber.Map{"message": fmt.Sprintf("invalid uuid %s. %s", id, err.Error())})
		return nil
	}

	err = server.store.DeleteSchedule(ctx.Context(), suuid)

	if err != nil {
		ctx.Status(http.StatusInternalServerError).JSON(&fiber.Map{"message": err.Error()})
		return nil
	}

	return nil
}

type updatescheduleRequest struct {
	Cron   string    `json:"cron" validate:"cron"`
	Hook   string    `json:"hook"`
	Active bool      `json:"active"`
	Till   time.Time `json:"till" validate:"datetime"`
}

func (server *Server) editSchedule(ctx *fiber.Ctx) error {
	id := ctx.Params("id")
	if id == "" {
		ctx.Status(http.StatusBadRequest).JSON(
			&fiber.Map{"message": "provide parameter id"})
		return nil
	}

	suuid, err := uuid.Parse(id)

	if err != nil {
		ctx.Status(http.StatusBadRequest).JSON(&fiber.Map{"message": fmt.Sprintf("invalid uuid %s. %s", id, err.Error())})
		return nil
	}

	scheduleReq := updatescheduleRequest{}

	err = ctx.BodyParser(&scheduleReq)
	if err != nil {
		ctx.Status(http.StatusUnprocessableEntity).JSON(
			&fiber.Map{"message": "request failed. " + err.Error()})
		return err
	}

	// err = server.validate.Struct(scheduleReq)

	// if err != nil {
	// 	ctx.Status(http.StatusBadRequest).JSON(&fiber.Map{"message": err.Error()})
	// 	return err
	// }

	scheduleParams := db.UpdateAccountParams{
		ID:     suuid,
		Cron:   scheduleReq.Cron,
		Hook:   scheduleReq.Hook,
		Active: scheduleReq.Active,
		Till:   scheduleReq.Till,
	}

	schedule, err := server.store.UpdateAccount(ctx.Context(), scheduleParams)

	if err != nil {
		if err == sql.ErrNoRows {
			ctx.Status(http.StatusNotFound).JSON(&fiber.Map{"message": fmt.Sprintf("schedule with id %s not found. %s", id, err.Error())})
			return nil
		}

		ctx.Status(http.StatusInternalServerError).JSON(&fiber.Map{"message": err.Error()})
		return nil
	}

	ctx.Status(http.StatusOK).JSON(schedule)

	return nil
}

func (server *Server) triggerSchedule(ctx *fiber.Ctx) error {
	id := ctx.Params("id")
	if id == "" {
		ctx.Status(http.StatusBadRequest).JSON(
			&fiber.Map{"message": "provide parameter id"})
		return nil
	}

	suuid, err := uuid.Parse(id)

	if err != nil {
		ctx.Status(http.StatusBadRequest).JSON(&fiber.Map{"message": fmt.Sprintf("invalid uuid %s. %s", id, err.Error())})
		return nil
	}

	schedule, err := server.store.GetSchedule(ctx.Context(), suuid)

	if err != nil {
		if err == sql.ErrNoRows {
			ctx.Status(http.StatusNotFound).JSON(&fiber.Map{"message": fmt.Sprintf("schedule with id %s not found. %s", id, err.Error())})
			return nil
		}

		ctx.Status(http.StatusInternalServerError).JSON(&fiber.Map{"message": err.Error()})
		return nil
	}

	occurenceParams := db.CreateOccurenceParams{
		Schedule: schedule.ID,
		Manual:   true,
		Status:   db.StatusPending,
		Worker: uuid.NullUUID{
			UUID:  server.worker.GetWorkerId(),
			Valid: true,
		},
	}

	occurence, err := server.store.CreateOccurence(ctx.Context(), occurenceParams)

	if err != nil {
		ctx.Status(http.StatusBadRequest).JSON(&fiber.Map{"message": err.Error()})
		return nil
	}

	message := hooks.Message{
		Type:      hooks.TRIGGER,
		Occurence: occurence,
	}

	server.executor.Submit(message)
	return nil
}

func (server *Server) lisSchedule(ctx *fiber.Ctx) error {
	page := int32(ctx.QueryInt("page", 1))
	size := int32(ctx.QueryInt("size", 10))

	listScheduleParams := db.ListSchedulesParams{
		Owner:  "",
		Limit:  size,
		Offset: size * (page - 1),
	}
	schedules, err := server.store.ListSchedules(ctx.Context(), listScheduleParams)

	if err != nil {
		ctx.Status(http.StatusInternalServerError).JSON(&fiber.Map{"message": err.Error()})
		return nil
	}

	ctx.Status(http.StatusOK).JSON(schedules)

	return nil
}
