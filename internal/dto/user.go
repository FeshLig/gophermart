package dto

import (
	"github.com/FeshLig/gophermart/internal/model"
)

type UserCredentials struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

func ToUserModel(creds UserCredentials, hash string) model.User {
	return model.User{
		Login:        creds.Login,
		PasswordHash: hash,
	}
}
