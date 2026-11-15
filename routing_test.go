package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// Buttondown redirects to /newsletter/thanks/?email=…; shared links carry
// ?utm_…: the router must see the path alone, the handler the full query.
func TestPathOnly(t *testing.T) {
	var gotURI, gotEmail string
	h := pathOnly(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotURI, gotEmail = r.RequestURI, r.URL.Query().Get("email")
	}))
	r := httptest.NewRequest("GET", "/newsletter/thanks/?email=a%40b.c", nil)
	h.ServeHTTP(httptest.NewRecorder(), r)
	if gotURI != "/newsletter/thanks/" || gotEmail != "a@b.c" {
		t.Errorf("RequestURI %q, email %q", gotURI, gotEmail)
	}
	r = httptest.NewRequest("GET", "/croquis/c%C3%A9sped/", nil)
	h.ServeHTTP(httptest.NewRecorder(), r)
	if gotURI != "/croquis/c%C3%A9sped/" {
		t.Errorf("path without query changed: %q", gotURI)
	}
}
