//go:build ignore

package main

import (
	"github.com/nixigaj/secure-login/internal/build"
	"os"
)

func main() {
	os.Exit(build.Run())
}
