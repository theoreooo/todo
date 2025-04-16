package handlers

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	resp "todo/internal/lib/api/response"
	"todo/internal/lib/logger/sl"
	"todo/internal/models"
	"todo/internal/storage"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
)

// Response представляет ответ обработчика
// @Description Ответ обработчика
type Response struct {
	// Статус ответа (OK/Error)
	Status string `json:"status" example:"OK"`
	// Сообщение об ошибке (если есть)
	Error string `json:"error,omitempty" example:"something went wrong"`
	// Данные ответа
	Data interface{} `json:"data,omitempty"`
}

// internal/http-server/handlers/todo_handler.go
//
//go:generate mockery --name=TaskService --output=mocks --outpkg=mocks
type TaskService interface {
	CreateTask(task models.Task) error
	GetByID(id uint) (*models.Task, error)
	UpdateTask(task *models.Task) error
	DeleteTask(id uint) error
	List(page, limit int, completed *bool, date *time.Time) (*models.TasksList, error)
}

// New godoc
// @Summary Создать новую задачу
// @Description Создать новую задачу с указанными данными
// @Tags tasks
// @Accept json
// @Produce json
// @Param request body models.Task true "Данные задачи"
// @Success 200 {object} handlers.Response
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /newtask [post]
func New(log *slog.Logger, TaskService TaskService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.New"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var req models.Task

		err := render.DecodeJSON(r.Body, &req)
		if err != nil {
			log.Error("failed to decode request body", sl.Err(err))
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, resp.Error("failed to decode request"))
			return
		}

		log.Info("request body decoded", slog.Any("request", req))

		if err := validator.New().Struct(req); err != nil {
			validateErr := err.(validator.ValidationErrors)
			log.Error("invalid request", sl.Err(err))
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, resp.ValidatorError(validateErr))
			return
		}

		err = TaskService.CreateTask(req)
		if err != nil {
			log.Error("failed to add url", sl.Err(err))
			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, resp.Error("failed to create task"))
			return
		}

		log.Info("task created")

		w.WriteHeader(http.StatusOK)
		respObj := Response{
			Status: "OK",
		}
		render.JSON(w, r, respObj)
	}
}

// GetByID godoc
// @Summary Получить задачу по ID
// @Description Получить задачу по её идентификатору
// @Tags tasks
// @Accept json
// @Produce json
// @Param id path int true "ID задачи"
// @Success 200 {object} models.Task
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /tasks/{id} [get]
func GetByID(log *slog.Logger, taskService TaskService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.GetByID"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		idStr := chi.URLParam(r, "id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			log.Error("failed to parse id", sl.Err(err))
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, resp.Error("invalid id"))
			return
		}

		task, err := taskService.GetByID(uint(id))
		if err != nil {
			if errors.Is(err, storage.ErrTaskNotFound) {
				log.Info("task not found", slog.Uint64("id", uint64(id)))
				w.WriteHeader(http.StatusNotFound)
				render.JSON(w, r, resp.Error("task not found"))
				return
			}
			log.Error("failed to get task", sl.Err(err))
			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, resp.Error("failed to get task"))
			return
		}

		w.WriteHeader(http.StatusOK)
		respObj := Response{
			Status: "OK",
			Data:   task,
		}
		render.JSON(w, r, respObj)
	}
}

// UpdateTask godoc
// @Summary Обновить задачу
// @Description Обновить существующую задачу по ID
// @Tags tasks
// @Accept json
// @Produce json
// @Param id path int true "ID задачи"
// @Param request body models.Task true "Данные задачи для обновления"
// @Success 200 {object} handlers.Response
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /tasks/{id} [put]
func UpdateTask(log *slog.Logger, taskService TaskService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.Update"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		idStr := chi.URLParam(r, "id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			log.Error("failed to parse id", sl.Err(err))
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, resp.Error("invalid id"))
			return
		}

		var req models.Task
		err = render.DecodeJSON(r.Body, &req)
		if err != nil {
			log.Error("failed to decode request body", sl.Err(err))
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, resp.Error("failed to decode request"))
			return
		}

		req.ID = id
		log.Info("request body decoded", slog.Any("request", req))

		if err := validator.New().Struct(req); err != nil {
			validateErr := err.(validator.ValidationErrors)
			log.Error("invalid request", sl.Err(err))
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, resp.ValidatorError(validateErr))
			return
		}

		err = taskService.UpdateTask(&req)
		if err != nil {
			if errors.Is(err, storage.ErrTaskNotFound) {
				log.Info("task not found", slog.Int64("id", id))
				w.WriteHeader(http.StatusNotFound)
				render.JSON(w, r, resp.Error("task not found"))
				return
			}
			log.Error("failed to update task", sl.Err(err))
			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, resp.Error("failed to update task"))
			return
		}

		log.Info("task updated", slog.Int64("id", id))
		w.WriteHeader(http.StatusOK)
		respObj := Response{
			Status: "OK",
		}
		render.JSON(w, r, respObj)
	}
}

// DeleteTask godoc
// @Summary Удалить задачу
// @Description Удалить задачу по её идентификатору
// @Tags tasks
// @Accept json
// @Produce json
// @Param id path int true "ID задачи"
// @Success 200 {object} handlers.Response
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /tasks/{id} [delete]
func DeleteTask(log *slog.Logger, taskService TaskService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.Delete"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		idStr := chi.URLParam(r, "id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			log.Error("failed to parse id", sl.Err(err))
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, resp.Error("invalid id"))
			return
		}

		err = taskService.DeleteTask(uint(id))
		if err != nil {
			if errors.Is(err, storage.ErrTaskNotFound) {
				log.Info("task not found", slog.Uint64("id", uint64(id)))
				w.WriteHeader(http.StatusNotFound)
				render.JSON(w, r, resp.Error("task not found"))
				return
			}
			log.Error("failed to delete task", sl.Err(err))
			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, resp.Error("failed to delete task"))
			return
		}

		log.Info("task deleted", slog.Uint64("id", uint64(id)))
		w.WriteHeader(http.StatusOK)
		respObj := Response{
			Status: "OK",
		}
		render.JSON(w, r, respObj)
	}
}

// List godoc
// @Summary Получить список задач
// @Description Получить список задач с пагинацией и фильтрацией
// @Tags tasks
// @Accept json
// @Produce json
// @Param page query int false "Номер страницы" default(1)
// @Param limit query int false "Количество элементов на странице" default(10)
// @Param completed query bool false "Статус задачи (true - выполнена, false - не выполнена)"
// @Param date query string false "Дата в формате YYYY-MM-DD"
// @Success 200 {object} models.TasksList
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /tasks [get]
func List(log *slog.Logger, taskService TaskService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.List"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		pageStr := r.URL.Query().Get("page")
		page, err := strconv.Atoi(pageStr)
		if err != nil || page < 1 {
			page = 1
		}

		limitStr := r.URL.Query().Get("limit")
		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit < 1 {
			limit = 10
		}

		var completed *bool
		completedStr := r.URL.Query().Get("completed")
		if completedStr != "" {
			comp, err := strconv.ParseBool(completedStr)
			if err != nil {
				log.Error("invalid completed parameter", sl.Err(err))
				w.WriteHeader(http.StatusBadRequest)
				render.JSON(w, r, resp.Error("invalid completed parameter"))
				return
			}
			completed = &comp
		}

		var date *time.Time
		dateStr := r.URL.Query().Get("date")
		if dateStr != "" {
			parsedDate, err := time.ParseInLocation("2006-01-02", dateStr, time.Local)
			if err != nil {
				log.Error("invalid date format", sl.Err(err))
				w.WriteHeader(http.StatusBadRequest)
				render.JSON(w, r, resp.Error("invalid date format, use YYYY-MM-DD"))
				return
			}
			date = &parsedDate
		}

		log.Info("parsed query parameters",
			slog.Int("page", page),
			slog.Int("limit", limit),
			slog.Any("completed", completed),
			slog.Any("date", date),
		)

		tasksList, err := taskService.List(page, limit, completed, date)
		if err != nil {
			log.Error("failed to list tasks", sl.Err(err))
			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, resp.Error("failed to list tasks"))
			return
		}

		log.Info("tasks retrieved", slog.Int64("total", tasksList.Total))
		w.WriteHeader(http.StatusOK)
		render.JSON(w, r, tasksList)
	}
}
