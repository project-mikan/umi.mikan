package model

import (
	"github.com/project-mikan/umi.mikan/backend/infrastructure/database"
)

type AuthType int

const (
	AuthTypeEmailPassword AuthType = iota
)

type AuthTableType database.UserPasswordAuthe

func (a AuthType) Int16() int16 {
	return int16(a)
}

func GetAuthTypeFromInt16(authType int16) AuthType {
	switch authType {
	case AuthTypeEmailPassword.Int16():
		return AuthTypeEmailPassword
	default:
		return AuthTypeEmailPassword
	}
}
