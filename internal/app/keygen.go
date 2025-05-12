package app

import "fmt"

type Base62Generator struct {
}

func NewBase62Generator() *Base62Generator {
	return &Base62Generator{}
}

func (g *Base62Generator) Generate(url string) (key string) {
	fmt.Println("Generating key for URL:", url)

	return "EwHXdJfB"
}
