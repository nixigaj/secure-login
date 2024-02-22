package main

import (
	"fmt"
	"github.com/alexflint/go-arg"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
)

const (
	publicDir         = "public"
	distDir           = "embed/dist"
	compressThreshold = 1000
)

var (
	generatedFilesFolders = []string{
		"secure-login",
		"secure-login.exe",
		"data.db",
		"embed/dist/",
	}

	distExtensions = []string{
		".js",
		".ico",
		".svg",
		".html",
		".css",
	}

	compressExtensions = []string{
		".js",
		".svg",
		".html",
		".css",
	}
)

func main() {
	var args struct {
		Option *string `arg:"positional" default:"build" help:"build | release | clean"`
	}
	parser := arg.MustParse(&args)

	var err error
	switch *args.Option {
	case "build":
		err = build(false)
		break
	case "release":
		err = build(true)
		break
	case "clean":
		err = clean()
		break
	default:
		parser.Fail(fmt.Sprintf("Unrecognized option: %s", *args.Option))
	}
	if err != nil {
		fmt.Println("Build error", err)
		os.Exit(1)
	}

	os.Exit(0)
}

func build(release bool) error {
	if release {
		err := genDist()
		if err != nil {
			return fmt.Errorf("frontend bundling: %v", err)
		}
	} else {
		// TODO create dist dir anyways
	}

	files, err := filepath.Glob("cmd/*.go")
	if err != nil {
		return err
	}

	cmdArgs := []string{"build"}
	if release {
		cmdArgs = append(cmdArgs, "--ldflags=-s", "--ldflags=-w")
	}
	cmdArgs = append(cmdArgs, "-o", "secure-login")
	cmdArgs = append(cmdArgs, files...)

	cmd := exec.Command("go", cmdArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func genDist() error {
	err := filepath.Walk(publicDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		destPath := filepath.Join(distDir, path[len(publicDir):])

		if info.IsDir() {
			err := os.MkdirAll(destPath, os.ModePerm)
			if err != nil {
				return err
			}
			return nil
		}

		if !slices.Contains(distExtensions, filepath.Ext(path)) {
			return nil
		}

		inFile, err := os.Open(path)
		if err != nil {
			return err
		}

		outFile, err := os.Create(destPath)
		if err != nil {
			return err
		}

		_, err = io.Copy(outFile, inFile)
		if err != nil {
			return err
		}

		err = inFile.Close()
		if err != nil {
			return err
		}

		err = outFile.Close()
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func clean() error {
	for _, path := range generatedFilesFolders {
		err := os.RemoveAll(path)
		if err != nil {
			return err
		}
	}
	return nil
}
