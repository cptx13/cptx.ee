package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// Buttondown redirects to /newsletter/thanks/?email=…; shared links carry
// ?utm_…: the router must see the path alone (it matches r.URL.RequestURI()).
func TestPathOnly(t *testing.T) {
	var gotURI, gotURLURI string
	h := pathOnly(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotURI, gotURLURI = r.RequestURI, r.URL.RequestURI()
	}))
	r := httptest.NewRequest("GET", "/newsletter/thanks/?email=a%40b.c", nil)
	h.ServeHTTP(httptest.NewRecorder(), r)
	if gotURI != "/newsletter/thanks/" || gotURLURI != "/newsletter/thanks/" {
		t.Errorf("RequestURI %q, URL.RequestURI %q", gotURI, gotURLURI)
	}
	if r.URL.RawQuery == "" {
		t.Errorf("the original request lost its query")
	}
	r = httptest.NewRequest("GET", "/croquis/c%C3%A9sped/", nil)
	h.ServeHTTP(httptest.NewRecorder(), r)
	if gotURI != "/croquis/c%C3%A9sped/" {
		t.Errorf("path without query changed: %q", gotURI)
	}
}
