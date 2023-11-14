package api

import (
	"database/sql"
	"fmt"
	"math"
	"net/http"
	"time"

	"github.com/PSKP-95/scheduler/cron"
	db "github.com/PSKP-95/scheduler/db/sqlc"
	"github.com/PSKP-95/scheduler/hooks"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type createscheduleRequest struct {
	Cron   string    `json:"cron" validate:"required,cron"`
	Hook   string    `json:"hook" validate:"required"`
	Owner  string    `json:"owner"`
	Active bool      `json:"active" validate:"required"`
	Data   string    `json:"data"`
	Till   time.Time `json:"till" validate:"required"`
}

func (server *Server) createSchedule(ctx *fiber.Ctx) error {
	scheduleReq := createscheduleRequest{}

	err := ctx.BodyParser(&scheduleReq)
	if err != nil {
		return ctx.Status(http.StatusUnprocessableEntity).JSON(
			&fiber.Map{"message": "request failed. " + err.Error()})
	}

	err = server.validate.Struct(scheduleReq)
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(&fiber.Map{"message": err.Error()})
	}

	// validate cron expression
	nextOccurence, err := cron.CalculateNextOccurence(scheduleReq.Cron)
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(&fiber.Map{"message": err.Error()})
	}

	scheduleParams := db.CreateScheduleParams{
		ID:     uuid.New(),
		Cron:   scheduleReq.Cron,
		Hook:   scheduleReq.Hook,
		Owner:  scheduleReq.Owner,
		Active: scheduleReq.Active,
		Till:   scheduleReq.Till,
		Data:   scheduleReq.Data,
	}

	occurenceParams := db.CreateOccurenceParams{
		Manual: false,
		Status: db.StatusPending,
		Occurence: sql.NullTime{
			Time:  nextOccurence,
			Valid: true,
		},
	}

	schedule, err := server.store.CreateScheduleAddNextOccurence(ctx.Context(), scheduleParams, occurenceParams)
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(&fiber.Map{"message": err.Error()})
	}

	return ctx.Status(http.StatusCreated).JSON(schedule)
}

func (server *Server) getSchedule(ctx *fiber.Ctx) error {
	id := ctx.Params("id")
	if id == "" {
		return ctx.Status(http.StatusBadRequest).JSON(
			&fiber.Map{"message": "provide parameter id"})
	}

	suuid, err := uuid.Parse(id)
	if err != nil {
		return ctx.
			Status(http.StatusBadRequest).
			JSON(&fiber.Map{"message": fmt.Sprintf("invalid uuid %s. %s", id, err.Error())})
	}

	schedule, err := server.store.GetSchedule(ctx.Context(), suuid)
	if err != nil {
		if err == sql.ErrNoRows {
			return ctx.Status(http.StatusNotFound).JSON(&fiber.Map{"message": fmt.Sprintf("schedule with id %s not found. %s", id, err.Error())})
		}

		return ctx.Status(http.StatusInternalServerError).JSON(&fiber.Map{"message": err.Error()})
	}

	return ctx.Status(http.StatusOK).JSON(schedule)
}

func (server *Server) deleteSchedule(ctx *fiber.Ctx) error {
	id := ctx.Params("id")
	if id == "" {
		return ctx.Status(http.StatusBadRequest).JSON(
			&fiber.Map{"message": "provide parameter id"})
	}

	suuid, err := uuid.Parse(id)
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(&fiber.Map{"message": fmt.Sprintf("invalid uuid %s. %s", id, err.Error())})
	}

	err = server.store.DeleteSchedule(ctx.Context(), suuid)
	if err != nil {
		return ctx.Status(http.StatusInternalServerError).JSON(&fiber.Map{"message": err.Error()})
	}

	return ctx.SendStatus(http.StatusNoContent)
}

type updatescheduleRequest struct {
	Cron   string    `json:"cron" validate:"cron"`
	Hook   string    `json:"hook"`
	Active bool      `json:"active"`
	Till   time.Time `json:"till" validate:"datetime"`
	Data   string    `json:"data"`
}

func (server *Server) editSchedule(ctx *fiber.Ctx) error {
	id := ctx.Params("id")
	if id == "" {
		return ctx.Status(http.StatusBadRequest).JSON(
			&fiber.Map{"message": "provide parameter id"})
	}

	suuid, err := uuid.Parse(id)
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(&fiber.Map{"message": fmt.Sprintf("invalid uuid %s. %s", id, err.Error())})
	}

	scheduleReq := updatescheduleRequest{}

	if err = ctx.BodyParser(&scheduleReq); err != nil {
		return ctx.Status(http.StatusUnprocessableEntity).JSON(
			&fiber.Map{"message": "request failed. " + err.Error()})
	}

	// err = server.validate.Struct(scheduleReq)

	// if err != nil {
	// 	ctx.Status(http.StatusBadRequest).JSON(&fiber.Map{"message": err.Error()})
	// 	return err
	// }

	scheduleParams := db.UpdateScheduleParams{
		ID:     suuid,
		Cron:   scheduleReq.Cron,
		Hook:   scheduleReq.Hook,
		Active: scheduleReq.Active,
		Till:   scheduleReq.Till,
		Data:   scheduleReq.Data,
	}

	schedule, err := server.store.UpdateSchedule(ctx.Context(), scheduleParams)
	if err != nil {
		if err == sql.ErrNoRows {
			return ctx.Status(http.StatusNotFound).JSON(&fiber.Map{"message": fmt.Sprintf("schedule with id %s not found. %s", id, err.Error())})
		}

		return ctx.Status(http.StatusInternalServerError).JSON(&fiber.Map{"message": err.Error()})
	}

	return ctx.Status(http.StatusOK).JSON(schedule)
}

func (server *Server) triggerSchedule(ctx *fiber.Ctx) error {
	id := ctx.Params("id")
	if id == "" {
		return ctx.Status(http.StatusBadRequest).JSON(
			&fiber.Map{"message": "provide parameter id"})
	}

	suuid, err := uuid.Parse(id)
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(&fiber.Map{"message": fmt.Sprintf("invalid uuid %s. %s", id, err.Error())})
	}

	schedule, err := server.store.GetSchedule(ctx.Context(), suuid)
	if err != nil {
		if err == sql.ErrNoRows {
			return ctx.Status(http.StatusNotFound).JSON(&fiber.Map{"message": fmt.Sprintf("schedule with id %s not found. %s", id, err.Error())})
		}

		return ctx.Status(http.StatusInternalServerError).JSON(&fiber.Map{"message": err.Error()})
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
		return ctx.Status(http.StatusBadRequest).JSON(&fiber.Map{"message": err.Error()})
	}

	message := hooks.Message{
		Type:      hooks.TRIGGER,
		Occurence: occurence,
	}

	server.executor.Submit(message)
	return nil
}

type ListSchedulesResponse struct {
	Page      Page                  `json:"page"`
	Schedules []db.ListSchedulesRow `json:"schedules"`
}

func (server *Server) listSchedules(ctx *fiber.Ctx) error {
	page := int32(ctx.QueryInt("page", 1))
	size := int32(ctx.QueryInt("size", 10))

	listScheduleParams := db.ListSchedulesParams{
		Owner:  "",
		Limit:  size,
		Offset: size * (page - 1),
	}

	schedules, err := server.store.ListSchedules(ctx.Context(), listScheduleParams)
	if err != nil {
		return ctx.Status(http.StatusInternalServerError).JSON(&fiber.Map{"message": err.Error()})
	}

	schedulesResp := ListSchedulesResponse{
		Schedules: schedules,
		Page: Page{
			Number:        page,
			Size:          size,
			TotalPages:    int32(math.Ceil(float64(schedules[0].TotalRecords) / float64(size))),
			TotalElements: int32(schedules[0].TotalRecords),
		},
	}

	return ctx.Status(http.StatusOK).JSON(schedulesResp)
}
