package build

import (
	"bytes"
	"flag"
	"fmt"
	"maps"
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

	if conf.verActionlint == "" {
		conf.verActionlint = verActionlint
	}

	if conf.verGolangCILint == "" {
		conf.verGolangCILint = verGolangCILint
	}

	if conf.verGoPrettier == "" {
		conf.verGoPrettier = verGoPrettier
	}

	if conf.verGoShellcheck == "" {
		conf.verGoShellcheck = verGoShellcheck
	}

	if conf.verGoYamllint == "" {
		conf.verGoYamllint = verGoYamllint
	}

	if conf.verPinact == "" {
		conf.verPinact = verPinact
	}

	if conf.verReviewdog == "" {
		conf.verReviewdog = verReviewdog
	}

	golangciTargets := []string{"./..."}
	// Uses of go-build will very commonly have a build folder, if it is also a module,
	// then let's automatically run checks on it.
	if fileExists(filepath.Join("build", "go.mod")) {
		golangciTargets = append(golangciTargets, "./build")
	}

	root, target := pathRelativeToRoot()

	if !conf.excluded("format-go") {
		RegisterModuleDownloads("github.com/golangci/golangci-lint/v2@" + conf.verGolangCILint)
		RegisterFormatTask(goyek.Define(goyek.Task{
			Name:     "format-go",
			Usage:    "Formats Go code.",
			Parallel: true,
			Action: func(a *goyek.A) {
				cmd.Exec(a, fmt.Sprintf(`go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@%s fmt %s`, conf.verGolangCILint, strings.Join(golangciTargets, " ")))
				cmd.Exec(a, "go mod tidy")
			},
		}))
	}

	if !conf.excluded("lint-go") {
		RegisterModuleDownloads(
			"github.com/golangci/golangci-lint/v2@"+conf.verGolangCILint,
			"github.com/reviewdog/reviewdog@"+conf.verReviewdog,
		)
		RegisterLintTask(goyek.Define(goyek.Task{
			Name:     "lint-go",
			Usage:    "Lints Go code.",
			Parallel: true,
			Action: func(a *goyek.A) {
				execReviewdog(conf, a, "-f=golangci-lint -name=golangci-lint",
					fmt.Sprintf(`go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@%s run --build-tags "%s" --timeout=20m %s`,
						conf.verGolangCILint, strings.Join(conf.buildTags, ","), strings.Join(golangciTargets, " ")))
				cmd.Exec(a, "go mod tidy -diff")
			},
		}))
	}

	if !conf.excluded("format-markdown") {
		RegisterModuleDownloads("github.com/wasilibs/go-prettier/v3@" + conf.verGoPrettier)
		RegisterFormatTask(goyek.Define(goyek.Task{
			Name:     "format-markdown",
			Usage:    "Formats Markdown code.",
			Parallel: true,
			Action: func(a *goyek.A) {
				cmd.Exec(a, fmt.Sprintf("go run github.com/wasilibs/go-prettier/v3/cmd/prettier@%s --no-error-on-unmatched-pattern --write '**/*.md'", conf.verGoPrettier))
			},
		}))
	}

	if !conf.excluded("lint-markdown") {
		RegisterModuleDownloads("github.com/wasilibs/go-prettier/v3@" + conf.verGoPrettier)
		RegisterLintTask(goyek.Define(goyek.Task{
			Name:     "lint-markdown",
			Usage:    "Lints Markdown code.",
			Parallel: true,
			Action: func(a *goyek.A) {
				cmd.Exec(a, fmt.Sprintf("go run github.com/wasilibs/go-prettier/v3/cmd/prettier@%s --no-error-on-unmatched-pattern --check '**/*.md'", conf.verGoPrettier))
			},
		}))
	}

	if !conf.excluded("format-shell") {
		RegisterModuleDownloads("github.com/wasilibs/go-prettier/v3@" + conf.verGoPrettier)
		RegisterFormatTask(goyek.Define(goyek.Task{
			Name:     "format-shell",
			Usage:    "Formats shell-like code, including Dockerfile, ignore, dotenv.",
			Parallel: true,
			Action: func(a *goyek.A) {
				cmd.Exec(a, fmt.Sprintf("go run github.com/wasilibs/go-prettier/v3/cmd/prettier@%s --no-error-on-unmatched-pattern --write '**/*.sh' '**/*.bash' '**/Dockerfile' '**/*.dockerfile' '**/.*ignore' '**/.env*'", conf.verGoPrettier))
			},
		}))
	}

	if !conf.excluded("lint-shell") {
		RegisterModuleDownloads("github.com/wasilibs/go-prettier/v3@" + conf.verGoPrettier)
		RegisterLintTask(goyek.Define(goyek.Task{
			Name:     "lint-shell",
			Usage:    "Lints shell-like code, including Dockerfile, ignore, dotenv.",
			Parallel: true,
			Action: func(a *goyek.A) {
				cmd.Exec(a, fmt.Sprintf("go run github.com/wasilibs/go-prettier/v3/cmd/prettier@%s --no-error-on-unmatched-pattern --check '**/*.sh' '**/*.bash' '**/Dockerfile' '**/*.dockerfile' '**/.*ignore' '**/.env*'", conf.verGoPrettier))
			},
		}))
	}

	if !conf.excluded("format-yaml") {
		RegisterModuleDownloads("github.com/wasilibs/go-prettier/v3@" + conf.verGoPrettier)
		RegisterFormatTask(goyek.Define(goyek.Task{
			Name:     "format-yaml",
			Usage:    "Formats YAML code.",
			Parallel: true,
			Action: func(a *goyek.A) {
				cmd.Exec(a, fmt.Sprintf("go run github.com/wasilibs/go-prettier/v3/cmd/prettier@%s --no-error-on-unmatched-pattern --write '**/*.yaml' '**/*.yml'", conf.verGoPrettier))
			},
		}))
	}

	if !conf.excluded("lint-yaml") {
		RegisterModuleDownloads(
			"github.com/wasilibs/go-prettier/v3@"+conf.verGoPrettier,
			"github.com/wasilibs/go-yamllint@"+conf.verGoYamllint,
		)
		RegisterLintTask(goyek.Define(goyek.Task{
			Name:     "lint-yaml",
			Usage:    "Lints YAML code.",
			Parallel: true,
			Action: func(a *goyek.A) {
				cmd.Exec(a, fmt.Sprintf("go run github.com/wasilibs/go-prettier/v3/cmd/prettier@%s --no-error-on-unmatched-pattern --check '**/*.yaml' '**/*.yml'", conf.verGoPrettier))

				if root == "" {
					cmd.Exec(a, fmt.Sprintf("go run github.com/wasilibs/go-yamllint/cmd/yamllint@%s .", conf.verGoYamllint))
				} else {
					cmd.Exec(a, fmt.Sprintf("go run github.com/wasilibs/go-yamllint/cmd/yamllint@%s %s", conf.verGoYamllint, target), cmd.Dir(root))
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
		RegisterModuleDownloads(
			"github.com/suzuki-shunsuke/pinact/v3@"+conf.verPinact,
			"github.com/rhysd/actionlint@"+conf.verActionlint,
		)
		RegisterLintTask(goyek.Define(goyek.Task{
			Name:     "lint-github",
			Usage:    "Lints GitHub Actions workflows.",
			Parallel: true,
			Action: func(a *goyek.A) {
				cmd.Exec(a, fmt.Sprintf("go run github.com/suzuki-shunsuke/pinact/v3/cmd/pinact@%s run -check", conf.verPinact))
				cmd.Exec(a, fmt.Sprintf(`go run github.com/rhysd/actionlint/cmd/actionlint@%s -shellcheck="go run github.com/wasilibs/go-shellcheck/cmd/shellcheck@%s"`, conf.verActionlint, conf.verGoShellcheck))
			},
		}))
	}

	goyek.Define(goyek.Task{
		Name:  "download",
		Usage: "Downloads build dependencies.",
		Action: func(a *goyek.A) {
			cmd.Exec(a, "go mod download")
			cmd.Exec(a, "go mod download "+strings.Join(slices.Collect(maps.Keys(moduleDownloads)), " "))
		},
	})

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
	artifactsPath    string
	excludeTasks     []string
	buildTags        []string
	disableReviewdog bool

	verActionlint   string
	verGolangCILint string
	verGoPrettier   string
	verGoYamllint   string
	verGoShellcheck string
	verPinact       string
	verReviewdog    string
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

// DisableReviewdog returns an Option to disable the use of reviewdog to process lint output.
// By default, reviewdog is used to report lint issues as GitHub checks.
func DisableReviewdog() Option {
	return disableReviewdog{}
}

type disableReviewdog struct{}

func (d disableReviewdog) apply(conf *config) {
	conf.disableReviewdog = true
}

func execReviewdog(conf config, a *goyek.A, format string, cmdLine string, opts ...cmd.Option) bool {
	if conf.disableReviewdog || os.Getenv("CI") != "true" {
		return cmd.Exec(a, cmdLine, opts...)
	}
	var stderr bytes.Buffer
	if cmd.Exec(a, cmdLine, append(opts, cmd.Stderr(&stderr))...) {
		return true
	}
	return cmd.Exec(a, fmt.Sprintf("go github.com/reviewdog/reviewdog/cmd/reviewdog@%s %s -fail-level=warning -reporter=github-check", conf.verReviewdog, format), cmd.Stdin(&stderr))
}

// VersionActionlint returns an Option to set the version of actionlint to use. If unset,
// a default version is used which may not be the latest.
func VersionActionlint(version string) Option {
	return versionActionlint(version)
}

type versionActionlint string

func (v versionActionlint) apply(c *config) {
	c.verActionlint = string(v)
}

// VersionGolangCILint returns an Option to set the version of golangci-lint to use. If unset,
// a default version is used which may not be the latest.
func VersionGolangCILint(version string) Option {
	return versionGolangCILint(version)
}

type versionGolangCILint string

func (v versionGolangCILint) apply(c *config) {
	c.verGolangCILint = string(v)
}

// VersionGoPrettier returns an Option to set the version of go-prettier to use. If unset,
// a default version is used which may not be the latest.
func VersionGoPrettier(version string) Option {
	return versionGoPrettier(version)
}

type versionGoPrettier string

func (v versionGoPrettier) apply(c *config) {
	c.verGoPrettier = string(v)
}

// VersionGoShellcheck returns an Option to set the version of go-shellcheck to use. If unset,
// a default version is used which may not be the latest.
func VersionGoShellcheck(version string) Option {
	return versionGoShellcheck(version)
}

type versionGoShellcheck string

func (v versionGoShellcheck) apply(c *config) {
	c.verGoShellcheck = string(v)
}

// VersionGoYamllint returns an Option to set the version of go-yamllint to use. If unset,
// a default version is used which may not be the latest.
func VersionGoYamllint(version string) Option {
	return versionGoYamllint(version)
}

type versionGoYamllint string

func (v versionGoYamllint) apply(c *config) {
	c.verGoYamllint = string(v)
}

// VersionPinact returns an Option to set the version of pinact to use. If unset,
// a default version is used which may not be the latest.
func VersionPinact(version string) Option {
	return versionPinact(version)
}

type versionPinact string

func (v versionPinact) apply(c *config) {
	c.verPinact = string(v)
}

// VersionReviewdog returns an Option to set the version of reviewdog to use. If unset,
// a default version is used which may not be the latest.
func VersionReviewdog(version string) Option {
	return versionReviewdog(version)
}

type versionReviewdog string

func (v versionReviewdog) apply(c *config) {
	c.verReviewdog = string(v)
}
