package main

import (
	"context"
	"fmt"
	"github.com/LiangNing7/onex/pkg/id"
)

func main() {
	sf := id.NewSonyflake(
		id.WithSonyflakeMachineId(1),
	)
	if sf.Error != nil {
		fmt.Println(sf.Error)
		return
	}
	fmt.Println("生成是唯一ID为:", sf.Id(context.Background()))
}
