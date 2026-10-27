package main

import (
	"bufio"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestTimezoneRendering(t *testing.T) {
	zone, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		t.Fatal(err)
	}
	got := clockAt(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), zone)
	if got.Digital != "09:00:00" || got.Hour != 9 {
		t.Fatalf("wrong timezone: %+v", got)
	}
	for _, path := range []string{"/client?tz=invalid", "/stream?tz=invalid", "/events?tz=invalid"} {
		response := httptest.NewRecorder()
		handler().ServeHTTP(response, httptest.NewRequest("GET", path, nil))
		if response.Code != http.StatusBadRequest {
			t.Fatalf("%s: status %d", path, response.Code)
		}
	}
}

func TestBothPagesRenderInitialClock(t *testing.T) {
	for _, path := range []string{"/client?tz=Asia%2FTokyo", "/stream?tz=Asia%2FTokyo"} {
		response := httptest.NewRecorder()
		handler().ServeHTTP(response, httptest.NewRequest("GET", path, nil))
		body := response.Body.String()
		if response.Code != http.StatusOK {
			t.Fatalf("%s: status %d", path, response.Code)
		}
		for _, want := range []string{`<figure id="face"`, `data-zone="Asia/Tokyo"`, `--hour:`, `--minute:`, `--second:`, `<time id="digital">`} {
			if !strings.Contains(body, want) {
				t.Errorf("%s: initial HTML missing %s", path, want)
			}
		}
		if strings.Contains(body, "Loading clock") || strings.Contains(body, "ZgotmplZ") {
			t.Fatalf("%s: invalid initial clock", path)
		}
	}
}

func TestStreamStaysOpen(t *testing.T) {
	server := httptest.NewServer(handler())
	defer server.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, "GET", server.URL+"/events?tz=UTC", nil)
	if err != nil {
		t.Fatal(err)
	}
	response, err := server.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	if !strings.HasPrefix(response.Header.Get("Content-Type"), "text/event-stream") {
		t.Fatal("not SSE")
	}
	scanner := bufio.NewScanner(response.Body)
	var times []string
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, `<time id="digital">`) {
			times = append(times, line)
			if len(times) == 2 {
				break
			}
		}
	}
	if len(times) != 2 || times[0] == times[1] {
		t.Fatalf("need two different times on one response: %v; error %v", times, scanner.Err())
	}
	cancel()
}
