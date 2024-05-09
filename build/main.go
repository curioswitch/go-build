package main

import (
	"github.com/goyek/x/boot"

	"github.com/curioswitch/go-build"
)

func main() {
	build.DefineTasks(
		build.LocalPackagePrefix("github.com/curioswitch/go-build"),
	)
	boot.Main()
}
