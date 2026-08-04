#!/usr/bin/env sh

set -e

NAME=$1

if [ -z "$NAME" ]; then
  echo "Error: Please specify module name."
  echo "Usage: ./scripts/create_module.sh <module_name>"
  exit 1
fi

DIR="internal/modules/${NAME}"

# Safety Check: Stop if module directory already exists
if [ -d "$DIR" ]; then
  echo "Error: Module directory '${DIR}' already exists! Operation aborted."
  exit 1
fi

# POSIX-compliant capitalization (e.g., "book" -> "Book")
FIRST_CHAR=$(echo "$NAME" | cut -c1 | tr '[:lower:]' '[:upper:]')
REST_CHARS=$(echo "$NAME" | cut -c2-)
NAME_CAP="${FIRST_CHAR}${REST_CHARS}"

mkdir -p "$DIR"

# 1. model.go
cat <<EOF >"$DIR/model.go"
// Package ${NAME} defines database models and entities.
package ${NAME}

import "gorm.io/gorm"

// ${NAME_CAP} represents the database entity for ${NAME}
type ${NAME_CAP} struct {
	gorm.Model
	// TODO: Define custom GORM struct fields here
}
EOF

# 2. dto.go
cat <<EOF >"$DIR/dto.go"
// Package ${NAME} defines Data Transfer Objects (DTOs) for incoming requests and outgoing responses.
package ${NAME}

// Create${NAME_CAP}Request defines the request payload for creating a ${NAME}
type Create${NAME_CAP}Request struct {
	// TODO: Define request payload fields with validation tags
}

// ${NAME_CAP}Response defines the response payload returned to the API client
type ${NAME_CAP}Response struct {
	// TODO: Define response payload fields
}
EOF

# 3. repository.go
cat <<EOF >"$DIR/repository.go"
// Package ${NAME} handles database access and raw storage logic (DAO / Repository layer).
package ${NAME}

import (
	"gorm.io/gorm"
)

// Repository defines data access operations for ${NAME}
type Repository interface {
	// TODO: Define repository methods (e.g., Create, FindByID, Update, Delete)
}

type repository struct {
	db *gorm.DB
}

// NewRepository initializes a new ${NAME} Repository
func NewRepository(db *gorm.DB) Repository {
	return &repository{
		db: db,
	}
}
EOF

# 4. service.go
cat <<EOF >"$DIR/service.go"
// Package ${NAME} contains core business logic, validations, and orchestration.
package ${NAME}

// Service defines business logic contracts for ${NAME}
type Service interface {
	// TODO: Define service layer methods
}

type service struct {
	repo Repository
}

// NewService initializes a new ${NAME} Service
func NewService(repo Repository) Service {
	return &service{
		repo: repo,
	}
}
EOF

# 5. handler.go
cat <<EOF >"$DIR/handler.go"
// Package ${NAME} handles HTTP routing, request parsing, and HTTP response delivery using Echo v5.
package ${NAME}

import (
	"github.com/labstack/echo/v5"
)

// Handler manages HTTP endpoints for ${NAME}
type Handler struct {
	service Service
}

// NewHandler initializes a new ${NAME} HTTP Handler
func NewHandler(service Service) *Handler {
	return &Handler{
		service: service,
	}
}

// RegisterRoutes registers endpoints into an Echo router group
func (h *Handler) RegisterRoutes(g *echo.Group) {
	// TODO: Define routes
	// Note: Extract context inside handlers for OpenTelemetry:
	// ctx := c.Request().Context()
}
EOF

echo "Successfully generated Echo v5 module scaffolding: ${DIR}"
