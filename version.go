package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"runtime/debug"
)

var (
	// ldflags
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

type V struct {
	Version   string          `json:"version"`
	Commit    string          `json:"commit"`
	Date      string          `json:"date"`
	BuildInfo debug.BuildInfo `json:"buildInfo"`
}

func VersionInfo(w http.ResponseWriter, r *http.Request) {
	bi, ok := debug.ReadBuildInfo()
	if !ok {
		_, _ = fmt.Fprintln(os.Stderr, "error reading build-info")
		_, _ = fmt.Fprintln(w, "error reading build-info")
	}
	v := V{
		Version:   version,
		Commit:    commit,
		Date:      date,
		BuildInfo: *bi,
	}
	if err := json.NewEncoder(w).Encode(v); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "error encoding version: %v", err)
		_, _ = fmt.Fprintf(w, "error encoding version: %v", err)
	}
}
