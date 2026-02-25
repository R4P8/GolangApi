package repository

import (
	"ApiSimple/internal/entity"
	"context"
)

type ProductRepo interface {
	Create(ctx context.Context, product entity.Product) (entity.Product, error)
	GetByID(ctx context.Context, id int) (entity.Product, error)
	GetAll(ctx context.Context) ([]entity.Product, error)
	Update(ctx context.Context, product entity.Product) (entity.Product, error)
	Delete(ctx context.Context, id int) error
}
