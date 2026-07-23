package dailybestsource

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	browseragenttypes "rudy_gc/remote/api/internal/types/browseragent"
)

func TestFlareSolverrClientFetchReusesSession(t *testing.T) {
	var sessionCreates atomic.Int64
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		var input map[string]any
		if err := json.NewDecoder(request.Body).Decode(&input); err != nil {
			t.Fatal(err)
		}
		switch input["cmd"] {
		case "sessions.create":
			sessionCreates.Add(1)
			_ = json.NewEncoder(writer).Encode(map[string]any{"status": "ok", "message": "Session created successfully."})
		case "request.get":
			if input["session"] != "dailybest" || input["maxTimeout"] != float64(time.Minute.Milliseconds()) {
				t.Fatalf("input=%+v", input)
			}
			_ = json.NewEncoder(writer).Encode(map[string]any{
				"status": "ok",
				"solution": map[string]any{
					"status": 200, "response": "<html></html>", "url": input["url"],
				},
			})
		default:
			t.Fatalf("cmd=%v", input["cmd"])
		}
	}))
	defer server.Close()

	client, err := NewFlareSolverrClient(FlareSolverrClientConfig{
		BaseURL: server.URL, Session: "dailybest", Timeout: time.Minute,
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, requestID := range []string{"request-1", "request-2"} {
		response, err := client.Fetch(context.Background(), requestID, "https://www.c97k.com/page", browseragenttypes.PurposeProbe)
		if err != nil {
			t.Fatal(err)
		}
		if response.StatusCode != http.StatusOK || string(response.Body) != "<html></html>" {
			t.Fatalf("response=%+v", response)
		}
	}
	if sessionCreates.Load() != 1 {
		t.Fatalf("sessionCreates=%d", sessionCreates.Load())
	}
}

func TestFlareSolverrClientAcceptsExistingSession(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		var input map[string]any
		if err := json.NewDecoder(request.Body).Decode(&input); err != nil {
			t.Fatal(err)
		}
		if input["cmd"] == "sessions.create" {
			writer.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(writer).Encode(map[string]any{
				"status": "error", "message": "A session with the name 'dailybest' already exists.",
			})
			return
		}
		_ = json.NewEncoder(writer).Encode(map[string]any{
			"status":   "ok",
			"solution": map[string]any{"status": 200, "response": "<html></html>", "url": input["url"]},
		})
	}))
	defer server.Close()

	client, err := NewFlareSolverrClient(FlareSolverrClientConfig{
		BaseURL: server.URL, Session: "dailybest", Timeout: time.Minute,
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := client.Get(context.Background(), "https://www.c97k.com/page"); err != nil {
		t.Fatal(err)
	}
}

func TestFlareSolverrClientStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(writer).Encode(map[string]any{
			"msg": "FlareSolverr is ready!", "version": "3.5.0", "userAgent": "test-agent",
		})
	}))
	defer server.Close()

	client, err := NewFlareSolverrClient(FlareSolverrClientConfig{
		BaseURL: server.URL, Session: "dailybest", Timeout: time.Minute,
	})
	if err != nil {
		t.Fatal(err)
	}
	status, err := client.Status(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if status.State != browseragenttypes.StateReady || status.Version != "3.5.0" || status.UserAgent != "test-agent" {
		t.Fatalf("status=%+v", status)
	}
}

func TestNewRequestID(t *testing.T) {
	first, err := newRequestID()
	if err != nil {
		t.Fatal(err)
	}
	second, err := newRequestID()
	if err != nil {
		t.Fatal(err)
	}
	if len(first) != 32 || first == second {
		t.Fatalf("first=%q second=%q", first, second)
	}
}
