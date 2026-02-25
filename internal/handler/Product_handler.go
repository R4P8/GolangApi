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

const productInstrumentationName = "ApiSimple/handler/product"

type ProductHandler struct {
	ProductService service.ProductService
	tracer         trace.Tracer
}

func NewProductHandler(svc service.ProductService) *ProductHandler {
	return &ProductHandler{
		ProductService: svc,
		tracer:         otel.Tracer(productInstrumentationName),
	}
}

// startSpan creates a child span with common HTTP attributes.
func (h *ProductHandler) startSpan(r *http.Request, spanName string) (*http.Request, trace.Span) {
	ctx, span := h.tracer.Start(r.Context(), spanName,
		trace.WithSpanKind(trace.SpanKindServer),
		trace.WithAttributes(
			semconv.HTTPMethod(r.Method),
			semconv.HTTPTarget(r.URL.Path),
		),
	)
	return r.WithContext(ctx), span
}

func (h *ProductHandler) Create(w http.ResponseWriter, r *http.Request) {
	r, span := h.startSpan(r, "ProductHandler.Create")
	defer span.End()
	ctx := r.Context()

	var product entity.Product
	if err := json.NewDecoder(r.Body).Decode(&product); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "invalid request body")
		slog.ErrorContext(ctx, "failed to decode request", slog.String("error", err.Error()))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	slog.InfoContext(ctx, "creating product", slog.String("name", product.Name))

	result, err := h.ProductService.Create(ctx, product)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to create product")
		slog.ErrorContext(ctx, "failed to create product", slog.String("error", err.Error()))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	span.SetAttributes(attribute.Int("product.id", result.IDProduct))
	span.SetStatus(codes.Ok, "")
	slog.InfoContext(ctx, "product created", slog.Int("id", result.IDProduct))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(result)
}

func (h *ProductHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	r, span := h.startSpan(r, "ProductHandler.GetByID")
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

	span.SetAttributes(attribute.Int("product.id", id))
	slog.InfoContext(ctx, "fetching product", slog.Int("id", id))

	product, err := h.ProductService.GetByID(ctx, id)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "product not found")
		slog.ErrorContext(ctx, "product not found", slog.Int("id", id), slog.String("error", err.Error()))
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	span.SetStatus(codes.Ok, "")
	slog.InfoContext(ctx, "product fetched", slog.Int("id", id))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(product)
}

func (h *ProductHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	r, span := h.startSpan(r, "ProductHandler.GetAll")
	defer span.End()
	ctx := r.Context()

	slog.InfoContext(ctx, "fetching all products")

	products, err := h.ProductService.GetAll(ctx)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to fetch products")
		slog.ErrorContext(ctx, "failed to fetch products", slog.String("error", err.Error()))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	span.SetAttributes(attribute.Int("product.count", len(products)))
	span.SetStatus(codes.Ok, "")
	slog.InfoContext(ctx, "products fetched", slog.Int("count", len(products)))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(products)
}

func (h *ProductHandler) Update(w http.ResponseWriter, r *http.Request) {
	r, span := h.startSpan(r, "ProductHandler.Update")
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

	var product entity.Product
	if err := json.NewDecoder(r.Body).Decode(&product); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "invalid request body")
		slog.ErrorContext(ctx, "failed to decode request", slog.String("error", err.Error()))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	product.IDProduct = id
	span.SetAttributes(attribute.Int("product.id", id))
	slog.InfoContext(ctx, "updating product", slog.Int("id", id))

	result, err := h.ProductService.Update(ctx, product)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to update product")
		slog.ErrorContext(ctx, "failed to update product", slog.Int("id", id), slog.String("error", err.Error()))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	span.SetStatus(codes.Ok, "")
	slog.InfoContext(ctx, "product updated", slog.Int("id", id))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func (h *ProductHandler) Delete(w http.ResponseWriter, r *http.Request) {
	r, span := h.startSpan(r, "ProductHandler.Delete")
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

	span.SetAttributes(attribute.Int("product.id", id))
	slog.InfoContext(ctx, "deleting product", slog.Int("id", id))

	if err := h.ProductService.Delete(ctx, id); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "product not found")
		slog.ErrorContext(ctx, "failed to delete product", slog.Int("id", id), slog.String("error", err.Error()))
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	span.SetStatus(codes.Ok, "")
	slog.InfoContext(ctx, "product deleted", slog.Int("id", id))
	w.WriteHeader(http.StatusNoContent)
}