package app

type Base62Generator struct {
	counter chan uint64
}

func NewBase62Generator() *Base62Generator {
	c := make(chan uint64)

	go count(c)

	return &Base62Generator{c}
}
func count(counter chan uint64) {
	for i := uint64(1); ; i++ {
		counter <- i
	}
}
func (g *Base62Generator) Generate(url string) (key string) {
	num := <-g.counter
	key = encode(num)

	return key
}
func encode(num uint64) string {
	const base62 = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const base = uint64(len(base62))
	var result string
	for num > 0 {
		result = string(base62[num%base]) + result
		num /= base
	}
	// Pad with leading zeros to ensure a fixed length
	for len(result) < 6 {
		result = string(base62[0]) + result
	}

	return result
}
