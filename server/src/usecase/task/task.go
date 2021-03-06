package task

import (
	"fmt"

	taskModel "github.com/team-gleam/kiwi-basket/server/src/domain/model/task"
	tokenModel "github.com/team-gleam/kiwi-basket/server/src/domain/model/user/token"
	"github.com/team-gleam/kiwi-basket/server/src/domain/model/user/username"
	taskRepository "github.com/team-gleam/kiwi-basket/server/src/domain/repository/task"
	credentialRepository "github.com/team-gleam/kiwi-basket/server/src/domain/repository/user/credential"
	loginRepository "github.com/team-gleam/kiwi-basket/server/src/domain/repository/user/login"
	credentialUsecase "github.com/team-gleam/kiwi-basket/server/src/usecase/user/credential"
)

type TaskUsecase struct {
	credentialUsecase credentialUsecase.CredentialUsecase
	taskRepository    taskRepository.ITaskRepository
}

func NewTaskUsecase(c credentialRepository.ICredentialRepository,
	l loginRepository.ILoginRepository,
	t taskRepository.ITaskRepository,
) TaskUsecase {
	return TaskUsecase{
		credentialUsecase.NewCredentialUsecase(c, l),
		t,
	}
}

const (
	IDIsNotZero = "ID is not zero"
	InvalidID   = "Invalid ID"
)

func (u TaskUsecase) Add(token tokenModel.Token, task taskModel.Task) error {
	credentialed, err := u.credentialUsecase.HasCredential(token)
	if err != nil {
		return err
	}
	if !credentialed {
		return fmt.Errorf(credentialUsecase.InvalidToken)
	}

	user, err := u.credentialUsecase.Whose(token)
	if err != nil {
		return err
	}

	return u.taskRepository.Create(user, task)
}

func (u TaskUsecase) Delete(token tokenModel.Token, id int) error {
	credentialed, err := u.credentialUsecase.HasCredential(token)
	if err != nil {
		return err
	}
	if !credentialed {
		return fmt.Errorf(credentialUsecase.InvalidToken)
	}

	if id == 0 {
		return fmt.Errorf(IDIsNotZero)
	}
	if id < 0 {
		return fmt.Errorf(InvalidID)
	}

	user, err := u.credentialUsecase.Whose(token)
	if err != nil {
		return err
	}

	tasks, err := u.taskRepository.GetAll(user)
	if err != nil {
		return err
	}

	if !isValidID(user, id, tasks) {
		return fmt.Errorf(InvalidID)
	}

	return u.taskRepository.Remove(user, id)
}

func (u TaskUsecase) DeleteAll(token tokenModel.Token) error {
	credentialed, err := u.credentialUsecase.HasCredential(token)
	if err != nil {
		return err
	}
	if !credentialed {
		return fmt.Errorf(credentialUsecase.InvalidToken)
	}

	user, err := u.credentialUsecase.Whose(token)
	if err != nil {
		return err
	}

	return u.taskRepository.RemoveAll(user)
}

func isValidID(username username.Username, id int, tasks []taskModel.Task) bool {
	for _, task := range tasks {
		if task.ID() == id {
			return true
		}
	}

	return false
}

func (u TaskUsecase) GetAll(token tokenModel.Token) ([]taskModel.Task, error) {
	credentialed, err := u.credentialUsecase.HasCredential(token)
	if err != nil {
		return nil, err
	}
	if !credentialed {
		return nil, fmt.Errorf(credentialUsecase.InvalidToken)
	}

	user, err := u.credentialUsecase.Whose(token)
	if err != nil {
		return nil, err
	}

	return u.taskRepository.GetAll(user)
}
