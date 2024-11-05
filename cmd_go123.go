//go:build go1.23

package build

import (
	"github.com/goyek/goyek/v2"
	"github.com/goyek/x/cmd"
)

func goModTidyDiff(a *goyek.A) bool {
	return cmd.Exec(a, "go mod tidy -diff")
}
