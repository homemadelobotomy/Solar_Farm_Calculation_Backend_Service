package service

import (
	"errors"
	dto "lab/internal/app/DTO"
	"lab/internal/app/config"
	"lab/internal/app/ds"
	"lab/internal/app/role"
	"time"

	"github.com/golang-jwt/jwt"
)

var ErrInternalServerError = errors.New("")

func (s *Service) GetUserData(userId uint) (dto.UserDataResposne, error) {

	user, err := s.repository.GetUser(userId)
	if err != nil {
		return dto.UserDataResposne{}, err
	}

	return dto.UserDataResposne{Login: user.Login,
		ID: user.ID}, nil
}

func (s *Service) AddNewUser(user dto.UserRegistration) (dto.UserDataResposne, error) {
	if len(user.Password) < 8 {
		return dto.UserDataResposne{}, ErrBadRequest
	}
	userId, err := s.repository.AddNewUser(&ds.User{
		Login:       user.Login,
		Password:    user.Password,
		IsModerator: role.User,
	})
	if err != nil {
		return dto.UserDataResposne{}, err
	}
	return s.GetUserData(userId)
}

func (s *Service) ChangeUserData(user dto.ChangeUserData, userId uint) (dto.UserDataResposne, error) {

	err := s.repository.ChangeUserData(userId, user)
	if err != nil {
		return dto.UserDataResposne{}, err
	}
	response, err := s.repository.GetUser(userId)
	if err != nil {
		return dto.UserDataResposne{}, err
	}
	return dto.UserDataResposne{ID: response.ID, Login: response.Login}, nil
}

func (s *Service) Login(user dto.LoginReq, cfg config.JWTConfig) (dto.LoginRes, error) {
	userFromDatabase, err := s.repository.GetUserByLogin(user.Login)
	if err != nil {
		return dto.LoginRes{}, err
	}

	if userFromDatabase.Password == user.Password {
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, &ds.JWTClaims{
			StandardClaims: jwt.StandardClaims{
				ExpiresAt: time.Now().Add(time.Hour).Unix(),
				IssuedAt:  time.Now().Unix(),
				Issuer:    "Admin",
			},
			UserId:      userFromDatabase.ID,
			IsModerator: userFromDatabase.IsModerator,
		})
		if token == nil {
			return dto.LoginRes{}, ErrInternalServerError
		}

		strToken, err := token.SignedString([]byte(cfg.Token))
		if err != nil {
			return dto.LoginRes{}, err
		}
		return dto.LoginRes{
			ExpiresIn:   time.Hour,
			AccessToken: strToken,
			TokenType:   "Bearer",
			IsModerator: userFromDatabase.IsModerator,
		}, nil

	}
	return dto.LoginRes{}, ErrForbidden
}
