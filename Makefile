.PHONY: module crud --crud fmt

MODULE := $(shell awk '$$1 == "module" { print $$2; exit }' go.mod)

# fmt sorts imports (std → internal → external) and formats all Go files.
fmt:
	gci write -s standard -s "prefix($(MODULE))" -s default --custom-order \
		. 2>/dev/null; gofmt -w .

# Usage:
#   make module name=book
#   make module name=book crud with=cache
#   make module name=book crud=true with=cache
#   make module name=book -- --crud --with=cache
# GNU Make requires `--` before custom long options.
module:
 	# make module name=sample -- --with=cache --crud
	@./scripts/create_module.sh "$(name)" \
		"$(if $(filter crud --crud,$(MAKECMDGOALS)),crud,crud=$(crud))" \
		"with=$(if $(with),$(with),$(--with))"

# Allows CRUD to be used as a flag-like make goal.
crud --crud:
	@:
