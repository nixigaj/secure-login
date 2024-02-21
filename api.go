package main

import (
	"fmt"
	"github.com/charmbracelet/log"
	"net/http"
)

func apiHandler(w http.ResponseWriter, r *http.Request, args *Args) {
	defer log.Debugf("Handled HTTP request from %s", r.Host)

	_, err := fmt.Fprintf(w, "Hello %s from %s\n", r.Host, *args.Bind)
	if err != nil {
		log.Errorf("HTTP write: %v", err)
	}
}
