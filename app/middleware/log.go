package middleware

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"
)

type reqInfo struct {
	http.ResponseWriter
	Request *http.Request
	Time    time.Time
	Elapsed time.Duration
	Method  string
	Status  int
}

// Logging simply wraps gorilla logging middleware to be used in the mux .Use()
func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		info := &reqInfo{
			ResponseWriter: w,
			Request:        r,
			Time:           time.Now(),
			Status:         http.StatusOK,
		}
		if r.Header.Get("Upgrade") == "websocket" {
			log.Println("[WEBSOCKET CONN]", info.Request.URL.String())
			next.ServeHTTP(info, r)
		} else {
			next.ServeHTTP(info, r)
			info.Elapsed = time.Since(info.Time)
			log.Println(info.String())
		}
	})
}

func (info *reqInfo) WriteHeader(statusCode int) {
	info.Status = statusCode
	info.ResponseWriter.WriteHeader(statusCode)
}

func (info *reqInfo) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	h, ok := info.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, errors.New("hijack not supported")
	}
	info.Status = http.StatusSwitchingProtocols
	return h.Hijack()
}

func (info *reqInfo) String() string {
	return fmt.Sprintf(
		"%v %v %v %v",
		info.Status,
		info.Request.Method,
		info.Request.URL.String(),
		info.Elapsed,
	)
}
