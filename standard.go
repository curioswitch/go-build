package build

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/goyek/goyek/v2"
	_ "github.com/goyek/x/boot" // define flags to override
	"github.com/goyek/x/cmd"
)

// DefineTasks defines common tasks for Go projects.
func DefineTasks(opts ...Option) {
	// Override the goyek verbosity default to true since it's generally better.
	// -v=false can still be used to disable it.
	_ = flag.Lookup("v").Value.Set("true")

	conf := config{
		artifactsPath: "out",
	}
	for _, o := range opts {
		o.apply(&conf)
	}

	golangciTargets := []string{"./..."}
	// Uses of go-build will very commonly have a build folder, if it is also a module,
	// then let's automatically run checks on it.
	if _, err := os.Stat(filepath.Join("build", "go.mod")); err == nil {
		golangciTargets = append(golangciTargets, "./build")
	}

	root, target := pathRelativeToRoot()

	if !conf.excluded("format-go") {
		RegisterFormatTask(goyek.Define(goyek.Task{
			Name:     "format-go",
			Usage:    "Formats Go code.",
			Parallel: true,
			Action: func(a *goyek.A) {
				cmd.Exec(a, fmt.Sprintf("go run github.com/golangci/golangci-lint/cmd/golangci-lint@%s run --fix --timeout=20m %s", verGolangCILint, strings.Join(golangciTargets, " ")))
			},
		}))
	}

	if !conf.excluded("lint-go") {
		RegisterLintTask(goyek.Define(goyek.Task{
			Name:     "lint-go",
			Usage:    "Lints Go code.",
			Parallel: true,
			Action: func(a *goyek.A) {
				cmd.Exec(a, fmt.Sprintf("go run github.com/golangci/golangci-lint/cmd/golangci-lint@%s run --timeout=20m %s", verGolangCILint, strings.Join(golangciTargets, " ")))
			},
		}))
	}

	if !conf.excluded("format-markdown") {
		RegisterFormatTask(goyek.Define(goyek.Task{
			Name:     "format-markdown",
			Usage:    "Formats Markdown code.",
			Parallel: true,
			Action: func(a *goyek.A) {
				cmd.Exec(a, fmt.Sprintf("go run github.com/wasilibs/go-prettier/cmd/prettier@%s --no-error-on-unmatched-pattern --write '**/*.md'", verGoPrettier))
			},
		}))
	}

	if !conf.excluded("lint-markdown") {
		RegisterLintTask(goyek.Define(goyek.Task{
			Name:     "lint-markdown",
			Usage:    "Lints Markdown code.",
			Parallel: true,
			Action: func(a *goyek.A) {
				cmd.Exec(a, fmt.Sprintf("go run github.com/wasilibs/go-prettier/cmd/prettier@%s --no-error-on-unmatched-pattern --check '**/*.md'", verGoPrettier))
			},
		}))
	}

	if !conf.excluded("format-shell") {
		RegisterFormatTask(goyek.Define(goyek.Task{
			Name:     "format-shell",
			Usage:    "Formats shell-like code, including Dockerfile, ignore, dotenv.",
			Parallel: true,
			Action: func(a *goyek.A) {
				cmd.Exec(a, fmt.Sprintf("go run github.com/wasilibs/go-prettier/cmd/prettier@%s --no-error-on-unmatched-pattern --write '**/*.sh' '**/*.bash' '**/Dockerfile' '**/*.dockerfile' '**/.*ignore' '**/.env*'", verGoPrettier))
			},
		}))
	}

	if !conf.excluded("lint-shell") {
		RegisterLintTask(goyek.Define(goyek.Task{
			Name:     "lint-shell",
			Usage:    "Lints shell-like code, including Dockerfile, ignore, dotenv.",
			Parallel: true,
			Action: func(a *goyek.A) {
				cmd.Exec(a, fmt.Sprintf("go run github.com/wasilibs/go-prettier/cmd/prettier@%s --no-error-on-unmatched-pattern --check '**/*.sh' '**/*.bash' '**/Dockerfile' '**/*.dockerfile' '**/.*ignore' '**/.env*'", verGoPrettier))
			},
		}))
	}

	if !conf.excluded("format-yaml") {
		RegisterFormatTask(goyek.Define(goyek.Task{
			Name:     "format-yaml",
			Usage:    "Formats YAML code.",
			Parallel: true,
			Action: func(a *goyek.A) {
				cmd.Exec(a, fmt.Sprintf("go run github.com/wasilibs/go-prettier/cmd/prettier@%s --no-error-on-unmatched-pattern --write '**/*.yaml' '**/*.yml'", verGoPrettier))
			},
		}))
	}

	if !conf.excluded("lint-yaml") {
		RegisterLintTask(goyek.Define(goyek.Task{
			Name:     "lint-yaml",
			Usage:    "Lints YAML code.",
			Parallel: true,
			Action: func(a *goyek.A) {
				cmd.Exec(a, fmt.Sprintf("go run github.com/wasilibs/go-prettier/cmd/prettier@%s --no-error-on-unmatched-pattern --check '**/*.yaml' '**/*.yml'", verGoPrettier))

				if root == "" {
					cmd.Exec(a, fmt.Sprintf("go run github.com/wasilibs/go-yamllint/cmd/yamllint@%s .", verGoYamllint))
				} else {
					cmd.Exec(a, fmt.Sprintf("go run github.com/wasilibs/go-yamllint/cmd/yamllint@%s %s", verGoYamllint, target), cmd.Dir(root))
				}
			},
		}))
	}

	if !conf.excluded("test-go") {
		RegisterTestTask(goyek.Define(goyek.Task{
			Name:  "test-go",
			Usage: "Runs Go unit tests.",
			Action: func(a *goyek.A) {
				if err := os.MkdirAll(conf.artifactsPath, 0o755); err != nil {
					a.Errorf("failed to create out directory: %v", err)
					return
				}
				cmd.Exec(a, fmt.Sprintf("go test -coverprofile=%s -covermode=atomic -v -timeout=20m ./...", filepath.Join(conf.artifactsPath, "coverage.txt")))
			},
		}))
	}

	goyek.Define(goyek.Task{
		Name:  "format",
		Usage: "Format code in various languages.",
		Deps:  formatTasks,
	})

	lint := goyek.Define(goyek.Task{
		Name:  "lint",
		Usage: "Lints code in various languages.",
		Deps:  lintTasks,
	})

	goyek.Define(goyek.Task{
		Name:  "generate",
		Usage: "Generates code.",
		Deps:  generateTasks,
	})

	test := goyek.Define(goyek.Task{
		Name:  "test",
		Usage: "Runs tests.",
		Deps:  testTasks,
	})

	goyek.Define(goyek.Task{
		Name:  "check",
		Usage: "Runs all checks.",
		Deps:  goyek.Deps{lint, test},
	})
}

type config struct {
	artifactsPath string
	excludeTasks  []string
}

func (c *config) excluded(task string) bool {
	return slices.Contains(c.excludeTasks, task)
}

// Option is a configuration option for DefineTasks.
type Option interface {
	apply(conf *config)
}

// ArtifactPath returns an Option to indicate the path to write temporary build artifacts to,
// for example coverage reports. If not provided, the default is "out".
func ArtifactsPath(path string) Option {
	return artifactsPath(path)
}

type artifactsPath string

func (a artifactsPath) apply(c *config) {
	c.artifactsPath = string(a)
}

// ExcludeTasks returns an Option to exclude tasks normally added by default. This can be used
// to avoid unneeded tasks, for example to disable linting of Markdown while still keeping the
// ability to manually autoformat it, or to redefine a task with a different implementation.
func ExcludeTasks(task ...string) Option {
	return excludeTasks{tasks: task}
}

type excludeTasks struct {
	tasks []string
}

func (e excludeTasks) apply(c *config) {
	c.excludeTasks = append(c.excludeTasks, e.tasks...)
}

func pathRelativeToRoot() (string, string) {
	dir, err := filepath.Abs(".")
	if err != nil {
		return "", ""
	}
	base := dir
	for {
		if anyFileExists(base, ".git", "go.work") {
			target, _ := filepath.Rel(base, dir)
			return base, target
		}

		parent := filepath.Dir(base)
		if parent == dir || parent == "" {
			break
		}

		base = parent
	}
	return "", ""
}

func anyFileExists(dir string, files ...string) bool {
	for _, f := range files {
		if _, err := os.Stat(filepath.Join(dir, f)); err == nil {
			return true
		}
	}
	return false
}
