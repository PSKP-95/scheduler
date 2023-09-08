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
		ctx.Status(http.StatusUnprocessableEntity).JSON(
			&fiber.Map{"message": "request failed. " + err.Error()})
		return err
	}

	err = server.validate.Struct(scheduleReq)
	if err != nil {
		ctx.Status(http.StatusBadRequest).JSON(&fiber.Map{"message": err.Error()})
		return err
	}

	// validate cron expression
	nextOccurence, err := cron.CalculateNextOccurence(scheduleReq.Cron)
	if err != nil {
		ctx.Status(http.StatusBadRequest).JSON(&fiber.Map{"message": err.Error()})
		return err
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
		ctx.Status(http.StatusBadRequest).JSON(&fiber.Map{"message": err.Error()})
		return err
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
	Data   string    `json:"data"`
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
		ctx.Status(http.StatusInternalServerError).JSON(&fiber.Map{"message": err.Error()})
		return nil
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

	ctx.Status(http.StatusOK).JSON(schedulesResp)

	return nil
}
