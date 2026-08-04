package scripts_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestCreateModule(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		wantCRUD  bool
		wantCache bool
	}{
		{name: "scaffold only", args: []string{"note"}},
		{name: "CRUD only", args: []string{"article", "crud"}, wantCRUD: true},
		{name: "cache only", args: []string{"cached", "with=cache"}, wantCache: true},
		{name: "CRUD with cache", args: []string{"book_review", "crud", "with=cache"}, wantCRUD: true, wantCache: true},
		{name: "long options", args: []string{"catalog", "--with=cache", "--crud"}, wantCRUD: true, wantCache: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := newGeneratorFixture(t)
			runGenerator(t, root, tt.args...)

			moduleName := tt.args[0]
			moduleDir := filepath.Join(root, "internal", "modules", moduleName)
			repoDir := filepath.Join(moduleDir, "repository")

			// Interface file at module root
			repositoryInterface := readFile(t, filepath.Join(moduleDir, "repository.go"))
			assertContains(t, repositoryInterface, "Create(ctx context.Context", tt.wantCRUD)

			// Implementation in repository/ sub-package
			repoImpl := readFile(t, filepath.Join(repoDir, "repository.go"))
			assertContains(t, repoImpl, "redis.Cmdable", tt.wantCache)

			// Handler
			handler := readFile(t, filepath.Join(moduleDir, "handler.go"))
			assertContains(t, handler, `g.POST("", h.Create)`, tt.wantCRUD)

			// Cache in repository/ sub-package
			cachePath := filepath.Join(repoDir, "cache.go")
			_, err := os.Stat(cachePath)
			if tt.wantCache {
				if err != nil {
					t.Fatalf("repository/cache.go was not generated: %v", err)
				}
				cache := readFile(t, cachePath)
				assertContains(t, cache, "constants.CACHE_PREFIX", true)
				assertContains(t, cache, "constants.CACHE_DEFAULT_TIMEOUT_MINS", true)
				assertContains(t, cache, `fmt.Sprintf("%s:`+moduleName+`:%d"`, true)
			} else if !os.IsNotExist(err) {
				t.Fatalf("repository/cache.go exists without with=cache (stat error: %v)", err)
			}
		})
	}
}

func TestMakeModuleFlags(t *testing.T) {
	tests := []struct {
		name       string
		moduleName string
		args       []string
	}{
		{
			name:       "make-style flags",
			moduleName: "book",
			args:       []string{"module", "name=book", "crud", "with=cache"},
		},
		{
			name:       "long options after terminator",
			moduleName: "catalog",
			args:       []string{"module", "name=catalog", "--", "--with=cache", "--crud"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := newGeneratorFixture(t)
			cmd := exec.Command("make", tt.args...)
			cmd.Dir = root
			if output, err := cmd.CombinedOutput(); err != nil {
				t.Fatalf("make module failed: %v\n%s", err, output)
			}

			moduleDir := filepath.Join(root, "internal", "modules", tt.moduleName)
			repoDir := filepath.Join(moduleDir, "repository")

			assertContains(t, readFile(t, filepath.Join(repoDir, "repository.go")), "redis.Cmdable", true)
			assertContains(t, readFile(t, filepath.Join(moduleDir, "handler.go")), `g.POST("", h.Create)`, true)
		})
	}
}

func TestCreateModuleRejectsUnknownOption(t *testing.T) {
	root := newGeneratorFixture(t)
	cmd := generatorCommand(t, root, "note", "with=unknown")
	if err := cmd.Run(); err == nil {
		t.Fatal("generator accepted an unknown option")
	}
	if _, err := os.Stat(filepath.Join(root, "internal", "modules", "note")); !os.IsNotExist(err) {
		t.Fatalf("generator left a module behind after rejecting an option: %v", err)
	}
}

func newGeneratorFixture(t *testing.T) string {
	t.Helper()

	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "internal", "constants"), 0o755); err != nil {
		t.Fatalf("create fixture: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(root, "scripts", "lib"), 0o755); err != nil {
		t.Fatalf("create scripts/lib directory: %v", err)
	}

	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("locate generator test")
	}
	scriptsDir := filepath.Dir(filename)
	copyFile(t, filepath.Join(scriptsDir, "create_module.sh"), filepath.Join(root, "scripts", "create_module.sh"), 0o700)
	copyFile(t, filepath.Join(scriptsDir, "lib", "templates.sh"), filepath.Join(root, "scripts", "lib", "templates.sh"), 0o600)
	copyFile(t, filepath.Join(scriptsDir, "..", "Makefile"), filepath.Join(root, "Makefile"), 0o600)

	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module example.com/generator\n\ngo 1.26.5\n"), 0o600); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "internal", "constants", "cache.go"), []byte("package constants\n"), 0o600); err != nil {
		t.Fatalf("write cache constants: %v", err)
	}
	return root
}

func runGenerator(t *testing.T, root string, args ...string) {
	t.Helper()
	output, err := generatorCommand(t, root, args...).CombinedOutput()
	if err != nil {
		t.Fatalf("generator failed: %v\n%s", err, output)
	}
}

func generatorCommand(t *testing.T, root string, args ...string) *exec.Cmd {
	t.Helper()
	cmd := exec.Command("sh", append([]string{filepath.Join(root, "scripts", "create_module.sh")}, args...)...)
	cmd.Dir = root
	return cmd
}

func copyFile(t *testing.T, source, destination string, mode os.FileMode) {
	t.Helper()
	content, err := os.ReadFile(source)
	if err != nil {
		t.Fatalf("read fixture source %s: %v", source, err)
	}
	if err := os.WriteFile(destination, content, mode); err != nil {
		t.Fatalf("write fixture %s: %v", destination, err)
	}
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(content)
}

func assertContains(t *testing.T, content, value string, want bool) {
	t.Helper()
	if got := strings.Contains(content, value); got != want {
		t.Errorf("contains %q = %v, want %v", value, got, want)
	}
}
