package worker

import (
	"fmt"

	"github.com/Assifar-Karim/apollo/internal/proto"
)

type Mapper struct {
}

func (m Mapper) HandleTask(task *proto.Task) {
	fmt.Println("mapper")
}
