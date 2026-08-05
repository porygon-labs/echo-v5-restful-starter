#!/usr/bin/env sh
set -eu

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"

usage() {
  cat <<'EOF'
Usage: ./scripts/create_module.sh <module_name> [--crud] [--with=cache|migrations]

Options:
  crud, --crud              Generate repository, service, and HTTP handler CRUD methods.
  with=cache, --with=cache  Add a go-redis v9 cache-aside repository using generated keys.
  with=migrations           Generate a goose migration SQL file for the module table.
  with=redis,migrations     Combine multiple --with options (comma-separated).

Boolean forms such as crud=true and crud=false are also supported.
EOF
}

fail() {
  echo "Error: $*" >&2
  exit 1
}

# ─── parse arguments ─────────────────────────────────────────────────────────
NAME=${1:-}

if [ -z "$NAME" ]; then
  usage >&2
  exit 1
fi
shift

CRUD=false
CACHE=false
MIGRATIONS=false

for OPTION in "$@"; do
  case "$OPTION" in
  crud | --crud | crud=true | --crud=true | crud=1 | --crud=1 | crud=yes | --crud=yes)
    CRUD=true ;;
  crud= | --crud= | crud=false | --crud=false | crud=0 | --crud=0 | crud=no | --crud=no)
    CRUD=false ;;
  with=cache | --with=cache)
    CACHE=true ;;
  with=migrations | --with=migrations)
    MIGRATIONS=true ;;
  with= | --with= | with=none | --with=none)
    CACHE=false ;;
  *)
    # Allow comma-separated --with= values: --with=redis,migrations
    _with_val="${OPTION#--with=}"
    _with_val="${_with_val#with=}"
    if [ "$_with_val" != "$OPTION" ]; then
      _valid=false
      IFS=','
      for _part in $_with_val; do
        case "$_part" in
        cache | redis)  CACHE=true; _valid=true ;;
        migrations)     MIGRATIONS=true; _valid=true ;;
        esac
      done
      IFS=' '
      if [ "$_valid" = true ]; then
        continue
      fi
    fi
    fail "Unknown option '${OPTION}'. Expected '--crud' or '--with=cache|migrations'." ;;
  esac
done

# ─── validate name ───────────────────────────────────────────────────────────
if ! printf '%s\n' "$NAME" | grep -Eq '^[a-z][a-z0-9_]*$'; then
  fail "Module name must be a lowercase Go identifier (for example: book or book_review)."
fi

case "$NAME" in
break|default|func|interface|select|case|defer|go|map|struct|chan|else|goto|package|switch|const|fallthrough|if|range|type|continue|for|import|return|var)
  fail "Module name '${NAME}' is a Go keyword." ;;
esac

DIR="internal/modules/${NAME}"

if [ -d "$DIR" ]; then
  fail "Module directory '${DIR}' already exists! Operation aborted."
fi

if [ "$CACHE" = true ] && [ ! -f "internal/constants/cache.go" ]; then
  fail "Cache constants are missing at internal/constants/cache.go."
fi

# ─── derive names ────────────────────────────────────────────────────────────
NAME_CAP=$(printf '%s\n' "$NAME" | awk -F_ '{
  for (i = 1; i <= NF; i++) {
    printf "%s%s", toupper(substr($i, 1, 1)), substr($i, 2)
  }
  print ""
}')

MODULE_PATH=$(awk '$1 == "module" { print $2; exit }' go.mod 2>/dev/null || true)
if [ -z "$MODULE_PATH" ]; then
  fail "Could not read the module path from go.mod."
fi

REPO_DIR="${DIR}/repository"
SVC_DIR="${DIR}/service"

# ─── generate ────────────────────────────────────────────────────────────────
mkdir -p "$DIR"
trap 'rm -rf "$DIR"' 0 1 2 3 15

# shellcheck source=lib/templates.sh
. "${SCRIPT_DIR}/lib/templates.sh"
render_all

trap - 0 1 2 3 15

# ─── report ──────────────────────────────────────────────────────────────────
FEATURES="scaffolding"
if [ "$CRUD" = true ]; then
  FEATURES="${FEATURES}, CRUD"
fi
if [ "$CACHE" = true ]; then
  FEATURES="${FEATURES}, Redis cache"
fi
if [ "$MIGRATIONS" = true ]; then
  FEATURES="${FEATURES}, Migrations"
fi

repo_ctor="${NAME}repo.New(deps.DB)"
if [ "$CACHE" = true ]; then
  repo_ctor="${NAME}repo.New(deps.DB, deps.Redis)"
fi

echo "Successfully generated Echo v5 module (${FEATURES}): ${DIR}"
echo
echo "  Don't forget to register the handler in internal/provider/modules.go:"
echo
echo "    import ("
echo "        \"${MODULE_PATH}/internal/modules/${NAME}\""
echo "        ${NAME}repo \"${MODULE_PATH}/internal/modules/${NAME}/repository\""
echo "        ${NAME}svc \"${MODULE_PATH}/internal/modules/${NAME}/service\""
echo "    )"
echo
echo "    // inside RegisterRoutes:"
echo "    ${NAME}Repo := ${repo_ctor}"
echo "    ${NAME}Svc := ${NAME}svc.New(${NAME}Repo)"
echo "    ${NAME}Handler := ${NAME}.NewHandler(${NAME}Svc)"
echo "    ${NAME}Handler.RegisterRoutes(v1.Group(\"/${NAME}\"))"
