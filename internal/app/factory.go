package app

func NewInMemoryController() ShortURLController {
	s := NewShortURLService(NewBase62Generator(), NewInMemStorage())

	return NewShortURLController(s)
}
