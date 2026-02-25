package handler

import (
	"ApiSimple/internal/entity"
	"ApiSimple/internal/service"
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"
)

const instrumentationName = "ApiSimple/handler/category"

type CategoryHandler struct {
	CategoryService service.CategoryService
	tracer          trace.Tracer
}

func NewCategoryHandler(svc service.CategoryService) *CategoryHandler {
	return &CategoryHandler{
		CategoryService: svc,
		tracer:          otel.Tracer(instrumentationName),
	}
}

func (h *CategoryHandler) startSpan(r *http.Request, spanName string) (*http.Request, trace.Span) {
	ctx, span := h.tracer.Start(r.Context(), spanName,
		trace.WithSpanKind(trace.SpanKindServer),
		trace.WithAttributes(
			semconv.HTTPMethod(r.Method),
			semconv.HTTPTarget(r.URL.Path),
		),
	)
	return r.WithContext(ctx), span
}

func (h *CategoryHandler) Create(w http.ResponseWriter, r *http.Request) {
	r, span := h.startSpan(r, "CategoryHandler.Create")
	defer span.End()
	ctx := r.Context()

	var category entity.Category
	if err := json.NewDecoder(r.Body).Decode(&category); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "invalid request body")
		slog.ErrorContext(ctx, "failed to decode request", slog.String("error", err.Error()))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	slog.InfoContext(ctx, "creating category", slog.String("name", category.Name))

	result, err := h.CategoryService.Create(ctx, category)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to create category")
		slog.ErrorContext(ctx, "failed to create category", slog.String("error", err.Error()))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	span.SetAttributes(attribute.Int("category.id", result.IDCategory))
	span.SetStatus(codes.Ok, "")
	slog.InfoContext(ctx, "category created", slog.Int("id", result.IDCategory))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(result)
}

func (h *CategoryHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	r, span := h.startSpan(r, "CategoryHandler.GetByID")
	defer span.End()
	ctx := r.Context()

	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "invalid id")
		slog.ErrorContext(ctx, "invalid id param", slog.String("error", err.Error()))
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	span.SetAttributes(attribute.Int("category.id", id))
	slog.InfoContext(ctx, "fetching category", slog.Int("id", id))

	category, err := h.CategoryService.GetByID(ctx, id)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "category not found")
		slog.ErrorContext(ctx, "category not found", slog.Int("id", id), slog.String("error", err.Error()))
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	span.SetStatus(codes.Ok, "")
	slog.InfoContext(ctx, "category fetched", slog.Int("id", id))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(category)
}

func (h *CategoryHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	r, span := h.startSpan(r, "CategoryHandler.GetAll")
	defer span.End()
	ctx := r.Context()

	slog.InfoContext(ctx, "fetching all categories")

	categories, err := h.CategoryService.GetAll(ctx)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to fetch categories")
		slog.ErrorContext(ctx, "failed to fetch categories", slog.String("error", err.Error()))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	span.SetAttributes(attribute.Int("category.count", len(categories)))
	span.SetStatus(codes.Ok, "")
	slog.InfoContext(ctx, "categories fetched", slog.Int("count", len(categories)))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(categories)
}

func (h *CategoryHandler) Update(w http.ResponseWriter, r *http.Request) {
	r, span := h.startSpan(r, "CategoryHandler.Update")
	defer span.End()
	ctx := r.Context()

	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "invalid id")
		slog.ErrorContext(ctx, "invalid id param", slog.String("error", err.Error()))
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	var category entity.Category
	if err := json.NewDecoder(r.Body).Decode(&category); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "invalid request body")
		slog.ErrorContext(ctx, "failed to decode request", slog.String("error", err.Error()))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	category.IDCategory = id
	span.SetAttributes(attribute.Int("category.id", id))
	slog.InfoContext(ctx, "updating category", slog.Int("id", id))

	result, err := h.CategoryService.Update(ctx, category)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to update category")
		slog.ErrorContext(ctx, "failed to update category", slog.Int("id", id), slog.String("error", err.Error()))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	span.SetStatus(codes.Ok, "")
	slog.InfoContext(ctx, "category updated", slog.Int("id", id))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func (h *CategoryHandler) Delete(w http.ResponseWriter, r *http.Request) {
	r, span := h.startSpan(r, "CategoryHandler.Delete")
	defer span.End()
	ctx := r.Context()

	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "invalid id")
		slog.ErrorContext(ctx, "invalid id param", slog.String("error", err.Error()))
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	span.SetAttributes(attribute.Int("category.id", id))
	slog.InfoContext(ctx, "deleting category", slog.Int("id", id))

	if err := h.CategoryService.Delete(ctx, id); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "category not found")
		slog.ErrorContext(ctx, "failed to delete category", slog.Int("id", id), slog.String("error", err.Error()))
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	span.SetStatus(codes.Ok, "")
	slog.InfoContext(ctx, "category deleted", slog.Int("id", id))
	w.WriteHeader(http.StatusNoContent)
}