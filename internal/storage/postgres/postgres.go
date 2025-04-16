package postgres

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"

	"todo/internal/models"
	"todo/internal/storage"
)

type Storage struct {
	db *sql.DB
}

func New(host, port, user, password, dbname string) (*Storage, error) {
	const op = "storage.postgres.New"

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
	var db *sql.DB
	var err error

	for attempts := 5; attempts > 0; attempts-- {
		db, err = sql.Open("postgres", connStr)
		if err != nil {
			return nil, fmt.Errorf("%s: open database: %w", op, err)
		}

		err = db.Ping()
		if err == nil {
			break
		}
		time.Sleep(2 * time.Second)
	}

	if err != nil {
		return nil, fmt.Errorf("%s: ping database: %w", op, err)
	}

	schema := `
	CREATE TABLE IF NOT EXISTS tasks (
		id BIGSERIAL PRIMARY KEY,
		title VARCHAR(255) NOT NULL,
		description TEXT,
		due_date TIMESTAMP WITH TIME ZONE NOT NULL,
		completed BOOLEAN DEFAULT FALSE,
		created_at TIMESTAMP WITH TIME ZONE NOT NULL,
		updated_at TIMESTAMP WITH TIME ZONE NOT NULL
	);
	CREATE INDEX IF NOT EXISTS idx_tasks_due_date ON tasks (due_date);
	CREATE INDEX IF NOT EXISTS idx_tasks_completed_due_date ON tasks (completed, due_date);
	`

	_, err = db.Exec(schema)
	if err != nil {
		return nil, fmt.Errorf("%s: execute schema: %w", op, err)
	}

	return &Storage{db: db}, nil
}

func (s *Storage) CreateTask(task models.Task) error {
	const op = "storage.postgres.Create"

	query := `
		INSERT INTO tasks (title, description, due_date, completed, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id`

	now := time.Now()
	task.CreatedAt = now
	task.UpdatedAt = now

	err := s.db.QueryRow(
		query,
		task.Title,
		task.Description,
		task.DueDate,
		task.Status,
		task.CreatedAt,
		task.UpdatedAt,
	).Scan(&task.ID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Storage) GetByID(id uint) (*models.Task, error) {
	const op = "storage.postgres.GetByID"

	query := `
		SELECT id, title, description, due_date, completed, created_at, updated_at
		FROM tasks
		WHERE id = $1`

	task := &models.Task{}
	err := s.db.QueryRow(query, id).Scan(
		&task.ID,
		&task.Title,
		&task.Description,
		&task.DueDate,
		&task.Status,
		&task.CreatedAt,
		&task.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, storage.ErrTaskNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return task, nil
}

func (s *Storage) UpdateTask(task *models.Task) error {
	const op = "storage.postgres.UpdateTask"

	query := `
		UPDATE tasks
		SET title = $1, description = $2, due_date = $3, completed = $4, updated_at = $5
		WHERE id = $6`

	task.UpdatedAt = time.Now()

	result, err := s.db.Exec(
		query,
		task.Title,
		task.Description,
		task.DueDate,
		task.Status,
		task.UpdatedAt,
		task.ID,
	)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s: get rows affected: %w", op, err)
	}

	if rowsAffected == 0 {
		return storage.ErrTaskNotFound
	}

	return nil
}

func (s *Storage) DeleteTask(id uint) error {
	const op = "storage.postgres.DeleteTask"

	query := `DELETE FROM tasks WHERE id = $1`

	result, err := s.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s: get rows affected: %w", op, err)
	}

	if rowsAffected == 0 {
		return storage.ErrTaskNotFound
	}

	return nil
}

func (s *Storage) List(page, limit int, completed *bool, date *time.Time) (*models.TasksList, error) {
	const op = "storage.postgres.List"

	var args []interface{}
	var conditions []string
	argPosition := 1

	query := `
		SELECT id, title, description, due_date, completed, created_at, updated_at
		FROM tasks
		WHERE 1=1`

	if completed != nil {
		conditions = append(conditions, fmt.Sprintf(" AND completed = $%d", argPosition))
		args = append(args, *completed)
		argPosition++
	}

	if date != nil {
		conditions = append(conditions, fmt.Sprintf(" AND DATE(due_date) = DATE($%d)", argPosition))
		args = append(args, *date)
		argPosition++
	}

	for _, condition := range conditions {
		query += condition
	}

	query += fmt.Sprintf(" ORDER BY due_date ASC LIMIT $%d OFFSET $%d", argPosition, argPosition+1)
	args = append(args, limit, (page-1)*limit)

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	var tasks []models.Task
	for rows.Next() {
		var task models.Task
		err := rows.Scan(
			&task.ID,
			&task.Title,
			&task.Description,
			&task.DueDate,
			&task.Status,
			&task.CreatedAt,
			&task.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		tasks = append(tasks, task)
	}

	var total int64
	countQuery := "SELECT COUNT(*) FROM tasks WHERE 1=1"
	for _, condition := range conditions {
		countQuery += condition
	}

	err = s.db.QueryRow(countQuery, args[:len(args)-2]...).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &models.TasksList{
		Data:  tasks,
		Total: total,
		Page:  page,
		Limit: limit,
	}, nil
}
