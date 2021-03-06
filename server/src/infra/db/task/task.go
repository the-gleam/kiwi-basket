package task

import (
	"fmt"
	"time"

	taskModel "github.com/team-gleam/kiwi-basket/server/src/domain/model/task"
	"github.com/team-gleam/kiwi-basket/server/src/domain/model/user/username"
	taskRepository "github.com/team-gleam/kiwi-basket/server/src/domain/repository/task"
	"github.com/team-gleam/kiwi-basket/server/src/infra/db/handler"
)

type TaskRepository struct {
	dbHandler *handler.DbHandler
}

func NewTaskRepository(h *handler.DbHandler) taskRepository.ITaskRepository {
	h.Db.AutoMigrate(Task{})
	return &TaskRepository{h}
}

type Task struct {
	ID       uint `gorm:"primary_key;auto_increment"`
	Username string
	Date     time.Time
	Title    string
}

func toRecord(t taskModel.Task, u username.Username) Task {
	if t.ID() == -1 {
		return Task{0, u.Name(), t.Date(), t.Title()}
	}

	return Task{uint(t.ID()), u.Name(), t.Date(), t.Title()}
}

func fromRecord(t Task) (taskModel.Task, username.Username, error) {
	task, err := taskModel.NewTask(int(t.ID), t.Date.Format(taskModel.Layout), t.Title)
	if err != nil {
		return taskModel.Task{}, username.Username{}, err
	}

	u, err := username.NewUsername(t.Username)
	return task, u, err
}

func (r *TaskRepository) Create(u username.Username, t taskModel.Task) error {
	d := toRecord(t, u)
	return r.dbHandler.Db.Create(&d).Error
}

func (r *TaskRepository) GetAll(u username.Username) ([]taskModel.Task, error) {
	ds := make([]Task, 0)
	err := r.dbHandler.Db.Where("username = ?", u.Name()).Find(&ds).Error
	if err != nil {
		return []taskModel.Task{}, err
	}

	tasks := make([]taskModel.Task, 0)
	for _, d := range ds {
		t, _, err := fromRecord(d)
		if err != nil {
			return tasks, err
		}
		tasks = append(tasks, t)
	}

	return tasks, nil
}

func (r *TaskRepository) Remove(u username.Username, id int) error {
	if id < 1 {
		return fmt.Errorf("invalid id")
	}

	return r.dbHandler.Db.Where("id = ?", uint(id)).Delete(Task{}).Error
}

func (r *TaskRepository) RemoveAll(u username.Username) error {
	return r.dbHandler.Db.Where("username = ?", u.Name()).Delete(Task{}).Error
}
