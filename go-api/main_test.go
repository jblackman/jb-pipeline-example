package main

import (
	"io/ioutil"
	"net/http/httptest"
	"os"
	"regexp"
	"testing"
)

func doGet(t *testing.T, queryPath string, expectStatusCode int, expectContentType, expectBody string) {
	req := httptest.NewRequest("GET", "http://foo.localhost"+queryPath, nil)
	w := httptest.NewRecorder()
	handler().ServeHTTP(w, req)
	resp := w.Result()
	statusCode := resp.StatusCode
	contentType := resp.Header.Get("Content-Type")
	bodyBytes, _ := ioutil.ReadAll(resp.Body)
	body := string(bodyBytes)

	if statusCode != expectStatusCode {
		t.Fatalf("GET %s, expected code %d, got %d", queryPath, expectStatusCode, statusCode)
	}

	if contentType != expectContentType {
		t.Fatalf("GET %s, expected content-type %s, got %s", queryPath, expectContentType, contentType)
	}

	if matched, err := regexp.MatchString(expectBody, body); err != nil {
		t.Fatalf("Regexp failure for /%s/: %v", body, err)
	} else if !matched {
		t.Fatalf("GET %s, body did not contain /%s/ body:\n%s", queryPath, expectBody, body)
	}
}

func TestIndexPageNoGreeting(t *testing.T) {
	os.Setenv("GREETING", "")
	doGet(t, "/", 200, "text/plain; charset=utf-8", "Hello world")
}

func TestIndexPageWithGreeting(t *testing.T) {
	os.Setenv("GREETING", "foobar")
	doGet(t, "/", 200, "text/plain; charset=utf-8", "foobar")
}

func TestGreetingPageNotSet(t *testing.T) {
	os.Setenv("GREETING", "")
	doGet(t, "/greeting", 200, "application/json", `"greeting":"not set"`)
}

func TestGreetingPageWithGreeting(t *testing.T) {
	os.Setenv("GREETING", "barfoo")
	doGet(t, "/greeting", 200, "application/json", `"greeting":"barfoo"`)
}
