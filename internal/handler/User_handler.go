package handler

import (
	"ApiSimple/internal/entity"
	"ApiSimple/internal/service"
	"encoding/json"
	"log/slog"
	"net/http"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"
)

const userInstrumentationName = "ApiSimple/handler/user"

type UserHandler struct {
	UserService service.UserService
	tracer      trace.Tracer
}

func NewUserHandler(svc service.UserService) *UserHandler {
	return &UserHandler{
		UserService: svc,
		tracer:      otel.Tracer(userInstrumentationName),
	}
}

func (h *UserHandler) startSpan(r *http.Request, spanName string) (*http.Request, trace.Span) {
	ctx, span := h.tracer.Start(r.Context(), spanName,
		trace.WithSpanKind(trace.SpanKindServer),
		trace.WithAttributes(
			semconv.HTTPMethod(r.Method),
			semconv.HTTPTarget(r.URL.Path),
		),
	)
	return r.WithContext(ctx), span
}

func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	r, span := h.startSpan(r, "UserHandler.Register")
	defer span.End()
	ctx := r.Context()

	var user entity.Users
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "invalid request body")
		slog.ErrorContext(ctx, "failed to decode register request", slog.String("error", err.Error()))
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	slog.InfoContext(ctx, "registering user", slog.String("username", user.Username))

	result, err := h.UserService.Register(ctx, user)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to register user")
		slog.ErrorContext(ctx, "failed to register user",
			slog.String("username", user.Username),
			slog.String("error", err.Error()),
		)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	span.SetStatus(codes.Ok, "")
	slog.InfoContext(ctx, "user registered", slog.String("username", result.Username))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(result)
}

func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	r, span := h.startSpan(r, "UserHandler.Login")
	defer span.End()
	ctx := r.Context()

	var request struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "invalid request body")
		slog.ErrorContext(ctx, "failed to decode login request", slog.String("error", err.Error()))
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	slog.InfoContext(ctx, "user login attempt", slog.String("username", request.Username))

	user, token, err := h.UserService.Login(ctx, request.Username, request.Password)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "login failed")
		slog.WarnContext(ctx, "login failed", slog.String("username", request.Username))
		http.Error(w, "invalid username or password", http.StatusUnauthorized)
		return
	}

	span.SetStatus(codes.Ok, "")
	slog.InfoContext(ctx, "login success", slog.String("username", user.Username))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "login success",
		"data":    user,
		"token":   token,
	})
}
