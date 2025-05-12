package app

func NewInMemoryController() ShortURLController {
	return NewShortURLController(NewBase62Generator(), NewInMemStorage())
}
