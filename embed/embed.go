// This directory is filled with minified and compressed

package embed

import "embed"

//go:embed dist/*
var Dir embed.FS
