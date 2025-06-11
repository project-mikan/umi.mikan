package model

import (
	"github.com/google/uuid"
	"github.com/project-mikan/umi.mikan/backend/infrastructure/database"
)

type User struct {
	ID       uuid.UUID
	Email    string
	Name     string
	AuthType AuthType
}

func GenUser(email, name string, authType AuthType) *User {
	return &User{
		ID:       uuid.New(),
		Email:    email,
		Name:     name,
		AuthType: authType,
	}
}

func (u *User) ConvertToDBModel() database.User {
	return database.User{
		ID:       u.ID,
		Email:    u.Email,
		Name:     u.Name,
		AuthType: u.AuthType.Int16(),
	}
}
