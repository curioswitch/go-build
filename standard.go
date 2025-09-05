package build

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
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

	if conf.verGoTestsum == "" {
		conf.verGoTestsum = verGoTestsum
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

	var golangciTargets []string
	// Rare to not have a go.mod, except for a monorepo root where it's common.
	hasGoMod := false
	if fileExists("go.mod") {
		golangciTargets = append(golangciTargets, "./...")
		hasGoMod = true
	}
	// Uses of go-build will very commonly have a build folder, if it is also a module,
	// then let's automatically run checks on it.
	if fileExists(filepath.Join("build", "go.mod")) {
		golangciTargets = append(golangciTargets, "./build")
	}

	root, target := pathRelativeToRoot()

	runActionlint := "go run github.com/rhysd/actionlint/cmd/actionlint@" + conf.verActionlint
	runGolangCILint := "go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@" + conf.verGolangCILint
	runGoPrettier := "go run github.com/wasilibs/go-prettier/v3/cmd/prettier@" + conf.verGoPrettier
	runGoShellcheck := "go run github.com/wasilibs/go-shellcheck/cmd/shellcheck@" + conf.verGoShellcheck
	runGoTestsum := "go run gotest.tools/gotestsum@" + conf.verGoTestsum
	runGoYamllint := "go run github.com/wasilibs/go-yamllint/cmd/yamllint@" + conf.verGoYamllint
	runPinact := "go run github.com/suzuki-shunsuke/pinact/v3/cmd/pinact@" + conf.verPinact
	runReviewDog := "go run github.com/reviewdog/reviewdog/cmd/reviewdog@" + conf.verReviewdog

	if !conf.excluded("format-go") {
		RegisterCommandDownloads(runGolangCILint)
		RegisterFormatTask(goyek.Define(goyek.Task{
			Name:     "format-go",
			Usage:    "Formats Go code.",
			Parallel: true,
			Action: func(a *goyek.A) {
				cmd.Exec(a, fmt.Sprintf(`%s fmt %s`, runGolangCILint, strings.Join(golangciTargets, " ")))
				if hasGoMod {
					cmd.Exec(a, "go mod tidy")
				}
			},
		}))
	}

	if !conf.excluded("lint-go") {
		RegisterCommandDownloads(runGolangCILint, runReviewDog+" -version")
		RegisterLintTask(goyek.Define(goyek.Task{
			Name:     "lint-go",
			Usage:    "Lints Go code.",
			Parallel: true,
			Action: func(a *goyek.A) {
				execReviewdog(conf, a, runReviewDog, "-f=golangci-lint -name=golangci-lint",
					fmt.Sprintf(`go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@%s run --build-tags "%s" --timeout=20m %s`,
						conf.verGolangCILint, strings.Join(conf.buildTags, ","), strings.Join(golangciTargets, " ")))
				if hasGoMod {
					cmd.Exec(a, "go mod tidy -diff")
				}
			},
		}))
	}

	if !conf.excluded("format-markdown") {
		RegisterCommandDownloads(runGoPrettier)
		RegisterFormatTask(goyek.Define(goyek.Task{
			Name:     "format-markdown",
			Usage:    "Formats Markdown code.",
			Parallel: true,
			Action: func(a *goyek.A) {
				cmd.Exec(a, runGoPrettier+" --no-error-on-unmatched-pattern --write '**/*.md'")
			},
		}))
	}

	if !conf.excluded("lint-markdown") {
		RegisterCommandDownloads(runGoPrettier)
		RegisterLintTask(goyek.Define(goyek.Task{
			Name:     "lint-markdown",
			Usage:    "Lints Markdown code.",
			Parallel: true,
			Action: func(a *goyek.A) {
				cmd.Exec(a, runGoPrettier+" --no-error-on-unmatched-pattern --check '**/*.md'")
			},
		}))
	}

	if !conf.excluded("format-shell") {
		RegisterCommandDownloads(runGoPrettier)
		RegisterFormatTask(goyek.Define(goyek.Task{
			Name:     "format-shell",
			Usage:    "Formats shell-like code, including Dockerfile, ignore, dotenv.",
			Parallel: true,
			Action: func(a *goyek.A) {
				cmd.Exec(a, runGoPrettier+" --no-error-on-unmatched-pattern --write '**/*.sh' '**/*.bash' '**/Dockerfile' '**/*.dockerfile' '**/.*ignore' '**/.env*'")
			},
		}))
	}

	if !conf.excluded("lint-shell") {
		RegisterCommandDownloads(runGoPrettier)
		RegisterLintTask(goyek.Define(goyek.Task{
			Name:     "lint-shell",
			Usage:    "Lints shell-like code, including Dockerfile, ignore, dotenv.",
			Parallel: true,
			Action: func(a *goyek.A) {
				cmd.Exec(a, runGoPrettier+" --no-error-on-unmatched-pattern --check '**/*.sh' '**/*.bash' '**/Dockerfile' '**/*.dockerfile' '**/.*ignore' '**/.env*'")
			},
		}))
	}

	if !conf.excluded("format-yaml") {
		RegisterCommandDownloads(runGoPrettier)
		RegisterFormatTask(goyek.Define(goyek.Task{
			Name:     "format-yaml",
			Usage:    "Formats YAML code.",
			Parallel: true,
			Action: func(a *goyek.A) {
				cmd.Exec(a, runGoPrettier+" --no-error-on-unmatched-pattern --write '**/*.yaml' '**/*.yml'")
			},
		}))
	}

	if !conf.excluded("lint-yaml") {
		RegisterCommandDownloads(runGoPrettier, runGoYamllint+" -v")
		RegisterLintTask(goyek.Define(goyek.Task{
			Name:     "lint-yaml",
			Usage:    "Lints YAML code.",
			Parallel: true,
			Action: func(a *goyek.A) {
				cmd.Exec(a, runGoPrettier+" --no-error-on-unmatched-pattern --check '**/*.yaml' '**/*.yml'")

				if root == "" {
					cmd.Exec(a, runGoYamllint+" .")
				} else {
					cmd.Exec(a, runGoYamllint+" "+target, cmd.Dir(root))
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
				format := ""
				if conf.goTestsumFormat != "" {
					format = "--format=" + conf.goTestsumFormat
				}
				cmd.Exec(a, fmt.Sprintf("%s %s -- -coverprofile=%s -covermode=atomic -v -timeout=20m ./...", runGoTestsum, format, filepath.Join(conf.artifactsPath, "coverage.txt")))
			},
		}))
	}

	if !conf.excluded("runall") {
		RegisterGenerateTask(goyek.Define(goyek.Task{
			Name:  "runall",
			Usage: "Runs a command in each module in the workspace.",
			Action: func(a *goyek.A) {
				if *command == "" {
					a.Error("missing -cmd flag required for runall")
					return
				}
				for _, dir := range modDirs(a) {
					cmd.Exec(a, *command, cmd.Dir(dir))
				}
			},
		}))
	}

	if !conf.excluded("lint-github") && fileExists(".github") {
		RegisterCommandDownloads(runPinact, runActionlint)
		RegisterLintTask(goyek.Define(goyek.Task{
			Name:     "lint-github",
			Usage:    "Lints GitHub Actions workflows.",
			Parallel: true,
			Action: func(a *goyek.A) {
				cmd.Exec(a, runPinact+" run -check")
				cmd.Exec(a, fmt.Sprintf(`%s -shellcheck="%s"`, runActionlint, runGoShellcheck))
			},
		}))
	}

	goyek.Define(goyek.Task{
		Name:  "download",
		Usage: "Downloads build dependencies.",
		Action: func(a *goyek.A) {
			for _, dir := range modDirs(a) {
				cmd.Exec(a, "go mod download", cmd.Dir(dir))
			}
			if conf.downloadToolsAllOSes || runtime.GOOS == "linux" {
				for c := range commandDownloads {
					cmd.Exec(a, c, cmd.Stdout(io.Discard))
				}
			}
			// Ignore downloadTools for gotestsum
			if !conf.excluded("test-go") {
				cmd.Exec(a, runGoTestsum+" -h", cmd.Stdout(io.Discard))
			}
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
	goTestsumFormat  string

	verActionlint   string
	verGolangCILint string
	verGoPrettier   string
	verGoYamllint   string
	verGoShellcheck string
	verGoTestsum    string
	verPinact       string
	verReviewdog    string

	downloadToolsAllOSes bool
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

func modDirs(a *goyek.A) []string {
	var out bytes.Buffer
	if !cmd.Exec(a, "go list -m -f {{.Dir}}", cmd.Stdout(&out)) {
		return nil
	}
	return strings.Split(strings.TrimSpace(out.String()), "\n")
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

func execReviewdog(conf config, a *goyek.A, runReviewdog string, format string, cmdLine string, opts ...cmd.Option) bool {
	if conf.disableReviewdog || os.Getenv("CI") != "true" {
		return cmd.Exec(a, cmdLine, opts...)
	}
	var stderr bytes.Buffer
	if cmd.Exec(a, cmdLine, append(opts, cmd.Stderr(&stderr))...) {
		return true
	}
	return cmd.Exec(a, fmt.Sprintf("%s %s -fail-level=warning -reporter=github-check", runReviewdog, format), cmd.Stdin(&stderr))
}

// GoTestsumFormat returns an Option to customize the format reported by test results via gotestsum.
// See https://github.com/gotestyourself/gotestsum#output-format
func GoTestsumFormat(format string) Option {
	return goTestsumFormat(format)
}

type goTestsumFormat string

func (g goTestsumFormat) apply(c *config) {
	c.goTestsumFormat = string(g)
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

// VersionGoTestsum returns an Option to set the version of gotestsum to use. If unset,
// a default version is used which may not be the latest.
func VersionGoTestsum(version string) Option {
	return versionGoTestsum(version)
}

type versionGoTestsum string

func (v versionGoTestsum) apply(c *config) {
	c.verGoTestsum = string(v)
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

// DownloadToolsAllOSes returns an Option to download tools for all operating systems.
// By default, the `download` task only downloads tools on Linux to reflect that it is
// common to only run lints on Linux and tests on other OSes.
func DownloadToolsAllOSes() Option {
	return downloadToolsAllOSes{}
}

type downloadToolsAllOSes struct{}

func (d downloadToolsAllOSes) apply(c *config) {
	c.downloadToolsAllOSes = true
}
