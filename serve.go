package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/charmbracelet/log"
	"github.com/nixigaj/secure-login/public"
	"github.com/vearutop/statigz"
	"github.com/vearutop/statigz/brotli"
	"net"
	"net/http"
	"strings"
	"time"
)

func serve(args *Args, sc *syncController) {
	fileServer := statigz.FileServer(
		public.Dir,
		brotli.AddEncoding)

	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fileServer.ServeHTTP(w, r)
	})
	mux.HandleFunc("/api", func(w http.ResponseWriter, r *http.Request) {
		apiHandler(w, r, args)
	})
	mux.HandleFunc("/api/", func(w http.ResponseWriter, r *http.Request) {
		apiHandler(w, r, args)
	})

	addr := parseBind(*args.Bind)
	srv := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	log.Infof("HTTP listening on %s", addr)

	go func() {
		err := srv.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			sc.fatalChan <- fmt.Errorf("HTTP server startup: %v", err)
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	<-sc.stopChan

	err := srv.Shutdown(ctx)
	if err != nil {
		log.Errorf("HTTP stutdown: %v", err)
	}
	cancel()

	sc.wg.Done()
}

func parseBind(bind string) string {
	if strings.Count(bind, ":") == 1 || strings.Contains(bind, "]:") {
		return bind
	} else if ip := net.ParseIP(bind); ip != nil {
		if ip.To4() != nil {
			return fmt.Sprintf("%s:8080", bind)
		} else {
			return fmt.Sprintf("[%s]:8080", bind)
		}
	}
	return fmt.Sprintf("%s:8080", bind)
}
