package repo

import (
	"github.com/ipaas-org/ipaas-backend/model"
)

// Interfaces ends with -er
type (
	//This interface has the methods declarations for the
	//repo components.
	UserRepoer interface {
		GetByID(id string) (*model.User, error)
		GetByStudentID(studentID int) (*model.User, error)
		Insert(u *model.User) (id string, err error)
		Update(u *model.User) error
		DeleteById(id string) error
		DeleteByStudentID(studentID int) error

		GetBy(props map[string]interface{}) []*model.User
	}

	TokenRepoer interface {
		GetByToken(token string) (*model.RefreshToken, error)
		Insert(t *model.RefreshToken) (id string, err error)
		Delete(token string) error

		GetBy(props map[string]interface{}) []*model.RefreshToken
	}

	StateRepoer interface {
		GetByState(state string) (*model.State, error)
		Insert(s *model.State) (id string, err error)
		Delete(state string) error

		GetBy(props map[string]interface{}) []*model.State
	}

	ApplicationRepoer interface {
		Insert(a *model.Application) (id string, err error)
		GetByContainerID(containerID string) (*model.Application, error)
		Update(a *model.Application) error
		Delete(containerID string) error

		GetBy(props map[string]interface{}) []*model.Application
	}
)
