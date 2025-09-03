package build

import (
	"github.com/goyek/goyek/v2"
)

var (
	formatTasks   goyek.Deps
	generateTasks goyek.Deps
	lintTasks     goyek.Deps
	testTasks     goyek.Deps

	commandDownloads = map[string]struct{}{}
)

// RegisterFormatTask adds a task that should be run during the format task.
func RegisterFormatTask(task *goyek.DefinedTask) {
	formatTasks = append(formatTasks, task)
}

// RegisterGenerateTask adds a task that should be run during the generate task.
func RegisterGenerateTask(task *goyek.DefinedTask) {
	generateTasks = append(generateTasks, task)
}

// RegisterLintTask adds a task that should be run during the lint task.
func RegisterLintTask(task *goyek.DefinedTask) {
	lintTasks = append(lintTasks, task)
}

// RegisterTestTask adds a task that should be run during the test task.
func RegisterTestTask(task *goyek.DefinedTask) {
	testTasks = append(testTasks, task)
}

// RegisterCommandDownloads registers the command to be downloaded by the download task.
// It will be executed as is - to download Go tools, it should be a valid `go run` command
// that will exit successfully.
func RegisterCommandDownloads(commands ...string) {
	for _, module := range commands {
		commandDownloads[module] = struct{}{}
	}
}
