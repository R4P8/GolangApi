package routes

import (
	"ApiSimple/internal/handler"
	"ApiSimple/internal/repository"
	"ApiSimple/internal/service"
	"ApiSimple/pkg/middleware"
	"database/sql"

	"github.com/gorilla/mux"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gorilla/mux/otelmux"
)

func RegisterRoutes(router *mux.Router, db *sql.DB) {

	router.Use(otelmux.Middleware("apisimple-service"))

	productRepo := repository.NewProductRepository(db)
	categoryRepo := repository.NewCategoryRepository(db)
	userRepo := repository.NewUserRepository(db)

	productService := service.NewProductService(productRepo)
	categoryService := service.NewCategoryService(categoryRepo)
	userService := service.NewUserService(userRepo)

	productHandler := handler.NewProductHandler(productService)
	categoryHandler := handler.NewCategoryHandler(categoryService)
	userHandler := handler.NewUserHandler(userService)

	jwtMiddleware := middleware.NewJWTMiddleware()

	api := router.PathPrefix("/api").Subrouter()

	api.HandleFunc("/users/register", userHandler.Register).Methods("POST")
	api.HandleFunc("/users/login", userHandler.Login).Methods("POST")

	protected := api.NewRoute().Subrouter()
	protected.Use(jwtMiddleware.Middleware)

	// PRODUCTS
	protected.HandleFunc("/products", productHandler.Create).Methods("POST")
	protected.HandleFunc("/products", productHandler.GetAll).Methods("GET")
	protected.HandleFunc("/products/{id}", productHandler.GetByID).Methods("GET")
	protected.HandleFunc("/products/{id}", productHandler.Update).Methods("PUT")
	protected.HandleFunc("/products/{id}", productHandler.Delete).Methods("DELETE")

	// CATEGORIES
	protected.HandleFunc("/categories", categoryHandler.Create).Methods("POST")
	protected.HandleFunc("/categories", categoryHandler.GetAll).Methods("GET")
	protected.HandleFunc("/categories/{id}", categoryHandler.GetByID).Methods("GET")
	protected.HandleFunc("/categories/{id}", categoryHandler.Update).Methods("PUT")
	protected.HandleFunc("/categories/{id}", categoryHandler.Delete).Methods("DELETE")
}
