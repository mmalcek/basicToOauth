package main

import (
	"encoding/base64"
	"errors"
	"net/http"
	"testing"
	"time"
)

// --- helpers -----------------------------------------------------------------

// resetTokens clears the package-global token cache between cases. It replaces
// only the map (not the whole struct) to avoid copying the embedded mutex.
func resetTokens() { tokensMap.tokens = make(map[string]*tToken) }

// stubAcquire swaps the acquireToken seam and returns a restore func.
func stubAcquire(fn func(string) (*tToken, error)) func() {
	prev := acquireToken
	acquireToken = fn
	return func() { acquireToken = prev }
}

func basicKey(user, pass string) string {
	return base64.StdEncoding.EncodeToString([]byte(user + ":" + pass))
}

func basicHeader(user, pass string) string { return "Basic " + basicKey(user, pass) }

// --- getAuthHeader: passthrough / malformed ----------------------------------

func TestGetAuthHeaderNonBasic(t *testing.T) {
	resetTokens()
	restore := stubAcquire(func(string) (*tToken, error) {
		t.Fatal("acquireToken must not be called for non-Basic input")
		return nil, nil
	})
	defer restore()

	for _, in := range []string{"", "Bearer abc", "Negotiate xyz"} {
		if got := getAuthHeader(in); got != in {
			t.Errorf("getAuthHeader(%q) = %q, want unchanged", in, got)
		}
	}
}

func TestGetAuthHeaderMalformedBasic(t *testing.T) {
	resetTokens()
	called := 0
	restore := stubAcquire(func(string) (*tToken, error) {
		called++
		return nil, nil
	})
	defer restore()

	for _, in := range []string{"Basic", "Basic a b"} {
		if got := getAuthHeader(in); got != in {
			t.Errorf("getAuthHeader(%q) = %q, want unchanged", in, got)
		}
	}
	if called != 0 {
		t.Errorf("acquireToken called %d times, want 0", called)
	}
}

// --- getAuthHeader: cache hit / refresh / miss --------------------------------

func TestGetAuthHeaderCachedValid(t *testing.T) {
	resetTokens()
	key := basicKey("user", "pass")
	tokensMap.add(key, &tToken{token: "cached", expire: time.Now().Add(time.Hour)})

	restore := stubAcquire(func(string) (*tToken, error) {
		t.Fatal("acquireToken must not be called when the cached token is valid")
		return nil, nil
	})
	defer restore()

	if got := getAuthHeader(basicHeader("user", "pass")); got != "Bearer cached" {
		t.Errorf("got %q, want %q", got, "Bearer cached")
	}
}

func TestGetAuthHeaderNearExpiryRefresh(t *testing.T) {
	resetTokens()
	key := basicKey("user", "pass")
	// expire within the 60s refresh window -> should trigger a refresh.
	tokensMap.add(key, &tToken{token: "stale", expire: time.Now().Add(30 * time.Second)})

	called := 0
	restore := stubAcquire(func(k string) (*tToken, error) {
		called++
		if k != key {
			t.Errorf("acquireToken got key %q, want %q", k, key)
		}
		return &tToken{token: "fresh", expire: time.Now().Add(time.Hour)}, nil
	})
	defer restore()

	if got := getAuthHeader(basicHeader("user", "pass")); got != "Bearer fresh" {
		t.Errorf("got %q, want %q", got, "Bearer fresh")
	}
	if called != 1 {
		t.Errorf("acquireToken called %d times, want 1", called)
	}
	if tok := tokensMap.get(key); tok == nil || tok.token != "fresh" {
		t.Errorf("cache not updated to fresh token: %+v", tok)
	}
}

func TestGetAuthHeaderCacheMissAcquire(t *testing.T) {
	resetTokens()
	key := basicKey("user", "pass")
	called := 0
	restore := stubAcquire(func(string) (*tToken, error) {
		called++
		return &tToken{token: "new", expire: time.Now().Add(time.Hour)}, nil
	})
	defer restore()

	if got := getAuthHeader(basicHeader("user", "pass")); got != "Bearer new" {
		t.Errorf("got %q, want %q", got, "Bearer new")
	}
	if called != 1 {
		t.Errorf("acquireToken called %d times, want 1", called)
	}
	if tok := tokensMap.get(key); tok == nil || tok.token != "new" {
		t.Errorf("cache not populated on miss: %+v", tok)
	}
}

// --- getAuthHeader: acquire failure falls back to the original header ---------
// These also exercise the nil-logger safety: the error path calls
// logger.Warningf, which must NOT panic (default nopLogger).

func TestGetAuthHeaderAcquireErrorMiss(t *testing.T) {
	resetTokens()
	key := basicKey("user", "pass")
	hdr := basicHeader("user", "pass")
	restore := stubAcquire(func(string) (*tToken, error) {
		return nil, errors.New("boom")
	})
	defer restore()

	if got := getAuthHeader(hdr); got != hdr {
		t.Errorf("got %q, want original %q", got, hdr)
	}
	if tokensMap.get(key) != nil {
		t.Errorf("cache should remain empty after acquire error")
	}
}

func TestGetAuthHeaderAcquireErrorRefresh(t *testing.T) {
	resetTokens()
	key := basicKey("user", "pass")
	tokensMap.add(key, &tToken{token: "stale", expire: time.Now().Add(30 * time.Second)})
	restore := stubAcquire(func(string) (*tToken, error) {
		return nil, errors.New("boom")
	})
	defer restore()

	hdr := basicHeader("user", "pass")
	if got := getAuthHeader(hdr); got != hdr {
		t.Errorf("got %q, want original %q", got, hdr)
	}
	if tok := tokensMap.get(key); tok == nil || tok.token != "stale" {
		t.Errorf("stale token should remain after failed refresh: %+v", tok)
	}
}

// --- injectBasicChallenge -----------------------------------------------------

func TestInjectBasicChallenge(t *testing.T) {
	cases := []struct {
		name    string
		status  int
		auth    string
		wantHdr string
	}{
		{"401 no auth sets challenge", http.StatusUnauthorized, "", basicAuthChallenge},
		{"401 with auth leaves none", http.StatusUnauthorized, "Basic xxx", ""},
		{"200 no auth leaves none", http.StatusOK, "", ""},
		{"403 no auth leaves none", http.StatusForbidden, "", ""},
		{"500 no auth leaves none", http.StatusInternalServerError, "", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			resp := &http.Response{StatusCode: tc.status, Header: make(http.Header)}
			injectBasicChallenge(resp, tc.auth)
			if got := resp.Header.Get("WWW-Authenticate"); got != tc.wantHdr {
				t.Errorf("WWW-Authenticate = %q, want %q", got, tc.wantHdr)
			}
		})
	}
}
