package session

import "fmt"

type RandFunc func() int

func StaticGenerateCode() int {
	return 123456
}

func GenerateCode(randFn RandFunc) string {
	n := ((randFn() % 1000000) + 1000000) % 1000000

	return fmt.Sprintf("%06d", n)
}
