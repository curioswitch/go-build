package build

import (
	"github.com/goyek/goyek/v2"
)

var (
	formatTasks   goyek.Deps
	generateTasks goyek.Deps
	lintTasks     goyek.Deps
	testTasks     goyek.Deps
)

// RegisterFormatTask adds a task that should be run during the format command.
func RegisterFormatTask(task *goyek.DefinedTask) {
	formatTasks = append(formatTasks, task)
}

// RegisterGenerateTask adds a task that should be run during the generate command.
func RegisterGenerateTask(task *goyek.DefinedTask) {
	generateTasks = append(generateTasks, task)
}

// RegisterLintTask adds a task that should be run during the lint command.
func RegisterLintTask(task *goyek.DefinedTask) {
	lintTasks = append(lintTasks, task)
}

// RegisterTestTask adds a task that should be run during the test command.
func RegisterTestTask(task *goyek.DefinedTask) {
	testTasks = append(testTasks, task)
}
