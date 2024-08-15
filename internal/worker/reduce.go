package worker

import (
	"bufio"
	"fmt"

	"github.com/Assifar-Karim/apollo/internal/io"
	"github.com/Assifar-Karim/apollo/internal/proto"
)

type Reducer struct {
}

func (r *Reducer) HandleTask(task *proto.Task) {
	fmt.Println("reducer")
}

func (r *Reducer) FetchInputData(task *proto.Task) (*bufio.Scanner, io.Closeable, error) {
	return nil, nil, nil
}
