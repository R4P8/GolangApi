package repository

import (
	"ApiSimple/internal/entity"
	"context"
	"database/sql"
	"errors"
	"time"
)

type CategoryRepositoryImpl struct {
	DB *sql.DB
}

func NewCategoryRepository(db *sql.DB) CategoryRepo {
	return &CategoryRepositoryImpl{DB: db}
}

func (r *CategoryRepositoryImpl) Create(ctx context.Context, category entity.Category) (entity.Category, error) {

	query := `
		INSERT INTO category (name, created_at, updated_at)
		VALUES ($1, $2, $3)
		RETURNING id_category
	`

	now := time.Now()
	category.CreatedAt = now
	category.UpdatedAt = now

	err := r.DB.QueryRowContext(
		ctx,
		query,
		category.Name,
		category.CreatedAt,
		category.UpdatedAt,
	).Scan(&category.IDCategory)

	if err != nil {
		return category, err
	}

	return category, nil
}

func (r *CategoryRepositoryImpl) GetByID(ctx context.Context, id int) (entity.Category, error) {

	var category entity.Category

	query := `
		SELECT id_category, name, created_at, updated_at
		FROM category
		WHERE id_category = $1
	`

	err := r.DB.QueryRowContext(ctx, query, id).Scan(
		&category.IDCategory,
		&category.Name,
		&category.CreatedAt,
		&category.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return category, errors.New("category not found")
		}
		return category, err
	}

	return category, nil
}

func (r *CategoryRepositoryImpl) GetAll(ctx context.Context) ([]entity.Category, error) {

	query := `
		SELECT id_category, name, created_at, updated_at
		FROM category
		ORDER BY id_category ASC
	`

	rows, err := r.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []entity.Category

	for rows.Next() {
		var category entity.Category

		err := rows.Scan(
			&category.IDCategory,
			&category.Name,
			&category.CreatedAt,
			&category.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		categories = append(categories, category)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return categories, nil
}

func (r *CategoryRepositoryImpl) Update(ctx context.Context, category entity.Category) (entity.Category, error) {

	query := `
		UPDATE category
		SET name = $1,
		    updated_at = $2
		WHERE id_category = $3
	`

	category.UpdatedAt = time.Now()

	result, err := r.DB.ExecContext(
		ctx,
		query,
		category.Name,
		category.UpdatedAt,
		category.IDCategory,
	)

	if err != nil {
		return category, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return category, err
	}

	if rowsAffected == 0 {
		return category, errors.New("category not found")
	}

	return category, nil
}

func (r *CategoryRepositoryImpl) Delete(ctx context.Context, id int) error {

	query := `DELETE FROM category WHERE id_category = $1`

	result, err := r.DB.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("category not found")
	}

	return nil
}
