package main

import (
	"sync"
	"testing"
	"time"
)

func TestTokensAddGet(t *testing.T) {
	tk := &tTokens{tokens: make(map[string]*tToken)}
	tk.add("key", &tToken{token: "abc", expire: time.Now().Add(time.Hour)})

	got := tk.get("key")
	if got == nil || got.token != "abc" {
		t.Fatalf("get(\"key\") = %+v, want token \"abc\"", got)
	}
	if tk.get("missing") != nil {
		t.Errorf("get(\"missing\") = non-nil, want nil")
	}
}

func TestTokensGetEmpty(t *testing.T) {
	tk := &tTokens{tokens: make(map[string]*tToken)}
	if tk.get("anything") != nil {
		t.Errorf("get on empty map = non-nil, want nil")
	}
}

func TestTokensDelExpired(t *testing.T) {
	tk := &tTokens{tokens: make(map[string]*tToken)}
	tk.add("past", &tToken{token: "old", expire: time.Now().Add(-time.Second)})
	tk.add("future", &tToken{token: "new", expire: time.Now().Add(time.Hour)})

	tk.delExpired()

	if tk.get("past") != nil {
		t.Errorf("expired token was not deleted")
	}
	if tk.get("future") == nil {
		t.Errorf("unexpired token was deleted")
	}
}

// TestTokensConcurrent is a race-detector smoke test for the RWMutex; run with
// `go test -race`. It asserts no panic / data race under concurrent access.
func TestTokensConcurrent(t *testing.T) {
	tk := &tTokens{tokens: make(map[string]*tToken)}
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			tk.add("k", &tToken{token: "t", expire: time.Now().Add(time.Hour)})
			_ = tk.get("k")
			tk.delExpired()
		}()
	}
	wg.Wait()
}
