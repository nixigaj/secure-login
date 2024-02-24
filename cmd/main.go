package main

import (
	"fmt"
	"github.com/alexflint/go-arg"
	"github.com/charmbracelet/log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

const (
	secureLoginVersion = "0.1.0"
)

var (
	// This variable is set at build time to determine whether
	// to serve files from embedded storage or the disk.
	secureLoginBuildMode   = "debug"
	secureLoginReleaseMode = false
)

type syncController struct {
	stopChan  chan struct{}
	fatalChan chan error
	wg        sync.WaitGroup
}

type Args struct {
	Bind         *string `arg:"positional" default:"localhost:8080" help:"IPv4 or IPv6 address with optional port to listen on"`
	DisableCSP   bool    `arg:"--disable-csp" help:"disable setting Content-Security-Policy header"`
	Debug        bool    `arg:"-d,--debug" help:"print debug logs"`
	TlsCert      *string `arg:"-c,--cert" help:"TLS certificate file path for HTTPS mode"`
	TlsKey       *string `arg:"-k,--key" help:"TLS key file path for HTTPS mode"`
	PrintVersion bool    `arg:"-V,-v,--version" help:"print program version"`
}

func init() {
	//noinspection GoBoolExpressions
	if secureLoginBuildMode == "release" {
		secureLoginReleaseMode = true
	}

	log.SetTimeFormat(time.DateTime)
	//noinspection GoBoolExpressions
	if secureLoginBuildMode == "debug" {
		log.SetLevel(log.DebugLevel)
		log.Debugf("Secure login demo built in debug mode")
	}
}

func main() {
	var args Args
	parser := arg.MustParse(&args)

	if args.PrintVersion {
		fmt.Println("Secure login demo version", secureLoginVersion)
		os.Exit(0)
	}

	if args.Debug {
		log.SetLevel(log.DebugLevel)
	}

	tlsMode := false
	if args.TlsCert != nil && args.TlsKey != nil {
		tlsMode = true
	} else if args.TlsCert == nil && args.TlsKey != nil {
		parser.Fail("TLS key provided but no TLS certificate provided")
	} else if args.TlsCert != nil && args.TlsKey == nil {
		parser.Fail("TLS certificate provided but no TLS key provided")
	}

	stopChan := make(chan struct{})
	go exitHandler(stopChan)

	sc := syncController{
		stopChan:  stopChan,
		fatalChan: make(chan error),
		wg:        sync.WaitGroup{},
	}

	sc.wg.Add(1)
	go serve(&args, &sc, tlsMode)

	select {
	case <-sc.stopChan:
		sc.wg.Wait()
		break
	case err := <-sc.fatalChan:
		log.Errorf("Exiting due to fatal runtime error: %v", err)
		close(sc.stopChan)
		sc.wg.Wait()
		log.Fatalf("Runtime: %v", err)
	}

	os.Exit(0)
}

func exitHandler(stopChan chan struct{}) {
	stopSig := make(chan os.Signal, 1)
	signal.Notify(stopSig, os.Interrupt, syscall.SIGTERM)

	log.Debug("received stop signal", "signal", <-stopSig)
	close(stopChan)

	// Force exit on second signal from stopSig
	log.Fatal("force exit:", "signal", <-stopSig)
}
