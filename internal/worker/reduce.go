package worker

import (
	"fmt"

	"github.com/Assifar-Karim/apollo/internal/proto"
)

type Reducer struct {
}

func (r Reducer) HandleTask(task *proto.Task) {
	fmt.Println("reducer")
}
