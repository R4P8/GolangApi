package repository

import (
	"ApiSimple/internal/entity"
	"context"
)

type CategoryRepo interface {
	Create(ctx context.Context, category entity.Category) (entity.Category, error)
	GetByID(ctx context.Context, id int) (entity.Category, error)
	GetAll(ctx context.Context) ([]entity.Category, error)
	Update(ctx context.Context, category entity.Category) (entity.Category, error)
	Delete(ctx context.Context, id int) error
}
