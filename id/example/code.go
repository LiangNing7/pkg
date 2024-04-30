package main

import (
	"fmt"
	"github.com/LiangNing7/onex/pkg/id"
)

func main() {
	fmt.Println(id.NewCode(1))
	fmt.Println(id.NewCode(2))
	fmt.Println(id.NewCode(3))
	fmt.Println(id.NewCode(4))

	fmt.Println(id.NewCode(
		1,
		id.WithCodeChars([]rune{'0', '1', '2', '3', '4', '5', '6', '7', '8', '9'}),
		id.WithCodeN1(9),
		id.WithCodeN2(3),
		id.WithCodeL(5),
		id.WithCodeSalt(99999),
	))
}
