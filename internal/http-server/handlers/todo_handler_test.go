package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/go-chi/chi"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"todo/internal/http-server/handlers"
	"todo/internal/http-server/handlers/mocks"
	"todo/internal/models"
	"todo/internal/storage"
)

func TestCreateHandler(t *testing.T) {
	cases := []struct {
		name        string
		title       string
		description string
		due_date    time.Time
		status      bool
		respError   string
		mockError   error
		expectCode  int
	}{
		{
			name:        "Success status false",
			title:       "test_title",
			description: "test_decription",
			due_date:    time.Now(),
			status:      false,
			expectCode:  http.StatusOK,
		},
		{
			name:        "Success status true",
			title:       "test_title",
			description: "test_decription",
			due_date:    time.Now(),
			status:      true,
			expectCode:  http.StatusOK,
		},
		{
			name:        "Empty title",
			title:       "",
			description: "test_decription",
			due_date:    time.Now(),
			status:      false,
			respError:   "field title is a required field",
			expectCode:  http.StatusBadRequest,
		},
		{
			name:        "Empty description",
			title:       "test_title",
			description: "",
			due_date:    time.Now(),
			status:      false,
			expectCode:  http.StatusOK,
		},
		{
			name:        "Empty due_date",
			title:       "test_title",
			description: "test_decription",
			status:      false,
			respError:   "field due_date is a required field",
			expectCode:  http.StatusBadRequest,
		},
		{
			name:        "Empty status",
			title:       "test_title",
			description: "test_decription",
			due_date:    time.Now(),
			expectCode:  http.StatusOK,
		},
		{
			name:        "Database error",
			title:       "test_title",
			description: "test_decription",
			due_date:    time.Now(),
			status:      false,
			mockError:   errors.New("db failed"),
			respError:   "failed to create task",
			expectCode:  http.StatusInternalServerError,
		},
	}

	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			taskCreaterMock := mocks.NewTaskService(t)

			if tc.respError == "" || tc.mockError != nil {
				taskCreaterMock.On("CreateTask", mock.AnythingOfType("models.Task")).
					Return(tc.mockError).
					Once()
			}

			logger := slog.New(slog.NewTextHandler(io.Discard, nil))
			handler := handlers.New(logger, taskCreaterMock)

			input := struct {
				Title       string    `json:"title"`
				Description string    `json:"description"`
				DueDate     time.Time `json:"due_date"`
				Status      bool      `json:"status"`
			}{
				Title:       tc.title,
				Description: tc.description,
				DueDate:     tc.due_date,
				Status:      tc.status,
			}

			jsonInput, err := json.Marshal(input)
			require.NoError(t, err)

			req, err := http.NewRequest(http.MethodPost, "/newtask", bytes.NewReader(jsonInput))
			require.NoError(t, err)

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			require.Equal(t, tc.expectCode, rr.Code)

			var resp handlers.Response
			err = json.Unmarshal(rr.Body.Bytes(), &resp)
			require.NoError(t, err)

			require.Equal(t, tc.respError, resp.Error)

			taskCreaterMock.AssertExpectations(t)
		})
	}
}

func TestGetByIDHandler(t *testing.T) {
	now := time.Now()

	task := &models.Task{
		ID:          1,
		Title:       "test_title",
		Description: "test_description",
		DueDate:     now,
		Status:      false,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	cases := []struct {
		name         string
		id           string
		mockResp     *models.Task
		mockError    error
		expectedCode int
		expectedErr  string
	}{
		{
			name:         "Success",
			id:           "1",
			mockResp:     task,
			expectedCode: http.StatusOK,
		},
		{
			name:         "Not Found",
			id:           "2",
			mockError:    storage.ErrTaskNotFound,
			expectedCode: http.StatusNotFound,
			expectedErr:  "task not found",
		},
		{
			name:         "Invalid ID format",
			id:           "abc",
			expectedCode: http.StatusBadRequest,
			expectedErr:  "invalid id",
		},
		{
			name:         "Internal error",
			id:           "3",
			mockError:    errors.New("unexpected error"),
			expectedCode: http.StatusInternalServerError,
			expectedErr:  "failed to get task",
		},
	}

	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			taskGetterMock := mocks.NewTaskService(t)

			if tc.mockResp != nil || tc.mockError != nil {
				id, err := strconv.Atoi(tc.id)
				if err == nil {
					taskGetterMock.On("GetByID", uint(id)).
						Return(tc.mockResp, tc.mockError).
						Once()
				}
			}

			logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
			handler := handlers.GetByID(logger, taskGetterMock)

			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/tasks/%s", tc.id), nil)
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tc.id)

			ctx := context.WithValue(req.Context(), chi.RouteCtxKey, rctx)
			req = req.WithContext(ctx)

			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			require.Equal(t, tc.expectedCode, rr.Code)

			if tc.expectedErr != "" {
				require.Contains(t, rr.Body.String(), tc.expectedErr)
			}

			taskGetterMock.AssertExpectations(t)
		})
	}
}

func TestUpdateTaskHandler(t *testing.T) {
	now := time.Now()

	task := &models.Task{
		ID:          1,
		Title:       "updated_title",
		Description: "updated_description",
		DueDate:     now,
		Status:      true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	cases := []struct {
		name       string
		id         string
		task       *models.Task
		mockError  error
		respError  string
		expectCode int
	}{
		{
			name:       "Success status true",
			id:         "1",
			task:       task,
			expectCode: http.StatusOK,
		},
		{
			name: "Success status false",
			id:   "1",
			task: &models.Task{
				ID:          1,
				Title:       "updated_title",
				Description: "updated_description",
				DueDate:     now,
				Status:      false,
				CreatedAt:   now,
				UpdatedAt:   now,
			},
			expectCode: http.StatusOK,
		},
		{
			name:       "Invalid ID format",
			id:         "abc",
			task:       task,
			respError:  "invalid id",
			expectCode: http.StatusBadRequest,
		},
		{
			name: "Empty title",
			id:   "1",
			task: &models.Task{
				ID:          1,
				Title:       "",
				Description: "updated_description",
				DueDate:     now,
				Status:      true,
				CreatedAt:   now,
				UpdatedAt:   now,
			},
			respError:  "field title is a required field",
			expectCode: http.StatusBadRequest,
		},
		{
			name: "Empty due_date",
			id:   "1",
			task: &models.Task{
				ID:          1,
				Title:       "updated_title",
				Description: "updated_description",
				DueDate:     time.Time{},
				Status:      true,
				CreatedAt:   now,
				UpdatedAt:   now,
			},
			respError:  "field due_date is a required field",
			expectCode: http.StatusBadRequest,
		},
		{
			name:       "Task not found",
			id:         "1",
			task:       task,
			mockError:  storage.ErrTaskNotFound,
			respError:  "task not found",
			expectCode: http.StatusNotFound,
		},
		{
			name:       "Internal error",
			id:         "1",
			task:       task,
			mockError:  errors.New("database error"),
			respError:  "failed to update task",
			expectCode: http.StatusInternalServerError,
		},
	}

	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			taskServiceMock := mocks.NewTaskService(t)

			if tc.expectCode == http.StatusOK || tc.mockError != nil {
				taskServiceMock.On("UpdateTask", mock.AnythingOfType("*models.Task")).
					Return(tc.mockError).
					Once()
			}

			logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
			handler := handlers.UpdateTask(logger, taskServiceMock)

			jsonInput, err := json.Marshal(tc.task)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/tasks/%s", tc.id), bytes.NewReader(jsonInput))
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tc.id)

			ctx := context.WithValue(req.Context(), chi.RouteCtxKey, rctx)
			req = req.WithContext(ctx)

			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			require.Equal(t, tc.expectCode, rr.Code)

			if tc.respError != "" {
				var resp handlers.Response
				err = json.Unmarshal(rr.Body.Bytes(), &resp)
				require.NoError(t, err)
				require.Equal(t, tc.respError, resp.Error)
			}

			taskServiceMock.AssertExpectations(t)
		})
	}
}

func TestDeleteTaskHandler(t *testing.T) {
	cases := []struct {
		name       string
		id         string
		mockError  error
		respError  string
		expectCode int
	}{
		{
			name:       "Success",
			id:         "1",
			expectCode: http.StatusOK,
		},
		{
			name:       "Invalid ID format",
			id:         "abc",
			respError:  "invalid id",
			expectCode: http.StatusBadRequest,
		},
		{
			name:       "Task not found",
			id:         "2",
			mockError:  storage.ErrTaskNotFound,
			respError:  "task not found",
			expectCode: http.StatusNotFound,
		},
		{
			name:       "Internal error",
			id:         "3",
			mockError:  errors.New("database error"),
			respError:  "failed to delete task",
			expectCode: http.StatusInternalServerError,
		},
	}

	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			taskServiceMock := mocks.NewTaskService(t)

			if tc.mockError != nil || tc.expectCode == http.StatusOK {
				id, err := strconv.Atoi(tc.id)
				if err == nil {
					taskServiceMock.On("DeleteTask", uint(id)).
						Return(tc.mockError).
						Once()
				}
			}

			logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
			handler := handlers.DeleteTask(logger, taskServiceMock)

			req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/tasks/%s", tc.id), nil)
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tc.id)

			ctx := context.WithValue(req.Context(), chi.RouteCtxKey, rctx)
			req = req.WithContext(ctx)

			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			require.Equal(t, tc.expectCode, rr.Code)

			if tc.respError != "" {
				var resp handlers.Response
				err := json.Unmarshal(rr.Body.Bytes(), &resp)
				require.NoError(t, err)
				require.Equal(t, tc.respError, resp.Error)
			}

			taskServiceMock.AssertExpectations(t)
		})
	}
}

func TestListTasksHandler(t *testing.T) {
	now := time.Now()
	date := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
	task := models.Task{
		ID:          1,
		Title:       "test_title",
		Description: "test_description",
		DueDate:     date,
		Status:      false,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	tasksList := &models.TasksList{
		Data:  []models.Task{task},
		Total: 1,
		Page:  1,
		Limit: 10,
	}

	cases := []struct {
		name        string
		queryParams string
		page        int
		limit       int
		completed   *bool
		date        *time.Time
		mockResp    *models.TasksList
		mockError   error
		respError   string
		expectCode  int
	}{
		{
			name:        "Success default params",
			queryParams: "",
			page:        1,
			limit:       10,
			completed:   nil,
			date:        nil,
			mockResp:    tasksList,
			expectCode:  http.StatusOK,
		},
		{
			name:        "Success with completed true",
			queryParams: "page=1&limit=5&completed=true",
			page:        1,
			limit:       5,
			completed:   boolPtr(true),
			date:        nil,
			mockResp:    tasksList,
			expectCode:  http.StatusOK,
		},
		{
			name:        "Success with date filter",
			queryParams: "page=2&limit=10&date=2025-04-17",
			page:        2,
			limit:       10,
			completed:   nil,
			date:        &date,
			mockResp:    tasksList,
			expectCode:  http.StatusOK,
		},
		{
			name:        "Invalid page",
			queryParams: "page=0&limit=10",
			page:        1,
			limit:       10,
			completed:   nil,
			date:        nil,
			mockResp:    tasksList,
			expectCode:  http.StatusOK,
		},
		{
			name:        "Invalid limit",
			queryParams: "page=1&limit=-5",
			page:        1,
			limit:       10,
			completed:   nil,
			date:        nil,
			mockResp:    tasksList,
			expectCode:  http.StatusOK,
		},
		{
			name:        "Invalid completed",
			queryParams: "completed=invalid",
			respError:   "invalid completed parameter",
			expectCode:  http.StatusBadRequest,
		},
		{
			name:        "Invalid date format",
			queryParams: "date=2025-13-01",
			respError:   "invalid date format, use YYYY-MM-DD",
			expectCode:  http.StatusBadRequest,
		},
		{
			name:        "Internal error",
			queryParams: "page=1&limit=10",
			page:        1,
			limit:       10,
			completed:   nil,
			date:        nil,
			mockError:   errors.New("database error"),
			respError:   "failed to list tasks",
			expectCode:  http.StatusInternalServerError,
		},
	}

	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			taskServiceMock := mocks.NewTaskService(t)

			if tc.mockResp != nil || tc.mockError != nil {
				taskServiceMock.On("List", tc.page, tc.limit, tc.completed, mock.AnythingOfType("*time.Time")).
					Return(tc.mockResp, tc.mockError).
					Once()
			}

			logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
			handler := handlers.List(logger, taskServiceMock)

			req := httptest.NewRequest(http.MethodGet, "/tasks?"+tc.queryParams, nil)
			rctx := chi.NewRouteContext()
			ctx := context.WithValue(req.Context(), chi.RouteCtxKey, rctx)
			req = req.WithContext(ctx)

			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			require.Equal(t, tc.expectCode, rr.Code)

			if tc.respError != "" {
				var resp handlers.Response
				err := json.Unmarshal(rr.Body.Bytes(), &resp)
				require.NoError(t, err)
				require.Equal(t, tc.respError, resp.Error)
			} else if tc.expectCode == http.StatusOK {
				var resp models.TasksList
				err := json.Unmarshal(rr.Body.Bytes(), &resp)
				require.NoError(t, err)
				require.Equal(t, tc.mockResp.Total, resp.Total)
				require.Equal(t, tc.mockResp.Page, resp.Page)
				require.Equal(t, tc.mockResp.Limit, resp.Limit)
				require.Len(t, resp.Data, len(tc.mockResp.Data))
				for i := range resp.Data {
					require.Equal(t, tc.mockResp.Data[i].ID, resp.Data[i].ID)
					require.Equal(t, tc.mockResp.Data[i].Title, resp.Data[i].Title)
					require.Equal(t, tc.mockResp.Data[i].Description, resp.Data[i].Description)
					require.True(t, tc.mockResp.Data[i].DueDate.Equal(resp.Data[i].DueDate))
					require.Equal(t, tc.mockResp.Data[i].Status, resp.Data[i].Status)
					require.True(t, tc.mockResp.Data[i].CreatedAt.Equal(resp.Data[i].CreatedAt))
					require.True(t, tc.mockResp.Data[i].UpdatedAt.Equal(resp.Data[i].UpdatedAt))
				}
			}

			taskServiceMock.AssertExpectations(t)
		})
	}
}

func boolPtr(b bool) *bool {
	return &b
}
