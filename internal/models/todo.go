package models

import "time"

// Task представляет задачу в системе
// @Description Задача пользователя
type Task struct {
	ID          int64     `json:"id" example:"1"` // Уникальный идентификатор задачи
	Title       string    `json:"title" validate:"required" example:"Купить молоко"` // Заголовок задачи
	Description string    `json:"description" example:"Купить 2 литра молока в магазине"` // Описание задачи
	DueDate     time.Time `json:"due_date" validate:"required" example:"2025-04-20T15:00:00Z"` // Дата выполнения
	Status      bool      `json:"status" example:"false"` // Статус выполнения (true - выполнена, false - не выполнена)
	CreatedAt   time.Time `json:"created_at" example:"2025-04-17T10:30:00Z"` // Дата создания
	UpdatedAt   time.Time `json:"updated_at" example:"2025-04-17T10:30:00Z"` // Дата обновления
}

// TasksList представляет список задач с пагинацией
// @Description Список задач с пагинацией
type TasksList struct {
	Data  []Task `json:"data"` // Массив задач
	Total int64  `json:"total" example:"42"` // Общее количество задач
	Page  int    `json:"page" example:"1"` // Текущая страница
	Limit int    `json:"limit" example:"10"` // Количество элементов на странице
}

type Response struct {
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}
