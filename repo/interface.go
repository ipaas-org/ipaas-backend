package repo

import (
	"github.com/ipaas-org/ipaas-backend/model"
)

// Interfaces ends with -er
type (
	//This interface has the methods declarations for the
	//repo components.
	UserRepoer interface {
		GetByID(id int) (*model.User, error)
		GetByStudentID(studentID int) (*model.User, error)
		GetByEmail(email string) (*model.User, error)
		Insert(u *model.User) (id int, err error)
		Update(u *model.User) error
		Delete(id int) error

		GetAll() []*model.User
	}

	TokenRepoer interface {
		Insert(t *model.RefreshToken) (id int, err error)
		GetByToken(token string) (*model.RefreshToken, error)
		Delete(token string) error
	}
)
