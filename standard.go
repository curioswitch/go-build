package build

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/goyek/goyek/v2"
	"github.com/goyek/x/cmd"
)

// DefineTasks defines common tasks for Go projects.
func DefineTasks(opts ...Option) {
	conf := config{
		artifactsPath: "out",
	}
	for _, o := range opts {
		o.apply(&conf)
	}

	var goModules []string
	if err := filepath.WalkDir(".", func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		if d.Name() == "go.mod" {
			dir := "./" + filepath.Dir(path)
			goModules = append(goModules, filepath.ToSlash(dir))
		}

		return nil
	}); err != nil {
		goModules = []string{"."}
	}

	formatGo := goyek.Define(goyek.Task{
		Name:     "format-go",
		Usage:    "Formats Go code.",
		Parallel: true,
		Action: func(a *goyek.A) {
			cmd.Exec(a, fmt.Sprintf("go run github.com/golangci/golangci-lint/cmd/golangci-lint@%s run --fix --timeout=20m %s", verGolangCILint, strings.Join(goModules, " ")))
		},
	})

	lintGo := goyek.Define(goyek.Task{
		Name:     "lint-go",
		Usage:    "Lints Go code.",
		Parallel: true,
		Action: func(a *goyek.A) {
			cmd.Exec(a, fmt.Sprintf("go run github.com/golangci/golangci-lint/cmd/golangci-lint@%s run --timeout=20m %s", verGolangCILint, strings.Join(goModules, " ")))
		},
	})

	formatMarkdown := goyek.Define(goyek.Task{
		Name:     "format-markdown",
		Usage:    "Formats Markdown code.",
		Parallel: true,
		Action: func(a *goyek.A) {
			cmd.Exec(a, fmt.Sprintf("go run github.com/wasilibs/go-prettier/cmd/prettier@%s --no-error-on-unmatched-pattern --write '**/*.md'", verGoPrettier))
		},
	})

	lintMarkdown := goyek.Define(goyek.Task{
		Name:     "lint-markdown",
		Usage:    "Lints Markdown code.",
		Parallel: true,
		Action: func(a *goyek.A) {
			cmd.Exec(a, fmt.Sprintf("go run github.com/wasilibs/go-prettier/cmd/prettier@%s --no-error-on-unmatched-pattern --check '**/*.md'", verGoPrettier))
		},
	})

	formatYaml := goyek.Define(goyek.Task{
		Name:     "format-yaml",
		Usage:    "Formats YAML code.",
		Parallel: true,
		Action: func(a *goyek.A) {
			cmd.Exec(a, fmt.Sprintf("go run github.com/wasilibs/go-prettier/cmd/prettier@%s --no-error-on-unmatched-pattern --write '**/*.yaml' '**/*.yml'", verGoPrettier))
		},
	})

	lintYaml := goyek.Define(goyek.Task{
		Name:     "lint-yaml",
		Usage:    "Lints YAML code.",
		Parallel: true,
		Action: func(a *goyek.A) {
			cmd.Exec(a, fmt.Sprintf("go run github.com/wasilibs/go-prettier/cmd/prettier@%s --no-error-on-unmatched-pattern --check '**/*.yaml' '**/*.yml'", verGoPrettier))
			cmd.Exec(a, fmt.Sprintf("go run github.com/wasilibs/go-yamllint/cmd/yamllint@%s .", verGoYamllint))
		},
	})

	goyek.Define(goyek.Task{
		Name:  "format",
		Usage: "Format code in various languages.",
		Deps:  append(goyek.Deps{formatGo, formatMarkdown, formatYaml}, formatTasks...),
	})

	lint := goyek.Define(goyek.Task{
		Name:  "lint",
		Usage: "Lints code in various languages.",
		Deps:  append(goyek.Deps{lintGo, lintMarkdown, lintYaml}, lintTasks...),
	})

	goyek.Define(goyek.Task{
		Name:  "generate",
		Usage: "Generates code.",
		Deps:  generateTasks,
	})

	test := goyek.Define(goyek.Task{
		Name:  "test",
		Usage: "Runs unit tests.",
		Action: func(a *goyek.A) {
			if err := os.MkdirAll(conf.artifactsPath, 0o755); err != nil {
				a.Errorf("failed to create out directory: %v", err)
				return
			}
			cmd.Exec(a, fmt.Sprintf("go test -coverprofile=%s -covermode=atomic -v -timeout=20m ./...", filepath.Join(conf.artifactsPath, "coverage.txt")))
		},
	})

	goyek.Define(goyek.Task{
		Name:  "check",
		Usage: "Runs all checks.",
		Deps:  goyek.Deps{lint, test},
	})
}

type config struct {
	artifactsPath string
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
