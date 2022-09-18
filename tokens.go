package main

import (
	"sync"
	"time"
)

type tTokens struct {
	tokens map[string]*tToken
	mutex  sync.RWMutex
}

type tToken struct {
	token  string
	expire time.Time
}

// Add/Put new token to map
func (t *tTokens) add(basic string, token *tToken) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	t.tokens[basic] = token
}

// Get token from map
func (t *tTokens) get(basic string) *tToken {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	return t.tokens[basic]
}

// Check if token is expired and delete it
func (t *tTokens) delExpired() {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	for k, v := range t.tokens {
		if time.Now().After(v.expire) {
			delete(t.tokens, k)
		}
	}
}
