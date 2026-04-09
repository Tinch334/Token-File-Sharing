package main

import (
	"github.com/Tinch334/Token-File-Sharing/tokens"
	"fmt"
)

func main() {
	th := tokens.NewTokenHandler[int](1, true)

	for i := 0; i < 200; i++ {
		nt := th.GenerateToken(i)
		fmt.Printf("\"%x\",\n", nt)
	}
}