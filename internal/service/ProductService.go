package service

import (
	"ApiSimple/internal/entity"
	"ApiSimple/internal/repository"
	"context"
	"log/slog"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

type ProductService interface {
	Create(ctx context.Context, product entity.Product) (entity.Product, error)
	GetByID(ctx context.Context, id int) (entity.Product, error)
	GetAll(ctx context.Context) ([]entity.Product, error)
	Update(ctx context.Context, product entity.Product) (entity.Product, error)
	Delete(ctx context.Context, id int) error
}

type ProductServiceImpl struct {
	ProductRepo  repository.ProductRepo
	tracer       trace.Tracer
	errorCounter metric.Int64Counter
}

func NewProductService(repo repository.ProductRepo) ProductService {
	tracer := otel.Tracer("product-service")
	meter := otel.Meter("product-service")

	errorCounter, _ := meter.Int64Counter(
		"product.service.errors.total",
		metric.WithDescription("Total business logic errors in product service"),
	)

	return &ProductServiceImpl{
		ProductRepo:  repo,
		tracer:       tracer,
		errorCounter: errorCounter,
	}
}

func (s *ProductServiceImpl) recordError(ctx context.Context, span trace.Span, operation string, err error, start time.Time) {
	span.RecordError(err)
	span.SetStatus(codes.Error, err.Error())
	s.errorCounter.Add(ctx, 1,
		metric.WithAttributes(attribute.String("operation", operation)),
	)
	slog.ErrorContext(ctx, "service error",
		slog.String("operation", operation),
		slog.String("error", err.Error()),
		slog.Float64("duration_ms", float64(time.Since(start).Milliseconds())),
	)
}

func (s *ProductServiceImpl) Create(ctx context.Context, product entity.Product) (entity.Product, error) {
	start := time.Now()
	ctx, span := s.tracer.Start(ctx, "ProductService.Create")
	defer span.End()

	slog.InfoContext(ctx, "creating product", slog.String("name", product.Name))

	result, err := s.ProductRepo.Create(ctx, product)
	if err != nil {
		s.recordError(ctx, span, "create", err, start)
		return result, err
	}

	span.SetAttributes(attribute.Int("product.id", result.IDProduct))
	span.SetStatus(codes.Ok, "")
	slog.InfoContext(ctx, "product created",
		slog.Int("id", result.IDProduct),
		slog.Float64("duration_ms", float64(time.Since(start).Milliseconds())),
	)
	return result, nil
}

func (s *ProductServiceImpl) GetByID(ctx context.Context, id int) (entity.Product, error) {
	start := time.Now()
	ctx, span := s.tracer.Start(ctx, "ProductService.GetByID")
	defer span.End()

	span.SetAttributes(attribute.Int("product.id", id))
	slog.InfoContext(ctx, "fetching product", slog.Int("id", id))

	result, err := s.ProductRepo.GetByID(ctx, id)
	if err != nil {
		s.recordError(ctx, span, "get_by_id", err, start)
		return result, err
	}

	span.SetStatus(codes.Ok, "")
	slog.InfoContext(ctx, "product fetched",
		slog.Int("id", id),
		slog.Float64("duration_ms", float64(time.Since(start).Milliseconds())),
	)
	return result, nil
}

func (s *ProductServiceImpl) GetAll(ctx context.Context) ([]entity.Product, error) {
	start := time.Now()
	ctx, span := s.tracer.Start(ctx, "ProductService.GetAll")
	defer span.End()

	slog.InfoContext(ctx, "fetching all products")

	result, err := s.ProductRepo.GetAll(ctx)
	if err != nil {
		s.recordError(ctx, span, "get_all", err, start)
		return nil, err
	}

	span.SetAttributes(attribute.Int("product.count", len(result)))
	span.SetStatus(codes.Ok, "")
	slog.InfoContext(ctx, "products fetched",
		slog.Int("count", len(result)),
		slog.Float64("duration_ms", float64(time.Since(start).Milliseconds())),
	)
	return result, nil
}

func (s *ProductServiceImpl) Update(ctx context.Context, product entity.Product) (entity.Product, error) {
	start := time.Now()
	ctx, span := s.tracer.Start(ctx, "ProductService.Update")
	defer span.End()

	span.SetAttributes(attribute.Int("product.id", product.IDProduct))
	slog.InfoContext(ctx, "updating product", slog.Int("id", product.IDProduct))

	result, err := s.ProductRepo.Update(ctx, product)
	if err != nil {
		s.recordError(ctx, span, "update", err, start)
		return result, err
	}

	span.SetStatus(codes.Ok, "")
	slog.InfoContext(ctx, "product updated",
		slog.Int("id", product.IDProduct),
		slog.Float64("duration_ms", float64(time.Since(start).Milliseconds())),
	)
	return result, nil
}

func (s *ProductServiceImpl) Delete(ctx context.Context, id int) error {
	start := time.Now()
	ctx, span := s.tracer.Start(ctx, "ProductService.Delete")
	defer span.End()

	span.SetAttributes(attribute.Int("product.id", id))
	slog.InfoContext(ctx, "deleting product", slog.Int("id", id))

	if err := s.ProductRepo.Delete(ctx, id); err != nil {
		s.recordError(ctx, span, "delete", err, start)
		return err
	}

	span.SetStatus(codes.Ok, "")
	slog.InfoContext(ctx, "product deleted",
		slog.Int("id", id),
		slog.Float64("duration_ms", float64(time.Since(start).Milliseconds())),
	)
	return nil
}
