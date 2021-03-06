package task

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-playground/validator"
	"github.com/labstack/echo/v4"
	TaskModel "github.com/team-gleam/kiwi-basket/server/src/domain/model/task"
	"github.com/team-gleam/kiwi-basket/server/src/domain/model/user/token"
	taskRepository "github.com/team-gleam/kiwi-basket/server/src/domain/repository/task"
	credentialRepository "github.com/team-gleam/kiwi-basket/server/src/domain/repository/user/credential"
	loginRepository "github.com/team-gleam/kiwi-basket/server/src/domain/repository/user/login"
	errorResponse "github.com/team-gleam/kiwi-basket/server/src/interfaces/controllers/error"
	taskUsecase "github.com/team-gleam/kiwi-basket/server/src/usecase/task"
	credentialUsecase "github.com/team-gleam/kiwi-basket/server/src/usecase/user/credential"
)

type TaskController struct {
	taskUsecase taskUsecase.TaskUsecase
}

func NewTaskController(
	c credentialRepository.ICredentialRepository,
	l loginRepository.ILoginRepository,
	t taskRepository.ITaskRepository,
) *TaskController {
	return &TaskController{
		taskUsecase.NewTaskUsecase(c, l, t),
	}
}

const (
	InvalidJSONFormat = "invalid JSON format"
	InvalidID         = "invalid ID"
)

type TaskResponse struct {
	ID    string `json:"id" validate:"required,numeric,ne=0,min=-1"`
	Date  string `json:"date" validate:"required"`
	Title string `json:"title" validate:"required,max=85"`
}

func (t TaskResponse) Validates() bool {
	return validator.New().Struct(t) == nil
}

func (t TaskResponse) toTask() (TaskModel.Task, error) {
	id, err := strconv.Atoi(t.ID)
	if err != nil {
		return TaskModel.Task{}, err
	}
	if id == 0 {
		return TaskModel.Task{}, fmt.Errorf(InvalidID)
	}

	return TaskModel.NewTask(id, t.Date, t.Title)
}

func (c TaskController) Add(ctx echo.Context) error {
	t := ctx.Request().Header.Get("Token")
	if t == "" {
		return ctx.JSON(
			http.StatusUnauthorized,
			errorResponse.NewError(fmt.Errorf(credentialUsecase.InvalidToken)),
		)
	}

	res := new(TaskResponse)
	err := ctx.Bind(res)
	if err != nil {
		return ctx.JSON(
			http.StatusBadRequest,
			errorResponse.NewError(fmt.Errorf(InvalidJSONFormat)),
		)
	}

	if !res.Validates() {
		return ctx.JSON(
			http.StatusBadRequest,
			errorResponse.NewError(fmt.Errorf(InvalidJSONFormat)),
		)
	}

	task, err := res.toTask()
	if err != nil {
		return ctx.JSON(
			http.StatusBadRequest,
			errorResponse.NewError(fmt.Errorf(InvalidJSONFormat)),
		)
	}

	err = c.taskUsecase.Add(token.NewToken(t), task)
	if err != nil && err.Error() == credentialUsecase.InvalidToken {
		return ctx.JSON(
			http.StatusUnauthorized,
			errorResponse.NewError(err),
		)
	}
	if err != nil {
		return ctx.JSON(
			http.StatusInternalServerError,
			errorResponse.NewError(fmt.Errorf(errorResponse.InternalServerError)),
		)
	}

	return ctx.NoContent(http.StatusOK)
}

type IDResponse struct {
	ID string `json:"id" validate:"required,numeric,ne=0,min=-1"`
}

func (i IDResponse) Validates() bool {
	return validator.New().Struct(i) == nil
}

func (c TaskController) Delete(ctx echo.Context) error {
	t := ctx.Request().Header.Get("Token")
	if t == "" {
		return ctx.JSON(
			http.StatusUnauthorized,
			errorResponse.NewError(fmt.Errorf(credentialUsecase.InvalidToken)),
		)
	}

	res := new(IDResponse)
	err := ctx.Bind(res)
	if err != nil || !res.Validates() {
		return ctx.JSON(
			http.StatusBadRequest,
			errorResponse.NewError(fmt.Errorf(InvalidJSONFormat)),
		)
	}

	id, err := strconv.Atoi(res.ID)
	if err != nil {
		return ctx.JSON(
			http.StatusInternalServerError,
			errorResponse.NewError(fmt.Errorf(errorResponse.InternalServerError)),
		)
	}

	err = c.taskUsecase.Delete(token.NewToken(t), id)
	if err != nil && err.Error() == credentialUsecase.InvalidToken {
		return ctx.JSON(
			http.StatusUnauthorized,
			errorResponse.NewError(fmt.Errorf(credentialUsecase.InvalidToken)),
		)
	}
	if err != nil && (err.Error() == taskUsecase.IDIsNotZero ||
		err.Error() == taskUsecase.InvalidID) {
		return ctx.JSON(
			http.StatusBadRequest,
			errorResponse.NewError(err),
		)
	}
	if err != nil {
		return ctx.JSON(
			http.StatusInternalServerError,
			errorResponse.NewError(fmt.Errorf(errorResponse.InternalServerError)),
		)
	}

	return ctx.NoContent(http.StatusOK)
}

type TasksResponse struct {
	Tasks []TaskResponse `json:"tasks"`
}

func toTasksResponse(ts []TaskModel.Task) TasksResponse {
	res := []TaskResponse{}
	for _, t := range ts {
		res = append(res, TaskResponse{
			ID:    strconv.Itoa(t.ID()),
			Date:  t.TextDate(),
			Title: t.Title(),
		})
	}

	return TasksResponse{res}
}

func (c TaskController) GetAll(ctx echo.Context) error {
	t := ctx.Request().Header.Get("Token")
	if t == "" {
		return ctx.JSON(
			http.StatusUnauthorized,
			errorResponse.NewError(fmt.Errorf(credentialUsecase.InvalidToken)),
		)
	}

	tasks, err := c.taskUsecase.GetAll(token.NewToken(t))
	if err != nil && err.Error() == credentialUsecase.InvalidToken {
		return ctx.JSON(
			http.StatusUnauthorized,
			errorResponse.NewError(err),
		)
	}
	if err != nil {
		return ctx.JSON(
			http.StatusInternalServerError,
			errorResponse.NewError(fmt.Errorf(errorResponse.InternalServerError)),
		)
	}

	return ctx.JSON(http.StatusOK, toTasksResponse(tasks))
}
