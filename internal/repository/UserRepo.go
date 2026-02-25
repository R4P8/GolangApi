package repository

import (
	"ApiSimple/internal/entity"
	"context"
)

type UserRepo interface {
	Register(ctx context.Context, users entity.Users) (entity.Users, error)
	Login(ctx context.Context, username string, password string) (entity.Users, error)
	GetByUsername(ctx context.Context, username string) (entity.Users, error)
}
