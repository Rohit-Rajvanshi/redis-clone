package main

import (
	"sync"
	"time"
)

type cfg struct {
		data map[string]*valueBody
		mu sync.RWMutex
	}

type valueBody struct {
	val any
	expiryTime time.Time
}
