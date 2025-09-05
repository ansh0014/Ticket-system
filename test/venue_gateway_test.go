package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
	"time"
)

func postJSON(t *testing.T, url string, payload interface{}) (int, []byte) {
	t.Helper()
	b, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Post(url, "application/json", bytes.NewReader(b))
	if err != nil {
		t.Fatalf("POST %s failed: %v", url, err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, body
}

func get(t *testing.T, url string) (int, []byte) {
	t.Helper()
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		t.Fatalf("GET %s failed: %v", url, err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, body
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

// TestVenueCRUDViaGateway creates a venue, a hall and a seat via the API Gateway and verifies reads.
func TestVenueCRUDViaGateway(t *testing.T) {
	gw := envOr("GATEWAY_URL", "http://localhost:8004")

	// 1) Create Venue
	name := fmt.Sprintf("Test Venue %d", time.Now().UnixNano())
	venuePayload := map[string]interface{}{
		"name":    name,
		"address": "123 Test St",
		"city":    "Testville",
	}
	status, body := postJSON(t, gw+"/venue/venues", venuePayload)
	if status != http.StatusCreated {
		t.Fatalf("create venue: expected status 201, got %d, body=%s", status, string(body))
	}
	var createdVenue map[string]interface{}
	if err := json.Unmarshal(body, &createdVenue); err != nil {
		t.Fatalf("unmarshal created venue: %v (body=%s)", err, string(body))
	}
	vid, ok := createdVenue["id"].(string)
	if !ok || vid == "" {
		t.Fatalf("created venue missing id, body=%s", string(body))
	}

	// 2) Get Venue
	status, body = get(t, gw+"/venue/venues/"+vid)
	if status != http.StatusOK {
		t.Fatalf("get venue: expected 200, got %d, body=%s", status, string(body))
	}
	var gotVenue map[string]interface{}
	if err := json.Unmarshal(body, &gotVenue); err != nil {
		t.Fatalf("unmarshal get venue: %v", err)
	}
	if gotVenue["name"] != name {
		t.Fatalf("venue name mismatch: expected %q got %q", name, gotVenue["name"])
	}

	// 3) Create Hall under Venue
	hallPayload := map[string]interface{}{
		"name": "Main Hall",
		"rows": 5,
		"cols": 10,
	}
	status, body = postJSON(t, gw+"/venue/venues/"+vid+"/halls", hallPayload)
	if status != http.StatusCreated {
		t.Fatalf("create hall: expected 201, got %d, body=%s", status, string(body))
	}
	var createdHall map[string]interface{}
	if err := json.Unmarshal(body, &createdHall); err != nil {
		t.Fatalf("unmarshal created hall: %v", err)
	}
	hid, ok := createdHall["id"].(string)
	if !ok || hid == "" {
		t.Fatalf("created hall missing id, body=%s", string(body))
	}

	// 4) Add a Seat to Hall
	seatPayload := map[string]interface{}{
		"number":   "A1",
		"category": "standard",   // you can change to "premium" or "vip"
		"price":    12.5,
	}
	status, body = postJSON(t, gw+"/venue/halls/"+hid+"/seats", seatPayload)
	if status != http.StatusCreated {
		t.Fatalf("create seat: expected 201, got %d, body=%s", status, string(body))
	}
	var createdSeat map[string]interface{}
	if err := json.Unmarshal(body, &createdSeat); err != nil {
		t.Fatalf("unmarshal created seat: %v", err)
	}
	sid, ok := createdSeat["id"].(string)
	if !ok || sid == "" {
		t.Fatalf("created seat missing id, body=%s", string(body))
	}

	// 5) List Seats for Hall
	status, body = get(t, gw+"/venue/halls/"+hid+"/seats")
	if status != http.StatusOK {
		t.Fatalf("list seats: expected 200, got %d, body=%s", status, string(body))
	}
	var seats []map[string]interface{}
	if err := json.Unmarshal(body, &seats); err != nil {
		// some implementations return object with "data" wrapper; try to decode that
		var wrapper map[string]interface{}
		if err2 := json.Unmarshal(body, &wrapper); err2 == nil {
			if d, ok := wrapper["data"]; ok {
				if bs, err3 := json.Marshal(d); err3 == nil {
					if err4 := json.Unmarshal(bs, &seats); err4 == nil {
						// ok
					}
				}
			}
		}
		if len(seats) == 0 {
			t.Fatalf("unmarshal list seats failed: %v (body=%s)", err, string(body))
		}
	}
	// verify created seat present
	found := false
	for _, s := range seats {
		if id, ok := s["id"].(string); ok && id == sid {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("created seat id %s not found in list, body=%s", sid, string(body))
	}
}
