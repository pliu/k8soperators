package middleware

import (
	"github.com/go-logr/logr"
	"github.com/gorilla/handlers"
	"io"
	"net"
	"net/http"
	"net/url"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"strconv"
	"time"
)

type LogWriter struct {
	log logr.Logger
}

var logWriter = LogWriter{log: logf.Log.WithName("LoggingMiddleware")}

func (lw LogWriter) Write(p []byte) (n int, err error) {
	lw.log.Info(string(p))
	return len(p), nil
}

func buildLogLine(req *http.Request, url url.URL, ts time.Time, status int, size int) []byte {
	host, _, err := net.SplitHostPort(req.RemoteAddr)

	if err != nil {
		host = req.RemoteAddr
	}

	uri := req.RequestURI

	// Requests using the CONNECT method over HTTP/2.0 must use
	// the authority field (aka r.Host) to identify the target.
	// Refer: https://httpwg.github.io/specs/rfc7540.html#CONNECT
	if req.ProtoMajor == 2 && req.Method == "CONNECT" {
		uri = req.Host
	}
	if uri == "" {
		uri = url.RequestURI()
	}

	buf := make([]byte, 0, 3*(len(host)+len(req.Method)+len(uri)+len(req.Proto)+50)/2)
	buf = append(buf, host...)
	buf = append(buf, " | "...)
	buf = append(buf, req.Method...)
	buf = append(buf, " | "...)
	buf = append(buf, uri...)
	buf = append(buf, " | "...)
	buf = append(buf, req.Proto...)
	buf = append(buf, " | "...)
	buf = append(buf, strconv.Itoa(status)...)
	buf = append(buf, " | "...)
	buf = append(buf, strconv.Itoa(size)...)
	return buf
}

func writeLog(writer io.Writer, params handlers.LogFormatterParams) {
	buf := buildLogLine(params.Request, params.URL, params.TimeStamp, params.StatusCode, params.Size)
	writer.Write(buf)
}

func loggingMiddleware(h http.Handler) http.Handler {
	return handlers.CustomLoggingHandler(logWriter, h, writeLog)
}
