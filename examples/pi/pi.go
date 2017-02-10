package main

import "fmt"

func main() {
	c := 100000000
	pi := 0.0
	n := 1.0
	for i := 0; i < c; i++ {
		pi += (4.0 / n) - (4.0 / (n + 2))
		n += 4.0
	}
	fmt.Println(pi)
}
