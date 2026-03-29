package mapper

import (
	"github.com/FeshLig/gophermart/internal/dto"
	"github.com/FeshLig/gophermart/internal/model"
)

func ToUserModel(creds dto.UserCredentials, hash string) model.User {
	return model.User{
		Login:        creds.Login,
		PasswordHash: hash,
	}
}
