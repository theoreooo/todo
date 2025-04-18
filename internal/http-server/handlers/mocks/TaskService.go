// Code generated by mockery v2.53.3. DO NOT EDIT.

package mocks

import (
	models "todo/internal/models"

	mock "github.com/stretchr/testify/mock"

	time "time"
)

// TaskService is an autogenerated mock type for the TaskService type
type TaskService struct {
	mock.Mock
}

// CreateTask provides a mock function with given fields: task
func (_m *TaskService) CreateTask(task models.Task) error {
	ret := _m.Called(task)

	if len(ret) == 0 {
		panic("no return value specified for CreateTask")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(models.Task) error); ok {
		r0 = rf(task)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// DeleteTask provides a mock function with given fields: id
func (_m *TaskService) DeleteTask(id uint) error {
	ret := _m.Called(id)

	if len(ret) == 0 {
		panic("no return value specified for DeleteTask")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(uint) error); ok {
		r0 = rf(id)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GetByID provides a mock function with given fields: id
func (_m *TaskService) GetByID(id uint) (*models.Task, error) {
	ret := _m.Called(id)

	if len(ret) == 0 {
		panic("no return value specified for GetByID")
	}

	var r0 *models.Task
	var r1 error
	if rf, ok := ret.Get(0).(func(uint) (*models.Task, error)); ok {
		return rf(id)
	}
	if rf, ok := ret.Get(0).(func(uint) *models.Task); ok {
		r0 = rf(id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*models.Task)
		}
	}

	if rf, ok := ret.Get(1).(func(uint) error); ok {
		r1 = rf(id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// List provides a mock function with given fields: page, limit, completed, date
func (_m *TaskService) List(page int, limit int, completed *bool, date *time.Time) (*models.TasksList, error) {
	ret := _m.Called(page, limit, completed, date)

	if len(ret) == 0 {
		panic("no return value specified for List")
	}

	var r0 *models.TasksList
	var r1 error
	if rf, ok := ret.Get(0).(func(int, int, *bool, *time.Time) (*models.TasksList, error)); ok {
		return rf(page, limit, completed, date)
	}
	if rf, ok := ret.Get(0).(func(int, int, *bool, *time.Time) *models.TasksList); ok {
		r0 = rf(page, limit, completed, date)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*models.TasksList)
		}
	}

	if rf, ok := ret.Get(1).(func(int, int, *bool, *time.Time) error); ok {
		r1 = rf(page, limit, completed, date)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// UpdateTask provides a mock function with given fields: task
func (_m *TaskService) UpdateTask(task *models.Task) error {
	ret := _m.Called(task)

	if len(ret) == 0 {
		panic("no return value specified for UpdateTask")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(*models.Task) error); ok {
		r0 = rf(task)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// NewTaskService creates a new instance of TaskService. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewTaskService(t interface {
	mock.TestingT
	Cleanup(func())
}) *TaskService {
	mock := &TaskService{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
