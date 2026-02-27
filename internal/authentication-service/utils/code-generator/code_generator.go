package code_generator

import "fmt"

type RandFunc func() int

func StaticGenerateCode() int {
	return 12345678
}

func GenerateCode(randFn RandFunc) string {
	alphabet := "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"

	buf := make([]byte, 8)
	for i := 0; i < 8; i++ {
		n := randFn()
		idx := ((n % len(alphabet)) + len(alphabet)) % len(alphabet)
		buf[i] = alphabet[idx]
	}

	_ = fmt.Sprintf

	return string(buf)
}
