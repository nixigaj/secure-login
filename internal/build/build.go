package build

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"github.com/alexflint/go-arg"
	"github.com/andybalholm/brotli"
	"github.com/klauspost/compress/zstd"
	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/css"
	"github.com/tdewolff/minify/v2/html"
	"github.com/tdewolff/minify/v2/js"
	"github.com/tdewolff/minify/v2/svg"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
)

const (
	publicDir         = "public"
	distDir           = "internal/embed/dist"
	compressThreshold = 1000
)

var (
	generatedFilesFolders = []string{
		"secure-login",
		"secure-login.exe",
		"build",
		"build.exe",
		"data/",
		distDir,
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

	minifier *minify.M
)

func init() {
	minifier = minify.New()
	minifier.AddFunc(".js", js.Minify)
	minifier.AddFunc(".svg", svg.Minify)
	minifier.AddFunc(".html", html.Minify)
	minifier.AddFunc(".css", css.Minify)
}

func Run() int {
	var args struct {
		Option *string `arg:"positional" default:"debug" help:"debug | release | clean"`
	}
	parser := arg.MustParse(&args)

	var err error
	switch *args.Option {
	case "debug":
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
		return 1
	}

	return 0
}

func build(release bool) error {
	if release {
		err := genDist()
		if err != nil {
			return fmt.Errorf("frontend bundling: %v", err)
		}
	} else {
		err := genDummyDist()
		if err != nil {
			return fmt.Errorf("embed directory: %v", err)
		}
	}

	files, err := filepath.Glob("cmd/*.go")
	if err != nil {
		return err
	}

	cmdArgs := []string{"build"}
	if release {
		cmdArgs = append(cmdArgs, "--ldflags=-s", "--ldflags=-w")
		cmdArgs = append(cmdArgs, "--ldflags=-X main.secureLoginBuildMode=release")
	}
	//noinspection GoBoolExpressions
	if runtime.GOOS == "windows" {
		cmdArgs = append(cmdArgs, "-o", "secure-login.exe")
	} else {
		cmdArgs = append(cmdArgs, "-o", "secure-login")
	}
	cmdArgs = append(cmdArgs, files...)

	cmd := exec.Command("go", cmdArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func genDummyDist() error {
	parentDir := filepath.Dir(distDir)
	parentInfo, err := os.Stat(parentDir)
	if err != nil {
		return err
	}
	err = os.MkdirAll(distDir, parentInfo.Mode().Perm())
	if err != nil {
		return err
	}

	containsFile, err := dirContainsFile(distDir)
	if err != nil {
		return err
	}
	if containsFile {
		return nil
	}

	file, err := os.Create(filepath.Join(distDir, ".dummy"))
	if err != nil {
		return err
	}
	err = file.Close()
	if err != nil {
		return err
	}

	return nil
}

func genDist() error {
	err := os.RemoveAll(distDir)
	if err != nil {
		return err
	}
	parentDir := filepath.Dir(distDir)
	parentInfo, err := os.Stat(parentDir)
	if err != nil {
		return err
	}
	err = os.MkdirAll(distDir, parentInfo.Mode().Perm())
	if err != nil {
		return err
	}

	err = filepath.Walk(publicDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		destPath := filepath.Join(distDir, path[len(publicDir):])

		if info.IsDir() {
			containsFile, err := dirContainsFile(path)
			if !containsFile {
				return nil
			}
			err = os.MkdirAll(destPath, parentInfo.Mode().Perm())
			if err != nil {
				return err
			}
			return nil
		}

		if !slices.Contains(distExtensions, filepath.Ext(path)) {
			return nil
		}

		err = handleDistFile(path, destPath, parentInfo.Mode().Perm())
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

func handleDistFile(srcPath string, destPath string, perm os.FileMode) error {
	inData, err := os.ReadFile(srcPath)
	if err != nil {
		return err
	}

	ext := filepath.Ext(srcPath)
	isCompressExtension := slices.Contains(compressExtensions, ext)

	var minData []byte
	if isCompressExtension && !strings.Contains(srcPath, ".min.") {
		minData, err = minifier.Bytes(ext, inData)
		if err != nil {
			return err
		}
	} else {
		minData = inData
	}

	err = os.WriteFile(destPath, minData, perm)
	if err != nil {
		return err
	}

	if !(isCompressExtension && len(minData) >= compressThreshold) {
		return nil
	}

	var gzipData bytes.Buffer
	gz := gzip.NewWriter(&gzipData)
	_, err = gz.Write(minData)
	if err != nil {
		return err
	}
	err = gz.Close()
	if err != nil {
		return err
	}
	err = os.WriteFile(destPath+".gz", gzipData.Bytes(), perm)
	if err != nil {
		return err
	}

	var brotliData bytes.Buffer
	br := brotli.NewWriter(&brotliData)
	_, err = br.Write(minData)
	if err != nil {
		return err
	}
	err = br.Close()
	if err != nil {
		return err
	}
	err = os.WriteFile(destPath+".br", brotliData.Bytes(), perm)
	if err != nil {
		return err
	}

	var zstData bytes.Buffer
	zst, err := zstd.NewWriter(&zstData)
	if err != nil {
		return err
	}
	_, err = zst.Write(minData)
	if err != nil {
		return err
	}
	err = zst.Close()
	if err != nil {
		return err
	}
	err = os.WriteFile(destPath+".zst", zstData.Bytes(), perm)
	if err != nil {
		return err
	}

	return nil
}

func dirContainsFile(path string) (bool, error) {
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			return io.EOF
		}
		return nil
	})
	if err == io.EOF {
		return true, nil
	}
	if err != nil {
		return false, err
	}
	return false, nil
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
