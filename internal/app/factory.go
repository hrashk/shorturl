package app

import "net/http"

func NewInMemoryController() ShortURLController {
	s := NewShortURLService(NewBase62Generator(), NewInMemStorage())

	return NewShortURLController(s)
}

func InMemoryHandler() http.Handler {
	c := NewInMemoryController()

	return http.HandlerFunc(c.RouteRequest)
}
