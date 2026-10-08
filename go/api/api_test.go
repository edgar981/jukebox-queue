package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/edgar981/jukebox-queue/go/queue"
)

func do(t *testing.T, h http.Handler, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestHTTPFlow(t *testing.T) {
	h := NewHandler(queue.NewStore(nil))

	if rec := do(t, h, "GET", "/health", ""); rec.Code != 200 {
		t.Fatalf("health: %d", rec.Code)
	}

	rec := do(t, h, "POST", "/venues/bar/queue", `{"songId":"s1","title":"A","tableId":"t1"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("add: %d %s", rec.Code, rec.Body)
	}
	var created queue.Item
	json.Unmarshal(rec.Body.Bytes(), &created)

	if rec := do(t, h, "POST", "/venues/bar/queue", `{"songId":"s1","title":"A","tableId":"t2"}`); rec.Code != http.StatusConflict {
		t.Fatalf("duplicate: want 409, got %d", rec.Code)
	}
	if rec := do(t, h, "POST", "/venues/bar/queue/"+created.ID+"/votes", `{"tableId":"t2"}`); rec.Code != 200 {
		t.Fatalf("vote: %d", rec.Code)
	}
	if rec := do(t, h, "POST", "/venues/bar/queue/"+created.ID+"/votes", `{"tableId":"t2"}`); rec.Code != http.StatusConflict {
		t.Fatalf("second vote: want 409, got %d", rec.Code)
	}
	if rec := do(t, h, "POST", "/venues/bar/queue/advance", ""); rec.Code != 200 {
		t.Fatalf("advance: %d", rec.Code)
	}
	if rec := do(t, h, "POST", "/venues/bar/queue/advance", ""); rec.Code != http.StatusNotFound {
		t.Fatalf("advance empty: want 404, got %d", rec.Code)
	}
}

func TestBadRequests(t *testing.T) {
	h := NewHandler(queue.NewStore(nil))
	if rec := do(t, h, "POST", "/venues/bar/queue", `not json`); rec.Code != 400 {
		t.Fatalf("invalid json: want 400, got %d", rec.Code)
	}
	if rec := do(t, h, "POST", "/venues/bar/queue", `{"songId":"","title":"A","tableId":"t1"}`); rec.Code != 400 {
		t.Fatalf("blank field: want 400, got %d", rec.Code)
	}
}

func TestTableLimitIs429(t *testing.T) {
	h := NewHandler(queue.NewStore(nil))
	do(t, h, "POST", "/venues/bar/queue", `{"songId":"s1","title":"A","tableId":"t1"}`)
	do(t, h, "POST", "/venues/bar/queue", `{"songId":"s2","title":"B","tableId":"t1"}`)
	if rec := do(t, h, "POST", "/venues/bar/queue", `{"songId":"s3","title":"C","tableId":"t1"}`); rec.Code != http.StatusTooManyRequests {
		t.Fatalf("want 429, got %d", rec.Code)
	}
}
