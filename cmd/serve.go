package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/charmbracelet/log"
	"github.com/nixigaj/secure-login/internal/embed"
	"github.com/nixigaj/secure-login/internal/fileserver"
	"net"
	"net/http"
	"strings"
	"time"
)

const (
	// As strict as possible while still allowing
	// WASM to execute for the client-side hashing.
	cspHeaderValue = "default-src 'self'; script-src 'self' 'wasm-unsafe-eval'"

	// Is only set if TLS mode is active.
	serverHeaderValue = "Secure login demo"
)

func serve(args *Args, sc *syncController, tlsMode bool) {
	mux := http.NewServeMux()

	if secureLoginReleaseMode {
		fileServer := fileserver.New(embed.Dir, "dist")

		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			if !args.DisableCSP {
				w.Header().Set("Content-Security-Policy", cspHeaderValue)
			}
			if tlsMode {
				w.Header().Set("Server", serverHeaderValue)
			}
			fileServer.ServeHTTP(w, r)
		})
	} else {
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			if !args.DisableCSP {
				w.Header().Set("Content-Security-Policy", cspHeaderValue)
			}
			if tlsMode {
				w.Header().Set("Server", serverHeaderValue)
			}
			http.FileServer(http.Dir("public")).ServeHTTP(w, r)
		})
	}

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

	go func() {
		var err error
		if tlsMode {
			log.Infof("Listening on https://%s", addr)
			err = srv.ListenAndServeTLS(*args.TlsCert, *args.TlsKey)
		} else {
			log.Infof("Listening on http://%s", addr)
			err = srv.ListenAndServe()
		}
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			sc.fatalChan <- fmt.Errorf("HTTP server startup: %v", err)
		}
	}()

	<-sc.stopChan

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err := srv.Shutdown(ctx)
	if err != nil {
		log.Errorf("HTTP stutdown: %v", err)
	}

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
