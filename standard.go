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
	"golang.org/x/mod/modfile"
)

// DefineTasks defines common tasks for Go projects.
func DefineTasks(opts ...Option) {
	// Override the goyek verbosity default to true since it's generally better.
	// -v=false can still be used to disable it.
	_ = flag.Lookup("v").Value.Set("true")

	command := flag.String("cmd", "", "Command to execute with runall.")

	conf := config{
		artifactsPath: "out",
	}
	for _, o := range opts {
		o.apply(&conf)
	}

	golangciTargets := []string{"./..."}
	// Uses of go-build will very commonly have a build folder, if it is also a module,
	// then let's automatically run checks on it.
	if fileExists(filepath.Join("build", "go.mod")) {
		golangciTargets = append(golangciTargets, "./build")
	}

	root, target := pathRelativeToRoot()

	if !conf.excluded("format-go") {
		RegisterFormatTask(goyek.Define(goyek.Task{
			Name:     "format-go",
			Usage:    "Formats Go code.",
			Parallel: true,
			Action: func(a *goyek.A) {
				cmd.Exec(a, "go tool golangci-lint fmt "+strings.Join(golangciTargets, " "))
				cmd.Exec(a, "go mod tidy")
			},
		}))
	}

	if !conf.excluded("lint-go") {
		RegisterLintTask(goyek.Define(goyek.Task{
			Name:     "lint-go",
			Usage:    "Lints Go code.",
			Parallel: true,
			Action: func(a *goyek.A) {
				cmd.Exec(a, fmt.Sprintf(`go tool golangci-lint run --build-tags "%s" --timeout=20m %s`, strings.Join(conf.buildTags, ","), strings.Join(golangciTargets, " ")))
				goModTidyDiff(a)
			},
		}))
	}

	if !conf.excluded("format-markdown") {
		RegisterFormatTask(goyek.Define(goyek.Task{
			Name:     "format-markdown",
			Usage:    "Formats Markdown code.",
			Parallel: true,
			Action: func(a *goyek.A) {
				cmd.Exec(a, "go tool prettier --no-error-on-unmatched-pattern --write '**/*.md'")
			},
		}))
	}

	if !conf.excluded("lint-markdown") {
		RegisterLintTask(goyek.Define(goyek.Task{
			Name:     "lint-markdown",
			Usage:    "Lints Markdown code.",
			Parallel: true,
			Action: func(a *goyek.A) {
				cmd.Exec(a, "go tool prettier --no-error-on-unmatched-pattern --check '**/*.md'")
			},
		}))
	}

	if !conf.excluded("format-shell") {
		RegisterFormatTask(goyek.Define(goyek.Task{
			Name:     "format-shell",
			Usage:    "Formats shell-like code, including Dockerfile, ignore, dotenv.",
			Parallel: true,
			Action: func(a *goyek.A) {
				cmd.Exec(a, "go tool prettier --no-error-on-unmatched-pattern --write '**/*.sh' '**/*.bash' '**/Dockerfile' '**/*.dockerfile' '**/.*ignore' '**/.env*'")
			},
		}))
	}

	if !conf.excluded("lint-shell") {
		RegisterLintTask(goyek.Define(goyek.Task{
			Name:     "lint-shell",
			Usage:    "Lints shell-like code, including Dockerfile, ignore, dotenv.",
			Parallel: true,
			Action: func(a *goyek.A) {
				cmd.Exec(a, "go tool prettier --no-error-on-unmatched-pattern --check '**/*.sh' '**/*.bash' '**/Dockerfile' '**/*.dockerfile' '**/.*ignore' '**/.env*'")
			},
		}))
	}

	if !conf.excluded("format-yaml") {
		RegisterFormatTask(goyek.Define(goyek.Task{
			Name:     "format-yaml",
			Usage:    "Formats YAML code.",
			Parallel: true,
			Action: func(a *goyek.A) {
				cmd.Exec(a, "go tool prettier --no-error-on-unmatched-pattern --write '**/*.yaml' '**/*.yml'")
			},
		}))
	}

	if !conf.excluded("lint-yaml") {
		RegisterLintTask(goyek.Define(goyek.Task{
			Name:     "lint-yaml",
			Usage:    "Lints YAML code.",
			Parallel: true,
			Action: func(a *goyek.A) {
				cmd.Exec(a, "go tool prettier --no-error-on-unmatched-pattern --check '**/*.yaml' '**/*.yml'")

				if root == "" {
					cmd.Exec(a, "go tool yamllint .")
				} else {
					cmd.Exec(a, "go tool yamllint "+target, cmd.Dir(root))
				}
			},
		}))
	}

	if !conf.excluded("test-go") {
		RegisterTestTask(goyek.Define(goyek.Task{
			Name:  "test-go",
			Usage: "Runs Go unit tests.",
			Action: func(a *goyek.A) {
				if err := os.MkdirAll(conf.artifactsPath, 0o755); err != nil { //nolint:gosec // common for build artifacts
					a.Errorf("failed to create out directory: %v", err)
					return
				}
				cmd.Exec(a, fmt.Sprintf("go test -coverprofile=%s -covermode=atomic -v -timeout=20m ./...", filepath.Join(conf.artifactsPath, "coverage.txt")))
			},
		}))
	}

	if !conf.excluded("runall") && fileExists("go.work") {
		RegisterGenerateTask(goyek.Define(goyek.Task{
			Name:  "runall",
			Usage: "Runs a command in each module in the workspace.",
			Action: func(a *goyek.A) {
				if *command == "" {
					a.Error("missing -cmd flag required for runall")
					return
				}
				content, err := os.ReadFile("go.work")
				if err != nil {
					a.Errorf("failed to read go.work: %v", err)
					return
				}
				wf, err := modfile.ParseWork("go.work", content, nil)
				if err != nil {
					a.Errorf("failed to parse go.work: %v", err)
					return
				}
				for _, u := range wf.Use {
					cmd.Exec(a, *command, cmd.Dir(filepath.Join(".", u.Path)))
				}
			},
		}))
	}

	if !conf.excluded("lint-github") && fileExists(".github") {
		RegisterLintTask(goyek.Define(goyek.Task{
			Name:     "lint-github",
			Usage:    "Lints GitHub Actions workflows.",
			Parallel: true,
			Action: func(a *goyek.A) {
				cmd.Exec(a, `go tool actionlint -shellcheck="go tool shellcheck"`)
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
	buildTags     []string
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
		if e := fileExists(filepath.Join(dir, f)); e {
			return true
		}
	}
	return false
}

func fileExists(p string) bool {
	if _, err := os.Stat(p); err == nil {
		return true
	}
	return false
}

// Tags returns an Option to add build tags to Go lint tasks. If any code is guarded by a build tag
// from default compilation, it should be added here to ensure it is linted.
func Tags(tags ...string) Option {
	return buildTags{tags: tags}
}

type buildTags struct {
	tags []string
}

func (b buildTags) apply(c *config) {
	c.buildTags = append(c.buildTags, b.tags...)
}
