package repository

import (
	"ApiSimple/internal/entity"
	"context"
	"database/sql"
	"errors"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type ProductRepositoryImpl struct {
	DB     *sql.DB
	Tracer trace.Tracer
}

func NewProductRepository(db *sql.DB) ProductRepo {
	return &ProductRepositoryImpl{
		DB:     db,
		Tracer: otel.Tracer("product-repository"),
	}
}

func (r *ProductRepositoryImpl) Create(ctx context.Context, product entity.Product) (entity.Product, error) {
	ctx, span := r.Tracer.Start(ctx, "ProductRepository.Create")
	defer span.End()

	query := `
		INSERT INTO product (id_category, name, qty, description,  created_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW())
		RETURNING id_product
	`

	span.SetAttributes(
		attribute.String("db.system", "postgresql"),
		attribute.String("db.operation", "INSERT"),
	)

	err := r.DB.QueryRowContext(
		ctx,
		query,
		product.IDCategory,
		product.Name,
		product.Qty,
		product.Description,
	).Scan(&product.IDProduct)

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return product, err
	}

	return product, nil
}

func (r *ProductRepositoryImpl) GetByID(ctx context.Context, id int) (entity.Product, error) {
	ctx, span := r.Tracer.Start(ctx, "ProductRepository.GetByID")
	defer span.End()

	var product entity.Product

	query := `
		SELECT id_product, id_category, name, qty, description
		FROM product
		WHERE id_product = $1
	`

	span.SetAttributes(
		attribute.String("db.system", "postgresql"),
		attribute.String("db.operation", "SELECT"),
		attribute.Int("product.id", id),
	)

	err := r.DB.QueryRowContext(ctx, query, id).Scan(
		&product.IDProduct,
		&product.IDCategory,
		&product.Name,
		&product.Qty,
		&product.Description,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = errors.New("product not found")
		}
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return product, err
	}

	return product, nil
}

func (r *ProductRepositoryImpl) GetAll(ctx context.Context) ([]entity.Product, error) {
	ctx, span := r.Tracer.Start(ctx, "ProductRepository.GetAll")
	defer span.End()

	query := `
		SELECT id_product, id_category, name, qty, description
		FROM product
	`

	span.SetAttributes(
		attribute.String("db.system", "postgresql"),
		attribute.String("db.operation", "SELECT"),
	)

	rows, err := r.DB.QueryContext(ctx, query)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}
	defer rows.Close()

	var products []entity.Product

	for rows.Next() {
		var product entity.Product
		err := rows.Scan(
			&product.IDProduct,
			&product.IDCategory,
			&product.Name,
			&product.Qty,
			&product.Description,
		)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return nil, err
		}
		products = append(products, product)
	}

	return products, nil
}

func (r *ProductRepositoryImpl) Update(ctx context.Context, product entity.Product) (entity.Product, error) {
	ctx, span := r.Tracer.Start(ctx, "ProductRepository.Update")
	defer span.End()

	query := `
		UPDATE product
		SET id_category = $1,
		    name = $2,
		    qty = $3,
		    description = $4
		WHERE id_product = $5
	`

	span.SetAttributes(
		attribute.String("db.system", "postgresql"),
		attribute.String("db.operation", "UPDATE"),
		attribute.Int("product.id", product.IDProduct),
	)

	result, err := r.DB.ExecContext(
		ctx,
		query,
		product.IDCategory,
		product.Name,
		product.Qty,
		product.Description,
		product.IDProduct,
	)

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return product, err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		err := errors.New("product not found")
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return product, err
	}

	return product, nil
}

func (r *ProductRepositoryImpl) Delete(ctx context.Context, id int) error {
	ctx, span := r.Tracer.Start(ctx, "ProductRepository.Delete")
	defer span.End()

	query := `DELETE FROM product WHERE id_product = $1`

	span.SetAttributes(
		attribute.String("db.system", "postgresql"),
		attribute.String("db.operation", "DELETE"),
		attribute.Int("product.id", id),
	)

	result, err := r.DB.ExecContext(ctx, query, id)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		err := errors.New("product not found")
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	return nil
}
