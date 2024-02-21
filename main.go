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

type syncController struct {
	stopChan  chan struct{}
	fatalChan chan error
	wg        sync.WaitGroup
}

type Args struct {
	Bind         *string `arg:"positional" default:"localhost:8080" help:"IPv4 or IPv6 address with optional port to listen on"`
	Debug        bool    `arg:"-d,--debug" help:"print debug logs"`
	PrintVersion bool    `arg:"-V,-v,--version" help:"print program version"`
}

func main() {
	log.SetTimeFormat(time.DateTime)

	var args Args
	arg.MustParse(&args)

	if args.PrintVersion {
		fmt.Println("Secure login demo version", secureLoginVersion)
		os.Exit(0)
	}

	stopChan := make(chan struct{})
	go exitHandler(stopChan)

	sc := syncController{
		stopChan:  stopChan,
		fatalChan: make(chan error),
		wg:        sync.WaitGroup{},
	}

	sc.wg.Add(1)
	go serve(&args, &sc)

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

	<-stopSig
	close(stopChan)

	// Force exit on second signal from stopSig
	log.Fatal("force exit:", "signal", <-stopSig)
}
