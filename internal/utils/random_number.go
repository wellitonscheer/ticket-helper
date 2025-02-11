package utils

import "math/rand"

func Random6Numbers() int {
	min := 100000
	max := 999999

	return min + rand.Intn(max-min)
}
