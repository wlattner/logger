package logger

import (
	"log"
	"net"
	"net/http"
	"time"
)

// Handler wrapps an http.Handler with a logger outputting data in the
// following format: remote ip, method, url, status, size, duration
func Handler(fn http.Handler) http.Handler {
	return logger{fn}
}

func (l logger) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	resp := &responseLogger{w: w}
	l.handler.ServeHTTP(resp, r)
	go printLog(r, resp.status, resp.size, time.Since(start))
}

type logger struct {
	handler http.Handler
}

type responseLogger struct {
	w      http.ResponseWriter
	status int
	size   int
}

func (l *responseLogger) Header() http.Header {
	return l.w.Header()
}

func (l *responseLogger) Write(b []byte) (int, error) {
	if l.status == 0 {
		l.status = http.StatusOK
	}
	size, err := l.w.Write(b)
	l.size += size
	return size, err
}

func (l *responseLogger) WriteHeader(s int) {
	l.w.WriteHeader(s)
	l.status = s
}

func printLog(req *http.Request, status int, size int, d time.Duration) {
	host, _, _ := net.SplitHostPort(req.RemoteAddr)
	requestTime := d.Nanoseconds() / 1e6
	// ip method path status size time
	// 0.0.0.0 GET /api/users 200 312 34
	log.Printf("%s %s %s %d %d %d\n",
		host,
		req.Method,
		req.URL.RequestURI(),
		status,
		size,
		requestTime,
	)
}
