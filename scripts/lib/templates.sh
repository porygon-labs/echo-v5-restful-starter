#!/usr/bin/env sh
# ─────────────────────────────────────────────────────────────────────────────
# templates.sh — render_* functions for create_module.sh
#
# Expects these variables to be set by the caller:
#   $NAME        — module name (e.g. "book")
#   $NAME_CAP    — PascalCase module name (e.g. "Book")
#   $MODULE_PATH — Go module path from go.mod
#   $DIR         — target module directory
#   $REPO_DIR    — $DIR/repository
#   $SVC_DIR     — $DIR/service
#   $CRUD        — "true" or "false"
#   $CACHE       — "true" or "false"
# ─────────────────────────────────────────────────────────────────────────────

# ─── model ───────────────────────────────────────────────────────────────────
render_model() {
  cat <<EOF >"$DIR/model.go"
// Package ${NAME} defines database models and entities.
package ${NAME}

import "gorm.io/gorm"

// ${NAME_CAP} represents the database entity for ${NAME}.
type ${NAME_CAP} struct {
	gorm.Model
	// TODO: Define custom GORM struct fields here.
}
EOF
}

# ─── dto ─────────────────────────────────────────────────────────────────────
render_dto() {
  if [ "$CRUD" = true ]; then
    cat <<EOF >"$DIR/dto.go"
// Package ${NAME} defines Data Transfer Objects (DTOs) for incoming requests and outgoing responses.
package ${NAME}

import "time"

// Create${NAME_CAP}Request defines the request payload for creating a ${NAME}.
type Create${NAME_CAP}Request struct {
	// TODO: Define request payload fields with validation tags.
}

// Update${NAME_CAP}Request defines the request payload for updating a ${NAME}.
type Update${NAME_CAP}Request struct {
	// TODO: Define request payload fields with validation tags.
}

// ${NAME_CAP}Response defines the response payload returned to the API client.
type ${NAME_CAP}Response struct {
	ID        uint      \`json:"id"\`
	CreatedAt time.Time \`json:"created_at"\`
	UpdatedAt time.Time \`json:"updated_at"\`
	// TODO: Add custom response fields.
}
EOF
  else
    cat <<EOF >"$DIR/dto.go"
// Package ${NAME} defines Data Transfer Objects (DTOs) for incoming requests and outgoing responses.
package ${NAME}

// Create${NAME_CAP}Request defines the request payload for creating a ${NAME}.
type Create${NAME_CAP}Request struct {
	// TODO: Define request payload fields with validation tags.
}

// ${NAME_CAP}Response defines the response payload returned to the API client.
type ${NAME_CAP}Response struct {
	// TODO: Define response payload fields.
}
EOF
  fi
}

# ─── repository interface (parent package) ───────────────────────────────────
render_repository_interface() {
  if [ "$CRUD" = true ]; then
    cat <<EOF >"$DIR/repository.go"
// Package ${NAME} handles database access and raw storage logic (DAO / Repository layer).
package ${NAME}

import (
	"context"
	"errors"
)

// ErrNotFound is returned when a ${NAME} does not exist.
var ErrNotFound = errors.New("${NAME} not found")

// Repository defines data access operations for ${NAME}.
type Repository interface {
	Create(ctx context.Context, entity *${NAME_CAP}) error
	FindAll(ctx context.Context) ([]${NAME_CAP}, error)
	FindByID(ctx context.Context, id uint) (*${NAME_CAP}, error)
	Update(ctx context.Context, entity *${NAME_CAP}) error
	Delete(ctx context.Context, id uint) error
}
EOF
  else
    cat <<EOF >"$DIR/repository.go"
// Package ${NAME} handles database access and raw storage logic (DAO / Repository layer).
package ${NAME}

// Repository defines data access operations for ${NAME}.
type Repository interface {
	// TODO: Define repository methods (for example: Create, FindByID, Update, Delete).
}
EOF
  fi
}

# ─── repository/ implementation ──────────────────────────────────────────────
render_repository_impl() {
  mkdir -p "$REPO_DIR"

  # repository/repository.go
  if [ "$CACHE" = true ]; then
    cat <<EOF >"$REPO_DIR/repository.go"
// Package repository implements the ${NAME}.Repository interface using GORM and Redis.
package repository

import (
	jsoniter "github.com/json-iterator/go"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"${MODULE_PATH}/internal/modules/${NAME}"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

// repository is the concrete implementation of ${NAME}.Repository.
type repository struct {
	db    *gorm.DB
	cache redis.Cmdable
}

// New initializes a new ${NAME} Repository with Redis caching.
func New(db *gorm.DB, cache redis.Cmdable) ${NAME}.Repository {
	return &repository{
		db:    db,
		cache: cache,
	}
}
EOF
  else
    cat <<EOF >"$REPO_DIR/repository.go"
// Package repository implements the ${NAME}.Repository interface using GORM.
package repository

import (
	jsoniter "github.com/json-iterator/go"
	"gorm.io/gorm"

	"${MODULE_PATH}/internal/modules/${NAME}"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

// repository is the concrete implementation of ${NAME}.Repository.
type repository struct {
	db *gorm.DB
}

// New initializes a new ${NAME} Repository.
func New(db *gorm.DB) ${NAME}.Repository {
	return &repository{
		db: db,
	}
}
EOF
  fi
}

render_repository_create() {
  [ "$CRUD" != true ] && return
  cat <<EOF >"$REPO_DIR/create.go"
package repository

import (
	"context"
	"fmt"

	"${MODULE_PATH}/internal/modules/${NAME}"
)

// Create persists a new ${NAME}.
func (r *repository) Create(ctx context.Context, entity *${NAME}.${NAME_CAP}) error {
	if err := r.db.WithContext(ctx).Create(entity).Error; err != nil {
		return fmt.Errorf("create ${NAME}: %w", err)
	}
EOF
  if [ "$CACHE" = true ]; then
    cat <<'EOF' >>"$REPO_DIR/create.go"

	r.setCached(ctx, entity)
	r.invalidateAllCache(ctx)
EOF
  fi
  cat <<EOF >>"$REPO_DIR/create.go"
	return nil
}
EOF
}

render_repository_find_all() {
  [ "$CRUD" != true ] && return
  cat <<EOF >"$REPO_DIR/find_all.go"
package repository

import (
	"context"
	"fmt"

	"${MODULE_PATH}/internal/modules/${NAME}"
)

// FindAll returns all ${NAME} records.
func (r *repository) FindAll(ctx context.Context) ([]${NAME}.${NAME_CAP}, error) {
EOF
  if [ "$CACHE" = true ]; then
    cat <<'EOF' >>"$REPO_DIR/find_all.go"
	if entities, ok := r.getAllCached(ctx); ok {
		return entities, nil
	}

EOF
  fi
  cat <<EOF >>"$REPO_DIR/find_all.go"
	var entities []${NAME}.${NAME_CAP}
	if err := r.db.WithContext(ctx).Find(&entities).Error; err != nil {
		return nil, fmt.Errorf("find all ${NAME}: %w", err)
	}
EOF
  if [ "$CACHE" = true ]; then
    cat <<'EOF' >>"$REPO_DIR/find_all.go"

	r.setAllCached(ctx, entities)
EOF
  fi
  cat <<EOF >>"$REPO_DIR/find_all.go"

	return entities, nil
}
EOF
}

render_repository_find_by_id() {
  [ "$CRUD" != true ] && return
  cat <<EOF >"$REPO_DIR/find_by_id.go"
package repository

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"${MODULE_PATH}/internal/modules/${NAME}"
)

// FindByID returns one ${NAME} by its primary key.
func (r *repository) FindByID(ctx context.Context, id uint) (*${NAME}.${NAME_CAP}, error) {
EOF
  if [ "$CACHE" = true ]; then
    cat <<'EOF' >>"$REPO_DIR/find_by_id.go"
	if entity, ok := r.getCached(ctx, id); ok {
		return entity, nil
	}

EOF
  fi
  cat <<EOF >>"$REPO_DIR/find_by_id.go"
	var entity ${NAME}.${NAME_CAP}
	if err := r.db.WithContext(ctx).First(&entity, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ${NAME}.ErrNotFound
		}
		return nil, fmt.Errorf("find ${NAME} by id: %w", err)
	}
EOF
  if [ "$CACHE" = true ]; then
    cat <<'EOF' >>"$REPO_DIR/find_by_id.go"

	r.setCached(ctx, &entity)
EOF
  fi
  cat <<EOF >>"$REPO_DIR/find_by_id.go"
	return &entity, nil
}
EOF
}

render_repository_update() {
  [ "$CRUD" != true ] && return
  cat <<EOF >"$REPO_DIR/update.go"
package repository

import (
	"context"
	"fmt"

	"${MODULE_PATH}/internal/modules/${NAME}"
)

// Update persists changes to an existing ${NAME}.
func (r *repository) Update(ctx context.Context, entity *${NAME}.${NAME_CAP}) error {
	if err := r.db.WithContext(ctx).Save(entity).Error; err != nil {
		return fmt.Errorf("update ${NAME}: %w", err)
	}
EOF
  if [ "$CACHE" = true ]; then
    cat <<'EOF' >>"$REPO_DIR/update.go"

	r.setCached(ctx, entity)
	r.invalidateAllCache(ctx)
EOF
  fi
  cat <<EOF >>"$REPO_DIR/update.go"
	return nil
}
EOF
}

render_repository_delete() {
  [ "$CRUD" != true ] && return
  cat <<EOF >"$REPO_DIR/delete.go"
package repository

import (
	"context"
	"fmt"

	"${MODULE_PATH}/internal/modules/${NAME}"
)

// Delete removes a ${NAME} by its primary key.
func (r *repository) Delete(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Delete(&${NAME}.${NAME_CAP}{}, id)
	if result.Error != nil {
		return fmt.Errorf("delete ${NAME}: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ${NAME}.ErrNotFound
	}
EOF
  if [ "$CACHE" = true ]; then
    cat <<'EOF' >>"$REPO_DIR/delete.go"

	r.deleteCached(ctx, id)
	r.invalidateAllCache(ctx)
EOF
  fi
  cat <<'EOF' >>"$REPO_DIR/delete.go"
	return nil
}
EOF
}

render_repository_cache() {
  [ "$CACHE" != true ] && return
  cat <<EOF >"$REPO_DIR/cache.go"
package repository

import (
	"context"
	"fmt"

	"${MODULE_PATH}/internal/constants"
	"${MODULE_PATH}/internal/modules/${NAME}"
)

// cacheKey generates a stable, module-scoped key for a single ${NAME} record.
// Example: go-restful-api:${NAME}:1.
func cacheKey(id uint) string {
	return fmt.Sprintf("%s:${NAME}:%d", constants.CACHE_PREFIX, id)
}

// cacheKeyAll returns the collection-level cache key for FindAll.
func cacheKeyAll() string {
	return fmt.Sprintf("%s:${NAME}:all", constants.CACHE_PREFIX)
}

// Cache failures are intentionally best-effort so Redis outages do not make
// the database-backed repository unavailable.
func (r *repository) getCached(ctx context.Context, id uint) (*${NAME}.${NAME_CAP}, bool) {
	if r.cache == nil {
		return nil, false
	}

	payload, err := r.cache.Get(ctx, cacheKey(id)).Bytes()
	if err != nil {
		return nil, false
	}

	var entity ${NAME}.${NAME_CAP}
	if err := json.Unmarshal(payload, &entity); err != nil {
		r.deleteCached(ctx, id)
		return nil, false
	}

	return &entity, true
}

func (r *repository) setCached(ctx context.Context, entity *${NAME}.${NAME_CAP}) {
	if r.cache == nil || entity == nil {
		return
	}

	payload, err := json.Marshal(entity)
	if err != nil {
		return
	}

	_ = r.cache.Set(ctx, cacheKey(entity.ID), payload, constants.CACHE_DEFAULT_TIMEOUT_MINS).Err()
}

func (r *repository) deleteCached(ctx context.Context, id uint) {
	if r.cache == nil {
		return
	}

	_ = r.cache.Del(ctx, cacheKey(id)).Err()
}

func (r *repository) getAllCached(ctx context.Context) ([]${NAME}.${NAME_CAP}, bool) {
	if r.cache == nil {
		return nil, false
	}

	payload, err := r.cache.Get(ctx, cacheKeyAll()).Bytes()
	if err != nil {
		return nil, false
	}

	var entities []${NAME}.${NAME_CAP}
	if err := json.Unmarshal(payload, &entities); err != nil {
		r.invalidateAllCache(ctx)
		return nil, false
	}

	return entities, true
}

func (r *repository) setAllCached(ctx context.Context, entities []${NAME}.${NAME_CAP}) {
	if r.cache == nil || entities == nil {
		return
	}

	payload, err := json.Marshal(entities)
	if err != nil {
		return
	}

	_ = r.cache.Set(ctx, cacheKeyAll(), payload, constants.CACHE_DEFAULT_TIMEOUT_MINS).Err()
}

func (r *repository) invalidateAllCache(ctx context.Context) {
	if r.cache == nil {
		return
	}

	_ = r.cache.Del(ctx, cacheKeyAll()).Err()
}
EOF
}

# ─── service interface (parent package) ──────────────────────────────────────
render_service_interface() {
  if [ "$CRUD" = true ]; then
    cat <<EOF >"$DIR/service.go"
// Package ${NAME} contains core business logic, validations, and orchestration.
package ${NAME}

import "context"

// Service defines business logic contracts for ${NAME}.
type Service interface {
	Create(ctx context.Context, request Create${NAME_CAP}Request) (${NAME_CAP}Response, error)
	List(ctx context.Context) ([]${NAME_CAP}Response, error)
	GetByID(ctx context.Context, id uint) (${NAME_CAP}Response, error)
	Update(ctx context.Context, id uint, request Update${NAME_CAP}Request) (${NAME_CAP}Response, error)
	Delete(ctx context.Context, id uint) error
}
EOF
  else
    cat <<EOF >"$DIR/service.go"
// Package ${NAME} contains core business logic, validations, and orchestration.
package ${NAME}

// Service defines business logic contracts for ${NAME}.
type Service interface {
	// TODO: Define service layer methods.
}
EOF
  fi
}

# ─── service/ implementation ─────────────────────────────────────────────────
render_service_impl() {
  mkdir -p "$SVC_DIR"

  # service/service.go
  cat <<EOF >"$SVC_DIR/service.go"
// Package service implements the ${NAME}.Service interface.
package service

import (
	jsoniter "github.com/json-iterator/go"

	"${MODULE_PATH}/internal/modules/${NAME}"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

// service is the concrete implementation of ${NAME}.Service.
type service struct {
	repo ${NAME}.Repository
}

// New initializes a new ${NAME} Service.
func New(repo ${NAME}.Repository) ${NAME}.Service {
	return &service{
		repo: repo,
	}
}
EOF
}

render_service_create() {
  [ "$CRUD" != true ] && return
  cat <<EOF >"$SVC_DIR/create.go"
package service

import (
	"context"

	"${MODULE_PATH}/internal/modules/${NAME}"
)

// Create creates a ${NAME}.
func (s *service) Create(ctx context.Context, request ${NAME}.Create${NAME_CAP}Request) (${NAME}.${NAME_CAP}Response, error) {
	entity := new${NAME_CAP}(request)
	if err := s.repo.Create(ctx, entity); err != nil {
		return ${NAME}.${NAME_CAP}Response{}, err
	}

	return to${NAME_CAP}Response(entity), nil
}

func new${NAME_CAP}(request ${NAME}.Create${NAME_CAP}Request) *${NAME}.${NAME_CAP} {
	// TODO: Map request fields to the model.
	_ = request
	return &${NAME}.${NAME_CAP}{}
}
EOF
}

render_service_list() {
  [ "$CRUD" != true ] && return
  cat <<EOF >"$SVC_DIR/list.go"
package service

import (
	"context"

	"${MODULE_PATH}/internal/modules/${NAME}"
)

// List returns all ${NAME} records.
func (s *service) List(ctx context.Context) ([]${NAME}.${NAME_CAP}Response, error) {
	entities, err := s.repo.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	response := make([]${NAME}.${NAME_CAP}Response, len(entities))
	for i := range entities {
		response[i] = to${NAME_CAP}Response(&entities[i])
	}

	return response, nil
}
EOF
}

render_service_get_by_id() {
  [ "$CRUD" != true ] && return
  cat <<EOF >"$SVC_DIR/get_by_id.go"
package service

import (
	"context"

	"${MODULE_PATH}/internal/modules/${NAME}"
)

// GetByID returns one ${NAME}.
func (s *service) GetByID(ctx context.Context, id uint) (${NAME}.${NAME_CAP}Response, error) {
	entity, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return ${NAME}.${NAME_CAP}Response{}, err
	}

	return to${NAME_CAP}Response(entity), nil
}
EOF
}

render_service_update() {
  [ "$CRUD" != true ] && return
  cat <<EOF >"$SVC_DIR/update.go"
package service

import (
	"context"

	"${MODULE_PATH}/internal/modules/${NAME}"
)

// Update updates one ${NAME}.
func (s *service) Update(ctx context.Context, id uint, request ${NAME}.Update${NAME_CAP}Request) (${NAME}.${NAME_CAP}Response, error) {
	entity, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return ${NAME}.${NAME_CAP}Response{}, err
	}

	apply${NAME_CAP}Update(entity, request)
	if err := s.repo.Update(ctx, entity); err != nil {
		return ${NAME}.${NAME_CAP}Response{}, err
	}

	return to${NAME_CAP}Response(entity), nil
}

func apply${NAME_CAP}Update(entity *${NAME}.${NAME_CAP}, request ${NAME}.Update${NAME_CAP}Request) {
	// TODO: Apply request fields to the model.
	_, _ = entity, request
}
EOF
}

render_service_delete() {
  [ "$CRUD" != true ] && return
  cat <<EOF >"$SVC_DIR/delete.go"
package service

import "context"

// Delete deletes one ${NAME}.
func (s *service) Delete(ctx context.Context, id uint) error {
	return s.repo.Delete(ctx, id)
}
EOF
}

render_service_mapping() {
  [ "$CRUD" != true ] && return
  cat <<EOF >"$SVC_DIR/mapping.go"
package service

import (
	"${MODULE_PATH}/internal/modules/${NAME}"
)

// to${NAME_CAP}Response maps a model entity to a response DTO.
func to${NAME_CAP}Response(entity *${NAME}.${NAME_CAP}) ${NAME}.${NAME_CAP}Response {
	return ${NAME}.${NAME_CAP}Response{
		ID:        entity.ID,
		CreatedAt: entity.CreatedAt,
		UpdatedAt: entity.UpdatedAt,
	}
}
EOF
}

# ─── handler ─────────────────────────────────────────────────────────────────
render_handler() {
  if [ "$CRUD" = true ]; then
    cat <<EOF >"$DIR/handler.go"
// Package ${NAME} handles HTTP routing, request parsing, and HTTP response delivery using Echo v5.
package ${NAME}

import (
	"errors"
	"net/http"
	"strconv"

	jsoniter "github.com/json-iterator/go"
	"github.com/labstack/echo/v5"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

// Handler manages HTTP endpoints for ${NAME}.
type Handler struct {
	service Service
}

// NewHandler initializes a new ${NAME} HTTP Handler.
func NewHandler(service Service) *Handler {
	return &Handler{
		service: service,
	}
}

// RegisterRoutes registers CRUD endpoints into an Echo router group.
func (h *Handler) RegisterRoutes(g *echo.Group) {
	g.POST("", h.Create)
	g.GET("", h.List)
	g.GET("/:id", h.GetByID)
	g.PUT("/:id", h.Update)
	g.DELETE("/:id", h.Delete)
}

// Create handles POST requests.
func (h *Handler) Create(c *echo.Context) error {
	var request Create${NAME_CAP}Request
	if err := c.Bind(&request); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body").Wrap(err)
	}

	response, err := h.service.Create(c.Request().Context(), request)
	if err != nil {
		return ${NAME}HTTPError(err)
	}

	return c.JSON(http.StatusCreated, response)
}

// List handles GET collection requests.
func (h *Handler) List(c *echo.Context) error {
	response, err := h.service.List(c.Request().Context())
	if err != nil {
		return ${NAME}HTTPError(err)
	}

	return c.JSON(http.StatusOK, response)
}

// GetByID handles GET item requests.
func (h *Handler) GetByID(c *echo.Context) error {
	id, err := parse${NAME_CAP}ID(c.Param("id"))
	if err != nil {
		return err
	}

	response, err := h.service.GetByID(c.Request().Context(), id)
	if err != nil {
		return ${NAME}HTTPError(err)
	}

	return c.JSON(http.StatusOK, response)
}

// Update handles PUT requests.
func (h *Handler) Update(c *echo.Context) error {
	id, err := parse${NAME_CAP}ID(c.Param("id"))
	if err != nil {
		return err
	}

	var request Update${NAME_CAP}Request
	if err := c.Bind(&request); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body").Wrap(err)
	}

	response, err := h.service.Update(c.Request().Context(), id, request)
	if err != nil {
		return ${NAME}HTTPError(err)
	}

	return c.JSON(http.StatusOK, response)
}

// Delete handles DELETE requests.
func (h *Handler) Delete(c *echo.Context) error {
	id, err := parse${NAME_CAP}ID(c.Param("id"))
	if err != nil {
		return err
	}

	if err := h.service.Delete(c.Request().Context(), id); err != nil {
		return ${NAME}HTTPError(err)
	}

	return c.NoContent(http.StatusNoContent)
}

func parse${NAME_CAP}ID(value string) (uint, error) {
	id, err := strconv.ParseUint(value, 10, strconv.IntSize)
	if err != nil || id == 0 {
		return 0, echo.NewHTTPError(http.StatusBadRequest, "invalid ${NAME} id")
	}
	return uint(id), nil
}

func ${NAME}HTTPError(err error) error {
	if errors.Is(err, ErrNotFound) {
		return echo.NewHTTPError(http.StatusNotFound, ErrNotFound.Error()).Wrap(err)
	}
	return echo.NewHTTPError(http.StatusInternalServerError, "internal server error").Wrap(err)
}
EOF
  else
    cat <<EOF >"$DIR/handler.go"
// Package ${NAME} handles HTTP routing, request parsing, and HTTP response delivery using Echo v5.
package ${NAME}

import (
	jsoniter "github.com/json-iterator/go"
	"github.com/labstack/echo/v5"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

// Handler manages HTTP endpoints for ${NAME}.
type Handler struct {
	service Service
}

// NewHandler initializes a new ${NAME} HTTP Handler.
func NewHandler(service Service) *Handler {
	return &Handler{
		service: service,
	}
}

// RegisterRoutes registers endpoints into an Echo router group.
func (h *Handler) RegisterRoutes(g *echo.Group) {
	// TODO: Define routes.
	// Note: Extract context inside handlers for OpenTelemetry:
	// ctx := c.Request().Context()
}
EOF
  fi
}

# ─── generate all ────────────────────────────────────────────────────────────
render_all() {
  render_model
  render_dto
  render_repository_interface
  render_repository_impl
  render_repository_create
  render_repository_find_all
  render_repository_find_by_id
  render_repository_update
  render_repository_delete
  render_repository_cache
  render_service_interface
  render_service_impl
  render_service_create
  render_service_list
  render_service_get_by_id
  render_service_update
  render_service_delete
  render_service_mapping
  render_handler

  gofmt -w "$DIR"/*.go "$REPO_DIR"/*.go "$SVC_DIR"/*.go 2>/dev/null || true
}
