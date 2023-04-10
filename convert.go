package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/volvo-cars/lingon/pkg/kube"
	"golang.org/x/exp/slog"
)

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

func convert(log *slog.Logger) http.HandlerFunc {
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

		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
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
