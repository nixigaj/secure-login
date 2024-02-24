package embed

import "embed"

//go:embed dist/*
var Dir embed.FS
