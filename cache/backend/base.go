package cachebackend

import "time"

type Backend interface {
	Has(string) (bool, error)
	Get(string) (interface{}, error)
	Set(string, interface{}, time.Duration) error
	Delete(string) error
}
