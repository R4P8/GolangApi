package middleware

import (
	"context"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

type contextKey string

const UserContextKey contextKey = "user"

type JWTMiddleware struct {
	SecretKey string

	Tracer trace.Tracer
	Meter  metric.Meter

	throughput  metric.Int64Counter
	errorCount  metric.Int64Counter
	latencyHist metric.Float64Histogram
}

func NewJWTMiddleware() *JWTMiddleware {

	tracer := otel.Tracer("jwt-middleware")
	meter := otel.Meter("jwt-middleware")

	throughput, _ := meter.Int64Counter("http_jwt_requests_total")
	errorCount, _ := meter.Int64Counter("http_jwt_errors_total")
	latencyHist, _ := meter.Float64Histogram("http_jwt_latency_ms")

	return &JWTMiddleware{
		SecretKey:   os.Getenv("JWT_SECRET"),
		Tracer:      tracer,
		Meter:       meter,
		throughput:  throughput,
		errorCount:  errorCount,
		latencyHist: latencyHist,
	}
}

func (m *JWTMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		start := time.Now()
		ctx, span := m.Tracer.Start(r.Context(), "JWTMiddleware")
		defer span.End()

		m.throughput.Add(ctx, 1,
			metric.WithAttributes(attribute.String("path", r.URL.Path)),
		)

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			m.handleError(w, ctx, span, "missing authorization header", http.StatusUnauthorized, start)
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			m.handleError(w, ctx, span, "invalid authorization format", http.StatusUnauthorized, start)
			return
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return []byte(m.SecretKey), nil
		})

		if err != nil || !token.Valid {
			m.handleError(w, ctx, span, "invalid or expired token", http.StatusUnauthorized, start)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			m.handleError(w, ctx, span, "invalid token claims", http.StatusUnauthorized, start)
			return
		}

		// Add claims to context
		ctx = context.WithValue(ctx, UserContextKey, claims)

		// Record latency
		duration := float64(time.Since(start).Milliseconds())
		m.latencyHist.Record(ctx, duration,
			metric.WithAttributes(attribute.String("path", r.URL.Path)),
		)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (m *JWTMiddleware) handleError(
	w http.ResponseWriter,
	ctx context.Context,
	span trace.Span,
	message string,
	status int,
	start time.Time,
) {

	log.Printf("JWT ERROR: %s", message)

	span.RecordError(context.Canceled)
	span.SetStatus(codes.Error, message)

	m.errorCount.Add(ctx, 1,
		metric.WithAttributes(attribute.String("error", message)),
	)

	duration := float64(time.Since(start).Milliseconds())
	m.latencyHist.Record(ctx, duration)

	http.Error(w, message, status)
}
