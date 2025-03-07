---
slug: /getting-started
sidebar_position: 3
---
# Getting Started
This page provides a simple way to setup Apollo on a kubernetes cluster
## Cluster Setup
Before starting this guide, make sure that kubectl is installed on your machine and that your kubeconfig points to where your cluster is located.\
Apply the global manifest that can be found on the [deploy/coordinator.yaml](https://github.com/Assifar-Karim/apollo/blob/main/deploy/coordinator.yaml) location.
```bash
kubectl apply -f deploy/coordinator.yaml
```
## Dev Mode Setup
Dev mode requires both a k8s cluster available to the developer and to have go installed on their machine with protoc and gRPC. The go version that needs to be installed can be checked on [go.mod](https://github.com/Assifar-Karim/apollo/blob/main/go.mod) file at the root of the project. On dev mode only the coordinator is ran locally and the workers are started on the associated cluster.
:::note

To install the required gRPC and protoc dependencies check the prerequisites on the following [page](https://grpc.io/docs/languages/go/quickstart/) 

:::
Before trying to start the coordinator, generate the gRPC code using the following command:
```bash
make generate_grpc_code
```
:::tip

The make utility should be installed on the local machine of the developer to simplify the generation process!  

:::
To start the coordinator run the following command:
```bash
go run cmd/coordinator/main.go --trace --dev
```
:::info

Adding --trace to the run command shows the SQL queries performed by the coordinator when communicating the metadata db. 

:::
In case a developer wishes to work on the workers, then dev mode doesn't require any additional config since at its core a worker is a simple gRPC server, the developer can then start a random worker instance directly using the following command:
```bash
go run cmd/worker/main.go
```
:::info

The worker requires the gRPC generated code to function.  

:::
## Apollo Configuration Reference
| Variable        | Description                                                                                  | Default value              |
|-----------------|----------------------------------------------------------------------------------------------|----------------------------|
| ARTIFACTS_PATH  | Path where the physical executable artifacts are stored.                                     | /coordinator/artifacts     |
| SPLIT_SIZE      | Size in bytes that each data chunk will have when the coordinator splits a file for mappers. | 67108864                   |
| KUBECONFIG_PATH | Path where the dev mode kubeconfig file resides.                                             | ~/.kube/config             |
| WORKER_NS       | Kubernetes namespace where the workers are going to be created.                              | apollo-workers             |
| INT_FILES_LOC   | Path where the resulting files of a mapper are stored.                                       | /apollo/intermediate-files |
## Uploading a binary artifact
To upload a binary artifact a user needs to run the following HTTP request:
```bash
curl --location --request PUT 'http://localhost:4750/api/v1/artifacts' \
--form 'program=@"/path/to/binary/executable"'
```
## Starting a job
To start a job a user needs to run the following HTTP request:
```bash
curl --location 'http://localhost:30369/api/v1/jobs' --header 'Content-Type: application/json' --data '{
    "nReducers": 1,
    "inputPath": "http://minio:9000/test/test.txt",
    "inputType": "file/txt",
    "outputPath": "http://minio:9000",
    "useSSL": false,
    "mapperName": "mapper",
    "reducerName": "reducer",
    "inputStorageCredentials": {
        "username": "username",
        "password": "password"
    },
    "outputStorageCredentials": {
        "username": "username",
        "password": "password"
    }
}' | jq .
```
:::note

The localhost:30369 part of the address will be different depending on how the developer did the setup of Apollo. Same for the Object Storage part. On this example we use Minio. However, any S3 compatible object storage can be accessed by Apollo. 

:::
:::info

Starting a job requires both the mapper and reducer artifacts to be uploaded to the coordinator before hand using the artifacts API. 

:::
## Example Word Count Mapper program written in Go
```go
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"time"
)

type KVPairArray struct {
	Pairs []KVPair `json:"pairs"`
}
type KVPair struct {
	Key   any `json:"key"`
	Value any `json:"value"`
}

func main() {
	args := os.Args[1:]
	if len(args) < 2 {
		log.Fatal("not enough args")
	}
	value := args[1]
	words := strings.Split(value, " ")

	pairs := []KVPair{}
	for _, word := range words {
		pairs = append(pairs, KVPair{
			Key:   word,
			Value: 1,
		})
	}
	pairsObj := KVPairArray{
		Pairs: pairs,
	}
	jsonRes, err := json.Marshal(pairsObj)
	if err != nil {
		log.Fatalf("json marshalling error: %v", err.Error())
	}
	// Connect to the unix socket and send the data to it
	fd, err := net.Dial("unix", "/tmp/map.sock")
	defer fd.Close()
	retryCount := 0
	for err != nil && retryCount > 3 {
		fd, err = net.Dial("unix", "/tmp/map.sock")
		log.Printf("couldn't connect to map.sock %v", err.Error())
		time.Sleep(5 * time.Second)
		retryCount++
	}
	if err != nil {
		log.Fatalf(err.Error())
	}
	_, err = fd.Write(jsonRes)
	retryCount = 0
	for err != nil && retryCount > 3 {
		_, err = fd.Write(jsonRes)
		log.Printf("couldn't write result to map.sock %v", err.Error())
		time.Sleep(2 * time.Second)
		retryCount++
	}
	if err != nil {
		log.Fatalf(err.Error())
	} else {
		fmt.Println("Success")
	}

}
```
## Example Word Count Reducer program written in Go
```go
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"time"
)

type KV struct {
	Key   any   `json:"key"`
	Value []any `json:"value"`
}

type KVPair struct {
	Key   any `json:"key"`
	Value any `json:"value"`
}

func main() {
	args := os.Args[1:]
	if len(args) < 1 {
		log.Fatal("not enough args")
	}
	order := args[0]
	socket, err := net.Listen("unix", fmt.Sprintf("/tmp/reduce-input-%v.sock", order))
	if err != nil {
		log.Fatalf("Couldn't connect to the input socket %v", err.Error())
	}
	defer socket.Close()
	fd, err := socket.Accept()
	if err != nil {
		log.Fatalf("Couldn't accept the connection to the input socket %v", err.Error())
	}
	buf := make([]byte, 1024)
	_, err = fd.Read(buf)
	if err != nil {
		log.Fatalf("Couldn't read %v", err.Error())
	}
	buf = bytes.Trim(buf, "\x00")
	var pair KV
	err = json.Unmarshal(buf, &pair)
	if err != nil {
		log.Fatalf("Couldn't unmarshal the json %v", err)
	}
	values := pair.Value
	sum := 0
	for _, value := range values {
		sum += int(value.(float64))
	}
	res := KVPair{
		Key:   pair.Key,
		Value: sum,
	}
	jsonRes, err := json.Marshal(res)
	if err != nil {
		log.Fatalf("json marshalling error: %v", err.Error())
	}
	// Connect to the unix socket and send the data to it
	fd, err = net.Dial("unix", "/tmp/reduce.sock")
	defer fd.Close()
	retryCount := 0
	for err != nil && retryCount < 3 {
		fd, err = net.Dial("unix", "/tmp/reduce.sock")
		log.Printf("couldn't connect to reduce.sock %v", err.Error())
		time.Sleep(5 * time.Second)
		retryCount++
	}
	if err != nil {
		log.Fatalf(err.Error())
	}
	_, err = fd.Write(jsonRes)
	retryCount = 0
	for err != nil && retryCount < 3 {
		_, err = fd.Write(jsonRes)
		log.Printf("couldn't write result to reduce.sock %v", err.Error())
		time.Sleep(2 * time.Second)
		retryCount++
	}
	if err != nil {
		log.Fatalf(err.Error())
	} else {
		fmt.Println("Success")
	}

}
```