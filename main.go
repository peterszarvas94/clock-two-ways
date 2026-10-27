package main

import (
	"bytes"
	"embed"
	"flag"
	"html/template"
	"log"
	"net/http"
	"time"
	_ "time/tzdata"

	"github.com/starfederation/datastar-go/datastar"
)

//go:embed templates static
var assets embed.FS
var templates = template.Must(template.ParseFS(assets, "templates/index.html"))

type clock struct {
	Epoch                int64
	Hour, Minute, Second int
	Digital, Date, Zone  string
}

func clockAt(now time.Time, zone *time.Location) clock {
	now = now.In(zone)
	return clock{
		Epoch: now.UnixMilli(),
		Hour:  now.Hour() % 12, Minute: now.Minute(), Second: now.Second(),
		Digital: now.Format("15:04:05"), Date: now.Format("Mon, 02 Jan 2006"), Zone: zone.String(),
	}
}

func location(w http.ResponseWriter, r *http.Request) (*time.Location, bool) {
	name := r.URL.Query().Get("tz")
	if name == "" {
		name = "UTC"
	}
	zone, err := time.LoadLocation(name)
	if err != nil {
		http.Error(w, "Invalid timezone", http.StatusBadRequest)
		return nil, false
	}
	return zone, true
}

func handler() http.Handler {
	mux := http.NewServeMux()
	mux.Handle("GET /static/", http.FileServerFS(assets))
	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) { http.Redirect(w, r, "/client", http.StatusFound) })
	for _, path := range []string{"/client", "/stream"} {
		mux.HandleFunc("GET "+path, func(w http.ResponseWriter, r *http.Request) {
			zone, ok := location(w, r)
			if !ok {
				return
			}
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Header().Set("Cache-Control", "no-store")
			data := struct {
				Stream bool
				Clock  clock
			}{r.URL.Path == "/stream", clockAt(time.Now(), zone)}
			if err := templates.ExecuteTemplate(w, "page", data); err != nil {
				log.Print(err)
			}
		})
	}
	mux.HandleFunc("GET /events", func(w http.ResponseWriter, r *http.Request) {
		zone, ok := location(w, r)
		if !ok {
			return
		}
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("X-Accel-Buffering", "no")
		sse := datastar.NewSSE(w, r)
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		for {
			var html bytes.Buffer
			if err := templates.ExecuteTemplate(&html, "face", clockAt(time.Now(), zone)); err != nil {
				log.Print(err)
				return
			}
			if err := sse.PatchElements(html.String()); err != nil {
				return
			}
			select {
			case <-r.Context().Done():
				return
			case <-ticker.C:
			}
		}
	})
	mux.HandleFunc("GET /favicon.ico", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusNoContent) })
	return mux
}

func main() {
	addr := flag.String("addr", "127.0.0.1:8091", "listen address")
	flag.Parse()
	log.Printf("Clock proofs of concept: http://%s", *addr)
	log.Fatal((&http.Server{Addr: *addr, Handler: handler(), ReadHeaderTimeout: 5 * time.Second}).ListenAndServe())
}
