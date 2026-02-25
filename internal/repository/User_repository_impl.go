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
	"golang.org/x/crypto/bcrypt"
)

type UserRepositoryImpl struct {
	DB     *sql.DB
	Tracer trace.Tracer
}

func NewUserRepository(db *sql.DB) UserRepo {
	return &UserRepositoryImpl{
		DB:     db,
		Tracer: otel.Tracer("user-repository"),
	}
}

func (r *UserRepositoryImpl) Register(ctx context.Context, user entity.Users) (entity.Users, error) {
	ctx, span := r.Tracer.Start(ctx, "UserRepository.Register")
	defer span.End()

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return user, err
	}

	query := `
		INSERT INTO users (username, password)
		VALUES ($1, $2)
		RETURNING id_user, created_at, updated_at
	`

	span.SetAttributes(
		attribute.String("db.system", "postgresql"),
		attribute.String("db.operation", "INSERT"),
	)

	err = r.DB.QueryRowContext(
		ctx,
		query,
		user.Username,
		string(hashedPassword),
	).Scan(
		&user.IDUser,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return user, err
	}

	user.Password = ""
	return user, nil
}

func (r *UserRepositoryImpl) GetByUsername(ctx context.Context, username string) (entity.Users, error) {
	ctx, span := r.Tracer.Start(ctx, "UserRepository.GetByUsername")
	defer span.End()

	var user entity.Users

	query := `
		SELECT id_user, username, password, created_at, updated_at
		FROM users
		WHERE username = $1
	`

	span.SetAttributes(
		attribute.String("db.system", "postgresql"),
		attribute.String("db.operation", "SELECT"),
		attribute.String("user.username", username),
	)

	err := r.DB.QueryRowContext(ctx, query, username).Scan(
		&user.IDUser,
		&user.Username,
		&user.Password,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = errors.New("invalid username or password")
		}
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return user, err
	}

	return user, nil
}

func (r *UserRepositoryImpl) Login(ctx context.Context, username string, password string) (entity.Users, error) {
	ctx, span := r.Tracer.Start(ctx, "UserRepository.Login")
	defer span.End()

	user, err := r.GetByUsername(ctx, username)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return user, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		err = errors.New("invalid username or password")
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return entity.Users{}, err
	}

	user.Password = ""
	return user, nil
}
