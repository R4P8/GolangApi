# API Golang

A simple RESTful API server built with Go (Golang). This project demonstrates clean architecture with layers for handlers, services, repositories, and entities. It includes support for JWT authentication, logging, tracing, and observability.

<!--toc-->

## 🔧 Prerequisites

- [Go 1.20+](https://golang.org/dl/)
- [Docker](https://www.docker.com/) (optional, for containerized setup)
- Git (for cloning)

## 🚀 Getting Started

Clone the repository:

```bash
git clone https://your-repo-url/ApiGolang.git
cd ApiGolang
```

Create a `.env` file or set environment variables for database credentials, JWT secret, etc.

### 🏃 Run Locally

```bash
go run ./cmd
```

The server listens on port `:8080` by default. You can change this in the configuration file.

### 🐳 With Docker

Build the image:

```bash
docker build -t apigolang .
```

Start containers (uses `docker-compose.yaml`):

```bash
docker-compose up --build
```

## 🗂 Project Structure

```text
cmd/                # entrypoint (main.go)
internal/
  config/            # configuration loaders
  entity/            # domain models
  handler/           # HTTP handlers
  repository/        # data access implementations & interfaces
  routes/            # router setup
  service/           # business logic
pkg/
  helpers/           # utility helpers
  logger/            # logging setup
  middleware/        # JWT, etc.
  tracing/           # OpenTelemetry tracer
```

## 📌 Endpoints

Endpoints are defined in `internal/routes/routes.go`. Example:

- `POST /login` – authenticate and return JWT
- `GET /products` – list products
- `POST /categories` – create a category

Refer to each handler for full details.

## 🛠 Features

- Clean architecture separation
- JWT authentication middleware
- Structured logging
- OpenTelemetry tracing
- Dockerized configuration
- Prometheus & Grafana (via `prometheus.yaml`)

## 📚 Resources

- [Go Documentation](https://pkg.go.dev/std)
- [OpenTelemetry Go](https://opentelemetry.io/)

---

Feel free to contribute by opening issues or pull requests!