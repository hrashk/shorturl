package app

type keyGenerator interface {
	Generate(url string) shortKey
}

type shortKey struct {
	uuid     uint64
	shortURL string
}

type base62Generator struct {
	counter chan uint64
}

func newBase62Generator(initial uint64) base62Generator {
	c := make(chan uint64)

	go count(c, initial)

	return base62Generator{c}
}
func count(counter chan uint64, initial uint64) {
	for i := uint64(initial); ; i++ {
		counter <- i
	}
}
func (g base62Generator) Generate(url string) shortKey {
	uuid := <-g.counter

	return shortKey{
		uuid:     uuid,
		shortURL: encode(uuid),
	}
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
