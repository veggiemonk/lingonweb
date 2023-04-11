package main

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"
	"time"

	"github.com/ardanlabs/conf/v3"
	_ "go.uber.org/automaxprocs"
	"golang.org/x/exp/slog"
)

const serviceName = "lingon"

var (
	// //go:embed embed
	// webapp embed.FS

	//go:embed embed/index.html
	indexhtml []byte
)

//
// init function to register types to runtime.NewScheme() in another file
//

func main() {
	log := makeLogger(os.Stderr)

	if err := run(log); err != nil {
		log.Error("run", "err", err)
		os.Exit(1) //nolint:gocritic
	}
}

func run(log *slog.Logger) error {
	cfg := struct {
		conf.Version
		Port            int           `conf:"default:8080,env:PORT"`
		Host            string        `conf:"default:0.0.0.0"`
		HealthPath      string        `conf:"default:/healthz"`
		VersionPath     string        `conf:"default:/version"`
		ReadTimeout     time.Duration `conf:"default:5s"`
		WriteTimeout    time.Duration `conf:"default:10s"`
		IdleTimeout     time.Duration `conf:"default:120s"`
		ShutdownTimeout time.Duration `conf:"default:5s"`
	}{
		Version: conf.Version{
			Build: commit,
			Desc:  serviceName + "web service",
		},
	}

	const prefix = serviceName
	help, err := conf.Parse(prefix, &cfg)
	if err != nil {
		if errors.Is(err, conf.ErrHelpWanted) {
			fmt.Println(help)
			return nil
		}

		return fmt.Errorf("parsing config: %w", err)
	}

	// Closing signal
	stopChan := make(chan os.Signal, 1)
	signal.Notify(
		stopChan,
		// syscall.SIGKILL, // never gets caught on POSIX
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)

	ms := &runtime.MemStats{}
	runtime.ReadMemStats(ms)
	log.Info(
		fmt.Sprintf("Starting service... %d", time.Now().UTC().Unix()),
		slog.Int("CPU cores", runtime.NumCPU()),
		slog.String("Available Memory", fmt.Sprintf("%d MB", ms.Sys/1024)),
	)
	defer log.Info("Service stopped")

	sm := http.NewServeMux()
	sm.HandleFunc("/convert", convert(log))
	sm.HandleFunc(cfg.VersionPath, VersionInfo)
	sm.HandleFunc(cfg.HealthPath, healthz)
	//
	// // static files when embedding a whole directory
	// // but we only want to serve index.html
	//
	// static, err := fs.Sub(webapp, "embed")
	// if err != nil {
	// 	return fmt.Errorf("getting webapp: %w", err)
	// }
	// sm.Handle("/", http.FileServer(http.FS(static)))
	sm.Handle("/", byteHandler(indexhtml))

	srv := &http.Server{
		Addr:                         fmt.Sprintf(":%d", cfg.Port),
		Handler:                      sm,
		DisableGeneralOptionsHandler: false,
		// Good practice to set timeouts to avoid Slowloris attacks.
		WriteTimeout:      cfg.WriteTimeout,
		ReadTimeout:       cfg.ReadTimeout,
		ReadHeaderTimeout: cfg.ReadTimeout,
		IdleTimeout:       cfg.IdleTimeout,
	}
	go func() {
		if zerr := srv.ListenAndServe(); zerr != nil &&
			!errors.Is(zerr, http.ErrServerClosed) {
			log.Error("failed to start server: %v", "err", zerr)
			panic(zerr)
		}
	}()

	<-stopChan
	ctxShutDown, cancel := context.WithTimeout(
		context.Background(),
		time.Minute,
	)
	defer func() { cancel() }()

	log.Info("shutting down the service...")
	if err = srv.Shutdown(ctxShutDown); err != nil {
		log.Error("server Shutdown Failed", "err", err)
	}

	return nil
}

func healthz(w http.ResponseWriter, _ *http.Request) {
	_, _ = io.WriteString(w, "ok")
}

func byteHandler(b []byte) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write(b)
	}
}

// makeLogger returns a logger that writes to w [io.Writer]. If w is nil, os.Stderr is used.
// Timestamp is removed and directory from the source's filename is shown.
func makeLogger(w io.Writer) *slog.Logger {
	if w == nil {
		w = os.Stderr
	}
	return slog.New(
		slog.HandlerOptions{
			AddSource:   true,
			ReplaceAttr: logReplace,
		}.NewJSONHandler(w).WithAttrs(
			[]slog.Attr{slog.String("app", serviceName)},
		),
	)
}

func logReplace(_ []string, a slog.Attr) slog.Attr {
	// // Remove time.
	// if a.Key == slog.TimeKey && len(groups) == 0 {
	// 	a.Key = ""
	// }
	// Remove the directory from the source's filename.
	if a.Key == slog.SourceKey {
		a.Value = slog.StringValue(filepath.Base(a.Value.String()))
	}
	return a
}
