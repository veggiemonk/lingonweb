package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"

	conf "github.com/ardanlabs/conf/v3"
	"github.com/volvo-cars/lingon/pkg/kube"
	_ "go.uber.org/automaxprocs"
	"golang.org/x/exp/slog"
)

const serviceName = "lingon"

var build = "develop"

var (
	// //go:embed embed
	// webapp embed.FS

	//go:embed embed/index.html
	indexhtml []byte
)

func main() {
	log := makeLogger(os.Stderr)

	if err := run(log); err != nil {
		log.Error("startup", "err", err)
		os.Exit(1) //nolint:gocritic
	}
}

func run(log *slog.Logger) error {
	cfg := struct {
		conf.Version
		Port            int           `conf:"default:8080,env:PORT"`
		Host            string        `conf:"default:0.0.0.0"`
		ReadTimeout     time.Duration `conf:"default:5s"`
		WriteTimeout    time.Duration `conf:"default:10s"`
		IdleTimeout     time.Duration `conf:"default:120s"`
		ShutdownTimeout time.Duration `conf:"default:5s"`
	}{
		Version: conf.Version{
			Build: build,
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
		fmt.Sprintf("Starting service... %d", time.Now().Unix()),
		slog.Int("CPU cores", runtime.NumCPU()),
		slog.String("Available Memory", fmt.Sprintf("%d MB", ms.Sys/1024)),
	)
	defer log.Info("Service stopped")

	// static, err := fs.Sub(webapp, "embed")
	// if err != nil {
	// 	return fmt.Errorf("getting webapp: %w", err)
	// }
	sm := http.NewServeMux()
	sm.HandleFunc("/convert", NewHandler(log))
	sm.HandleFunc(
		"/healthz", func(w http.ResponseWriter, r *http.Request) {
			_, _ = io.WriteString(w, "ok")
		},
	)
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
		if zerr := srv.ListenAndServe(); zerr != nil && zerr != http.ErrServerClosed {
			log.Error("failed to start server: %v", "err", zerr)
			os.Exit(1) //nolint:gocritic
		}
	}()

	<-stopChan
	ctxShutDown, cancel := context.WithTimeout(
		context.Background(),
		time.Minute,
	)
	defer func() { cancel() }()

	if err = srv.Shutdown(ctxShutDown); err != nil {
		log.Error("server Shutdown Failed", "err", err)
	}

	log.Info("shutting down the service...")
	return nil
}

func byteHandler(b []byte) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write(b)
	}
}

type Meta struct {
	Length        string  `json:"length,omitempty"`
	Uuid          string  `json:"uuid,omitempty"`
	Digest        []uint8 `json:"digest,omitempty"`
	AddMethods    bool    `json:"addmethods"`
	GroupByKind   bool    `json:"groupbykind"`
	IgnoreErrors  bool    `json:"ignoreerrors"`
	RemoveAppName bool    `json:"removeappname"`
	Verbose       bool    `json:"verbose"`
}
type Input struct {
	Meta Meta   `json:"meta"`
	Data string `json:"data"`
}
type MetaOut struct {
	Length string  `json:"length,omitempty"`
	Uuid   string  `json:"uuid,omitempty"`
	Digest []uint8 `json:"digest,omitempty"`
	Logs   string  `json:"logs,omitempty"`
}
type Output struct {
	Meta   MetaOut `json:"meta"`
	Data   string  `json:"data"`
	Errors string  `json:"errors,omitempty"`
}

func NewHandler(log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Set the CORS headers to the response.
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set(
			"Access-Control-Allow-Methods",
			"POST, OPTIONS",
		)
		w.Header().Set(
			"Access-Control-Allow-Headers",
			"Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization",
		)
		if r.Method == http.MethodOptions {
			return
		}

		defer func(rc io.ReadCloser) {
			if err := rc.Close(); err != nil {
				log.Error("close body", "err", err)
			}
		}(r.Body)

		var buf bytes.Buffer
		if _, err := io.Copy(&buf, r.Body); err != nil {
			log.Error("read body", "err", err)
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write(
				[]byte(fmt.Sprintf("{\"errors\":%q}", err.Error())),
			)
			return
		}

		var msg Input
		if err := json.NewDecoder(&buf).Decode(&msg); err != nil {
			log.Error("failed to decode body", "err", err)
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write(
				[]byte(fmt.Sprintf("{\"errors\":%q}", err.Error())),
			)

			return
		}

		if msg.Meta.Digest == nil {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte("{\"errors\":\"digest is required\"}"))
			return
		}

		// calculate the digest SHA-256 of the message
		h := sha256.New()
		h.Write([]byte(msg.Data))
		received := fmt.Sprintf("%x", h.Sum(nil))
		computed := fmt.Sprintf("%x", sha256.Sum256([]byte(msg.Data)))
		if computed != received {
			w.WriteHeader(http.StatusBadRequest)
			errMsg := fmt.Sprintf(
				"{\"errors\":\"digest is invalid: %s\"}",
				received,
			)
			log.Info("digest is invalid", "err", errMsg)
			_, _ = w.Write([]byte(errMsg))
			return

		}

		var resw, lb bytes.Buffer
		nl := makeLogger(&lb)

		if err := kube.Import(
			kube.WithImportReader(strings.NewReader(msg.Data)),
			kube.WithImportWriter(&resw),
			kube.WithImportLogger(nl), // send this to the client
			kube.WithImportAppName(serviceName),
			kube.WithImportPackageName(serviceName),
			kube.WithImportSerializer(Codecs.UniversalDeserializer()),
			kube.WithImportVerbose(msg.Meta.Verbose),
			kube.WithImportGroupByKind(msg.Meta.GroupByKind),
			kube.WithImportAddMethods(msg.Meta.AddMethods),
			kube.WithImportRemoveAppName(msg.Meta.RemoveAppName),
			kube.WithImportIgnoreErrors(msg.Meta.IgnoreErrors),
		); err != nil {

			o := Output{
				Meta: MetaOut{
					Length: fmt.Sprintf("%d", resw.Len()),
					Uuid:   msg.Meta.Uuid,
					Digest: []byte(fmt.Sprintf(
						"%x",
						sha256.Sum256(resw.Bytes()),
					)),
					Logs: lb.String(),
				},
				Errors: err.Error(),
			}
			log.Error("failed to import", "err", err, slog.Any("output", o))
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(o)
			return
		}
		out := Output{
			Meta: MetaOut{
				Length: fmt.Sprintf("%d", resw.Len()),
				Uuid:   msg.Meta.Uuid,
				Digest: []byte(fmt.Sprintf("%x", sha256.Sum256(resw.Bytes()))),
				Logs:   lb.String(),
			},
			Data: resw.String(),
		}
		_, _ = fmt.Fprint(os.Stderr, out.Meta.Logs)
		if err := json.NewEncoder(w).Encode(out); err != nil {
			log.Error("failed to encode response", "err", err)
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(err.Error()))
			return
		}
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

func logReplace(groups []string, a slog.Attr) slog.Attr {
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
