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

type CategoryService interface {
	Create(ctx context.Context, category entity.Category) (entity.Category, error)
	GetByID(ctx context.Context, id int) (entity.Category, error)
	GetAll(ctx context.Context) ([]entity.Category, error)
	Update(ctx context.Context, category entity.Category) (entity.Category, error)
	Delete(ctx context.Context, id int) error
}

type CategoryServiceImpl struct {
	CategoryRepo repository.CategoryRepo

	tracer       trace.Tracer
	errorCounter metric.Int64Counter
}

func NewCategoryService(repo repository.CategoryRepo) CategoryService {
	tracer := otel.Tracer("category-service")
	meter := otel.Meter("category-service")

	errorCounter, _ := meter.Int64Counter(
		"category.service.errors.total",
		metric.WithDescription("Total business logic errors in category service"),
	)

	return &CategoryServiceImpl{
		CategoryRepo: repo,
		tracer:       tracer,
		errorCounter: errorCounter,
	}
}

func (s *CategoryServiceImpl) recordError(ctx context.Context, span trace.Span, operation string, err error, start time.Time) {
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

func (s *CategoryServiceImpl) Create(ctx context.Context, category entity.Category) (entity.Category, error) {
	start := time.Now()
	ctx, span := s.tracer.Start(ctx, "CategoryService.Create")
	defer span.End()

	slog.InfoContext(ctx, "creating category", slog.String("name", category.Name))

	result, err := s.CategoryRepo.Create(ctx, category)
	if err != nil {
		s.recordError(ctx, span, "create", err, start)
		return result, err
	}

	span.SetStatus(codes.Ok, "")
	slog.InfoContext(ctx, "category created",
		slog.Int("id", result.IDCategory),
		slog.Float64("duration_ms", float64(time.Since(start).Milliseconds())),
	)
	return result, nil
}

func (s *CategoryServiceImpl) GetByID(ctx context.Context, id int) (entity.Category, error) {
	start := time.Now()
	ctx, span := s.tracer.Start(ctx, "CategoryService.GetByID")
	defer span.End()

	span.SetAttributes(attribute.Int("category.id", id))
	slog.InfoContext(ctx, "fetching category", slog.Int("id", id))

	result, err := s.CategoryRepo.GetByID(ctx, id)
	if err != nil {
		s.recordError(ctx, span, "get_by_id", err, start)
		return result, err
	}

	span.SetStatus(codes.Ok, "")
	slog.InfoContext(ctx, "category fetched",
		slog.Int("id", id),
		slog.Float64("duration_ms", float64(time.Since(start).Milliseconds())),
	)
	return result, nil
}

func (s *CategoryServiceImpl) GetAll(ctx context.Context) ([]entity.Category, error) {
	start := time.Now()
	ctx, span := s.tracer.Start(ctx, "CategoryService.GetAll")
	defer span.End()

	slog.InfoContext(ctx, "fetching all categories")

	result, err := s.CategoryRepo.GetAll(ctx)
	if err != nil {
		s.recordError(ctx, span, "get_all", err, start)
		return nil, err
	}

	span.SetAttributes(attribute.Int("category.count", len(result)))
	span.SetStatus(codes.Ok, "")
	slog.InfoContext(ctx, "categories fetched",
		slog.Int("count", len(result)),
		slog.Float64("duration_ms", float64(time.Since(start).Milliseconds())),
	)
	return result, nil
}

func (s *CategoryServiceImpl) Update(ctx context.Context, category entity.Category) (entity.Category, error) {
	start := time.Now()
	ctx, span := s.tracer.Start(ctx, "CategoryService.Update")
	defer span.End()

	span.SetAttributes(attribute.Int("category.id", category.IDCategory))
	slog.InfoContext(ctx, "updating category", slog.Int("id", category.IDCategory))

	result, err := s.CategoryRepo.Update(ctx, category)
	if err != nil {
		s.recordError(ctx, span, "update", err, start)
		return result, err
	}

	span.SetStatus(codes.Ok, "")
	slog.InfoContext(ctx, "category updated",
		slog.Int("id", category.IDCategory),
		slog.Float64("duration_ms", float64(time.Since(start).Milliseconds())),
	)
	return result, nil
}

func (s *CategoryServiceImpl) Delete(ctx context.Context, id int) error {
	start := time.Now()
	ctx, span := s.tracer.Start(ctx, "CategoryService.Delete")
	defer span.End()

	span.SetAttributes(attribute.Int("category.id", id))
	slog.InfoContext(ctx, "deleting category", slog.Int("id", id))

	if err := s.CategoryRepo.Delete(ctx, id); err != nil {
		s.recordError(ctx, span, "delete", err, start)
		return err
	}

	span.SetStatus(codes.Ok, "")
	slog.InfoContext(ctx, "category deleted",
		slog.Int("id", id),
		slog.Float64("duration_ms", float64(time.Since(start).Milliseconds())),
	)
	return nil
}
