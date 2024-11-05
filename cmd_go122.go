//go:build !go1.23

package build

import (
	"github.com/goyek/goyek/v2"
)

func goModTidyDiff(a *goyek.A) bool {
	a.Log("skipping go mod tidy -diff on < go1.23")
	return true
}
