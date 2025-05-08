package services

import (
	"context"
	"database/sql"
	"errors"

	"template/internal/repository"
	"template/utils"
)

var ErrEmailAlreadyExist = errors.New("email already exists")

type UserService struct {
	UR *repository.UserRepo
}

func NewUserService(ur *repository.UserRepo) *UserService {
	return &UserService{UR: ur}
}

func (us *UserService) Create(ctx context.Context, user repository.User) (*repository.User, error) {
	exist, _ := us.UR.GetUserByEmail(ctx, user.Email)
	if exist != nil && exist.ID != "" {
		return nil, ErrEmailAlreadyExist
	}
	// Make PasswordHash nullable
	hashedPassword := sql.NullString{}
	if user.PasswordHash.Valid {
		hash, _ := utils.HashPassword(user.PasswordHash.String)
		hashedPassword = sql.NullString{String: hash, Valid: true}
	}
	user.PasswordHash = hashedPassword
	return us.UR.CreateUser(ctx, &user)
}

func (us *UserService) RegisterGoogleUser(ctx context.Context, info *repository.GoogleUser) (*repository.User, error) {
	exist, _ := us.UR.GetUserByGoogleID(ctx, info.Id)
	if exist != nil && exist.ID != "" {
		return exist, nil
	}
	u := &repository.User{
		Email:    info.Email,
		GoogleID: sql.NullString{String: info.Id, Valid: true},
	}
	return us.UR.CreateUser(ctx, u)
}

func (us *UserService) Delete(ctx context.Context, user *repository.User) error {
	return us.UR.DeleteUser(ctx, user.ID)
}
