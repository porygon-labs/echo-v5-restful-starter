# Go RESTful API

Production-ready RESTful API boilerplate built with Go, Echo v5, GORM, Redis, and OpenTelemetry. Designed for fast iteration with a modular, convention-over-configuration architecture.

## Stack

| Concern          | Library                                                              |
| ---------------- | -------------------------------------------------------------------- |
| HTTP router      | [Echo v5](https://github.com/labstack/echo)                          |
| ORM              | [GORM](https://gorm.io) + PostgreSQL                                 |
| Caching          | [go-redis](https://github.com/redis/go-redis)                        |
| Telemetry        | [OpenTelemetry](https://opentelemetry.io) + Jaeger                   |
| Config           | [caarlos0/env](https://github.com/caarlos0/env) + `.env`             |
| ID generation    | [Sqids](https://sqids.org)                                           |
| Serialization    | [json-iterator](https://github.com/json-iterator/go)                 |
| Testing          | stdlib `testing` + [miniredis](https://github.com/alicebob/miniredis) |

## Project Structure

```
.
├── cmd/api/main.go              # Entrypoint
├── internal/
│   ├── config/                  # Env-based configuration
│   ├── constants/               # Shared constants (cache keys, test helpers)
│   ├── modules/
│   │   └── sample/              # Feature modules (each with repository & service)
│   ├── pkg/
│   │   ├── hash/                # Hashing utilities
│   │   ├── response/            # JSON envelope helpers
│   │   └── telemetry/           # OpenTelemetry setup
│   ├── provider/                # DI wiring (DB, Redis, route registration)
│   └── utils/                   # Shared utilities
├── scripts/
│   └── create_module.sh         # Module scaffolding (with optional CRUD + Redis cache)
├── .air.toml                    # Hot-reload config
├── compose.yaml                 # Dev dependencies (Postgres, Redis, Jaeger, Adminer)
├── Makefile                     # Code generation & formatting
└── .github/workflows/ci.yml     # CI pipeline
```

## Getting Started

### Prerequisites

- **Go** 1.26+
- **Docker** (for development dependencies)

### Setup

```bash
# Clone and enter the project
git clone <repo-url> && cd go-restful-api

# Copy environment file
cp .env.example .env

# Start development dependencies (Postgres, Redis, Jaeger, Adminer)
make up

# Run with hot-reload (requires air)
make dev

# Or run directly
make run
```

The server starts at `http://localhost:8080`.

### Environment Variables

| Variable   | Default                              | Required | Description                          |
| ---------- | ------------------------------------ | -------- | ------------------------------------ |
| `APP_ENV`  | `development`                        | No       | Environment label (`development`, `production`) |
| `APP_HOST` | `0.0.0.0`                            | No       | HTTP listen address                  |
| `PORT`     | `8080`                               | No       | HTTP listen port                     |
| `DB_DSN`   | —                                    | **Yes**  | PostgreSQL connection string         |
| `REDIS_URL`| —                                    | **Yes**  | Redis connection URL                 |

### Dev Tools

| URL                    | Tool      |
| ---------------------- | --------- |
| `http://localhost:9090` | Adminer (DB browser) |
| `http://localhost:16686`| Jaeger (tracing UI)  |

## API

### Health

```
GET /healthz          → { "meta": { "is_success": true, "message": "OK" }, "data": { "status": "ok" } }
GET /readyz           → { "meta": { ... }, "data": { "db": "ok", "redis": "ok" } }
```

### Response Envelope

Every endpoint returns a consistent JSON shape:

```json
{
  "meta": { "is_success": true, "message": "OK" },
  "data": { ... }
}
```

## Generating a Module

Scaffold a new feature module complete with repository and service layers:

```bash
# Bare scaffolding
make module name=book

# With CRUD methods
make module name=book -- --crud

# With Redis cache-aside
make module name=book -- --crud --with=cache
```

Then register the module in `internal/provider/modules.go` as shown in the script output.

## Makefile

```bash
make help        # List all available targets
```

| Target         | Description                              |
| -------------- | ---------------------------------------- |
| `make help`    | Show all targets with descriptions       |
| `make build`   | Compile binary to `./tmp/main`           |
| `make run`     | Start the server                         |
| `make dev`     | Start with hot-reload (requires `air`)   |
| `make test`    | Run unit tests with race detection       |
| `make test-cover` | Run tests + coverage report           |
| `make lint`    | Run `golangci-lint`                      |
| `make gosec`   | Run security scan                        |
| `make fmt`     | Sort imports (`gci`) + format (`gofmt`)  |
| `make tidy`    | Tidy module dependencies                 |
| `make clean`   | Remove build artifacts                   |
| `make up`      | Start dev containers (Postgres, Redis, Jaeger, Adminer) |
| `make down`    | Stop dev containers                      |
| `make module`  | Scaffold a new feature module            |

Raw commands are always available if you prefer:

```bash
go test -v -race ./...
go test -v -race -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## CI

The [CI workflow](.github/workflows/ci.yml) runs on every push and PR to `main`/`master`:

| Job    | Description                         | Fails pipeline? |
| ------ | ----------------------------------- | --------------- |
| Lint   | `golangci-lint` static analysis     | No              |
| Gosec  | Security-focused static analysis    | Yes             |
| Test   | Unit tests with race detection      | Yes             |

## License

[MIT](LICENSE)
