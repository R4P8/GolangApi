package service

import (
	"ApiSimple/internal/entity"
	"ApiSimple/internal/repository"
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/crypto/bcrypt"
)

type UserService interface {
	Register(ctx context.Context, user entity.Users) (entity.Users, error)
	Login(ctx context.Context, username string, password string) (entity.Users, string, error)
}

type UserServiceImpl struct {
	UserRepo     repository.UserRepo
	tracer       trace.Tracer
	errorCounter metric.Int64Counter
}

func NewUserService(repo repository.UserRepo) UserService {
	tracer := otel.Tracer("user-service")
	meter := otel.Meter("user-service")

	errorCounter, _ := meter.Int64Counter(
		"user.service.errors.total",
		metric.WithDescription("Total business logic errors in user service"),
	)

	return &UserServiceImpl{
		UserRepo:     repo,
		tracer:       tracer,
		errorCounter: errorCounter,
	}
}

func (s *UserServiceImpl) recordError(ctx context.Context, span trace.Span, operation string, err error, start time.Time) {
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

func (s *UserServiceImpl) Register(ctx context.Context, user entity.Users) (entity.Users, error) {
	start := time.Now()
	ctx, span := s.tracer.Start(ctx, "UserService.Register")
	defer span.End()

	slog.InfoContext(ctx, "registering user", slog.String("username", user.Username))

	result, err := s.UserRepo.Register(ctx, user)
	if err != nil {
		s.recordError(ctx, span, "register", err, start)
		return result, err
	}

	span.SetAttributes(attribute.Int("user.id", result.IDUser))
	span.SetStatus(codes.Ok, "")
	slog.InfoContext(ctx, "user registered",
		slog.Int("id", result.IDUser),
		slog.String("username", result.Username),
		slog.Float64("duration_ms", float64(time.Since(start).Milliseconds())),
	)
	return result, nil
}

func (s *UserServiceImpl) Login(ctx context.Context, username string, password string) (entity.Users, string, error) {
	start := time.Now()
	ctx, span := s.tracer.Start(ctx, "UserService.Login")
	defer span.End()

	span.SetAttributes(attribute.String("user.username", username))
	slog.InfoContext(ctx, "login attempt", slog.String("username", username))

	//  Cari user by username
	user, err := s.UserRepo.GetByUsername(ctx, username)
	if err != nil {
		s.recordError(ctx, span, "login", err, start)
		return entity.Users{}, "", err
	}

	// Verifikasi password
	if err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		credErr := fmt.Errorf("invalid credentials")
		s.recordError(ctx, span, "login", credErr, start)
		return entity.Users{}, "", credErr
	}

	//  Generate JWT
	claims := jwt.MapClaims{
		"user_id":  user.IDUser,
		"username": user.Username,
		"exp":      time.Now().Add(2 * time.Hour).Unix(),
		"iat":      time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		s.recordError(ctx, span, "login", err, start)
		return entity.Users{}, "", err
	}

	//  Hapus password sebelum return
	user.Password = ""

	span.SetAttributes(attribute.Int("user.id", user.IDUser))
	span.SetStatus(codes.Ok, "")
	slog.InfoContext(ctx, "login success",
		slog.String("username", username),
		slog.Int("user_id", user.IDUser),
		slog.Float64("duration_ms", float64(time.Since(start).Milliseconds())),
	)
	return user, tokenString, nil
}
