package worker

import (
	"bufio"
	"fmt"
	"time"

	"github.com/Assifar-Karim/apollo/internal/io"
	"github.com/Assifar-Karim/apollo/internal/proto"
)

type WorkerAlgorithm interface {
	FetchInputData(task *proto.Task) ([]*bufio.Scanner, []io.Closeable, error)
	HandleTask(task *proto.Task)
}

type Worker struct {
	workerAlgorithm WorkerAlgorithm
}

func (w *Worker) SetWorkerAlgorithm(algorithm WorkerAlgorithm) {
	w.workerAlgorithm = algorithm
}

// NOTE (KARIM): A worker will generally be subdivided into 3 steps
// STEP 1: Get the input data (object storage for a mapper, locally mounted pv for a reducer)
// STEP 2: Perform the worker task (map or reduce)
// STEP 3: Send the generated data into the sinks (locally mounted pv for a mapper, object storage for a reducer)

func (w Worker) TestWorkerType(task *proto.Task) error {
	w.workerAlgorithm.HandleTask(task)
	scanners, closeables, err := w.workerAlgorithm.FetchInputData(task)
	if err != nil {
		return err
	}

	for _, closeable := range closeables {
		defer closeable.Close()
	}

	for _, scanner := range scanners {
		for scanner.Scan() {
			line := scanner.Text()
			fmt.Println(line)
		}
	}

	time.Sleep(5 * time.Second)
	return nil
}
