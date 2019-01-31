package api

type Server interface {
	Start() error
	Stop() error
}
