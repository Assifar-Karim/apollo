"use strict";(self.webpackChunkdocs=self.webpackChunkdocs||[]).push([[924],{7161:(e,t,n)=>{n.r(t),n.d(t,{assets:()=>l,contentTitle:()=>s,default:()=>h,frontMatter:()=>i,metadata:()=>r,toc:()=>d});const r=JSON.parse('{"id":"getting-started","title":"Getting Started","description":"This page provides a simple way to setup Apollo on a kubernetes cluster","source":"@site/docs/getting-started.md","sourceDirName":".","slug":"/getting-started","permalink":"/getting-started","draft":false,"unlisted":false,"tags":[],"version":"current","sidebarPosition":3,"frontMatter":{"slug":"/getting-started","sidebar_position":3},"sidebar":"defaultSidebar","previous":{"title":"Overview","permalink":"/overview"},"next":{"title":"REST API Reference","permalink":"/api"}}');var o=n(4848),a=n(8453);const i={slug:"/getting-started",sidebar_position:3},s="Getting Started",l={},d=[{value:"Cluster Setup",id:"cluster-setup",level:2},{value:"Dev Mode Setup",id:"dev-mode-setup",level:2},{value:"Apollo Configuration Reference",id:"apollo-configuration-reference",level:2},{value:"Uploading a binary artifact",id:"uploading-a-binary-artifact",level:2},{value:"Starting a job",id:"starting-a-job",level:2},{value:"Example Word Count Mapper program written in Go",id:"example-word-count-mapper-program-written-in-go",level:2},{value:"Example Word Count Reducer program written in Go",id:"example-word-count-reducer-program-written-in-go",level:2}];function c(e){const t={a:"a",admonition:"admonition",br:"br",code:"code",h1:"h1",h2:"h2",header:"header",p:"p",pre:"pre",table:"table",tbody:"tbody",td:"td",th:"th",thead:"thead",tr:"tr",...(0,a.R)(),...e.components};return(0,o.jsxs)(o.Fragment,{children:[(0,o.jsx)(t.header,{children:(0,o.jsx)(t.h1,{id:"getting-started",children:"Getting Started"})}),"\n",(0,o.jsx)(t.p,{children:"This page provides a simple way to setup Apollo on a kubernetes cluster"}),"\n",(0,o.jsx)(t.h2,{id:"cluster-setup",children:"Cluster Setup"}),"\n",(0,o.jsxs)(t.p,{children:["Before starting this guide, make sure that kubectl is installed on your machine and that your kubeconfig points to where your cluster is located.",(0,o.jsx)(t.br,{}),"\n","Apply the global manifest that can be found on the ",(0,o.jsx)(t.a,{href:"https://github.com/Assifar-Karim/apollo/blob/main/deploy/coordinator.yaml",children:"deploy/coordinator.yaml"})," location."]}),"\n",(0,o.jsx)(t.pre,{children:(0,o.jsx)(t.code,{className:"language-bash",children:"kubectl apply -f deploy/coordinator.yaml\n"})}),"\n",(0,o.jsx)(t.h2,{id:"dev-mode-setup",children:"Dev Mode Setup"}),"\n",(0,o.jsxs)(t.p,{children:["Dev mode requires both a k8s cluster available to the developer and to have go installed on their machine with protoc and gRPC. The go version that needs to be installed can be checked on ",(0,o.jsx)(t.a,{href:"https://github.com/Assifar-Karim/apollo/blob/main/go.mod",children:"go.mod"})," file at the root of the project. On dev mode only the coordinator is ran locally and the workers are started on the associated cluster."]}),"\n",(0,o.jsx)(t.admonition,{type:"note",children:(0,o.jsxs)(t.p,{children:["To install the required gRPC and protoc dependencies check the prerequisites on the following ",(0,o.jsx)(t.a,{href:"https://grpc.io/docs/languages/go/quickstart/",children:"page"})]})}),"\n",(0,o.jsx)(t.p,{children:"Before trying to start the coordinator, generate the gRPC code using the following command:"}),"\n",(0,o.jsx)(t.pre,{children:(0,o.jsx)(t.code,{className:"language-bash",children:"make generate_grpc_code\n"})}),"\n",(0,o.jsx)(t.admonition,{type:"tip",children:(0,o.jsx)(t.p,{children:"The make utility should be installed on the local machine of the developer to simplify the generation process!"})}),"\n",(0,o.jsx)(t.p,{children:"To start the coordinator run the following command:"}),"\n",(0,o.jsx)(t.pre,{children:(0,o.jsx)(t.code,{className:"language-bash",children:"go run cmd/coordinator/main.go --trace --dev\n"})}),"\n",(0,o.jsx)(t.admonition,{type:"info",children:(0,o.jsx)(t.p,{children:"Adding --trace to the run command shows the SQL queries performed by the coordinator when communicating the metadata db."})}),"\n",(0,o.jsx)(t.p,{children:"In case a developer wishes to work on the workers, then dev mode doesn't require any additional config since at its core a worker is a simple gRPC server, the developer can then start a random worker instance directly using the following command:"}),"\n",(0,o.jsx)(t.pre,{children:(0,o.jsx)(t.code,{className:"language-bash",children:"go run cmd/worker/main.go\n"})}),"\n",(0,o.jsx)(t.admonition,{type:"info",children:(0,o.jsx)(t.p,{children:"The worker requires the gRPC generated code to function."})}),"\n",(0,o.jsx)(t.h2,{id:"apollo-configuration-reference",children:"Apollo Configuration Reference"}),"\n",(0,o.jsxs)(t.table,{children:[(0,o.jsx)(t.thead,{children:(0,o.jsxs)(t.tr,{children:[(0,o.jsx)(t.th,{children:"Variable"}),(0,o.jsx)(t.th,{children:"Description"}),(0,o.jsx)(t.th,{children:"Default value"})]})}),(0,o.jsxs)(t.tbody,{children:[(0,o.jsxs)(t.tr,{children:[(0,o.jsx)(t.td,{children:"ARTIFACTS_PATH"}),(0,o.jsx)(t.td,{children:"Path where the physical executable artifacts are stored."}),(0,o.jsx)(t.td,{children:"/coordinator/artifacts"})]}),(0,o.jsxs)(t.tr,{children:[(0,o.jsx)(t.td,{children:"SPLIT_SIZE"}),(0,o.jsx)(t.td,{children:"Size in bytes that each data chunk will have when the coordinator splits a file for mappers."}),(0,o.jsx)(t.td,{children:"67108864"})]}),(0,o.jsxs)(t.tr,{children:[(0,o.jsx)(t.td,{children:"KUBECONFIG_PATH"}),(0,o.jsx)(t.td,{children:"Path where the dev mode kubeconfig file resides."}),(0,o.jsx)(t.td,{children:"~/.kube/config"})]}),(0,o.jsxs)(t.tr,{children:[(0,o.jsx)(t.td,{children:"WORKER_NS"}),(0,o.jsx)(t.td,{children:"Kubernetes namespace where the workers are going to be created."}),(0,o.jsx)(t.td,{children:"apollo-workers"})]}),(0,o.jsxs)(t.tr,{children:[(0,o.jsx)(t.td,{children:"INT_FILES_LOC"}),(0,o.jsx)(t.td,{children:"Path where the resulting files of a mapper are stored."}),(0,o.jsx)(t.td,{children:"/apollo/intermediate-files"})]})]})]}),"\n",(0,o.jsx)(t.h2,{id:"uploading-a-binary-artifact",children:"Uploading a binary artifact"}),"\n",(0,o.jsx)(t.p,{children:"To upload a binary artifact a user needs to run the following HTTP request:"}),"\n",(0,o.jsx)(t.pre,{children:(0,o.jsx)(t.code,{className:"language-bash",children:"curl --location --request PUT 'http://localhost:4750/api/v1/artifacts' \\\n--form 'program=@\"/path/to/binary/executable\"'\n"})}),"\n",(0,o.jsx)(t.h2,{id:"starting-a-job",children:"Starting a job"}),"\n",(0,o.jsx)(t.p,{children:"To start a job a user needs to run the following HTTP request:"}),"\n",(0,o.jsx)(t.pre,{children:(0,o.jsx)(t.code,{className:"language-bash",children:'curl --location \'http://localhost:30369/api/v1/jobs\' --header \'Content-Type: application/json\' --data \'{\n    "nReducers": 1,\n    "inputPath": "http://minio:9000/test/test.txt",\n    "inputType": "file/txt",\n    "outputPath": "http://minio:9000",\n    "useSSL": false,\n    "mapperName": "mapper",\n    "reducerName": "reducer",\n    "inputStorageCredentials": {\n        "username": "username",\n        "password": "password"\n    },\n    "outputStorageCredentials": {\n        "username": "username",\n        "password": "password"\n    }\n}\' | jq .\n'})}),"\n",(0,o.jsx)(t.admonition,{type:"note",children:(0,o.jsx)(t.p,{children:"The localhost:30369 part of the address will be different depending on how the developer did the setup of Apollo. Same for the Object Storage part. On this example we use Minio. However, any S3 compatible object storage can be accessed by Apollo."})}),"\n",(0,o.jsx)(t.admonition,{type:"info",children:(0,o.jsx)(t.p,{children:"Starting a job requires both the mapper and reducer artifacts to be uploaded to the coordinator before hand using the artifacts API."})}),"\n",(0,o.jsx)(t.h2,{id:"example-word-count-mapper-program-written-in-go",children:"Example Word Count Mapper program written in Go"}),"\n",(0,o.jsx)(t.pre,{children:(0,o.jsx)(t.code,{className:"language-go",children:'package main\n\nimport (\n\t"encoding/json"\n\t"fmt"\n\t"log"\n\t"net"\n\t"os"\n\t"strings"\n\t"time"\n)\n\ntype KVPairArray struct {\n\tPairs []KVPair `json:"pairs"`\n}\ntype KVPair struct {\n\tKey   any `json:"key"`\n\tValue any `json:"value"`\n}\n\nfunc main() {\n\targs := os.Args[1:]\n\tif len(args) < 2 {\n\t\tlog.Fatal("not enough args")\n\t}\n\tvalue := args[1]\n\twords := strings.Split(value, " ")\n\n\tpairs := []KVPair{}\n\tfor _, word := range words {\n\t\tpairs = append(pairs, KVPair{\n\t\t\tKey:   word,\n\t\t\tValue: 1,\n\t\t})\n\t}\n\tpairsObj := KVPairArray{\n\t\tPairs: pairs,\n\t}\n\tjsonRes, err := json.Marshal(pairsObj)\n\tif err != nil {\n\t\tlog.Fatalf("json marshalling error: %v", err.Error())\n\t}\n\t// Connect to the unix socket and send the data to it\n\tfd, err := net.Dial("unix", "/tmp/map.sock")\n\tdefer fd.Close()\n\tretryCount := 0\n\tfor err != nil && retryCount > 3 {\n\t\tfd, err = net.Dial("unix", "/tmp/map.sock")\n\t\tlog.Printf("couldn\'t connect to map.sock %v", err.Error())\n\t\ttime.Sleep(5 * time.Second)\n\t\tretryCount++\n\t}\n\tif err != nil {\n\t\tlog.Fatalf(err.Error())\n\t}\n\t_, err = fd.Write(jsonRes)\n\tretryCount = 0\n\tfor err != nil && retryCount > 3 {\n\t\t_, err = fd.Write(jsonRes)\n\t\tlog.Printf("couldn\'t write result to map.sock %v", err.Error())\n\t\ttime.Sleep(2 * time.Second)\n\t\tretryCount++\n\t}\n\tif err != nil {\n\t\tlog.Fatalf(err.Error())\n\t} else {\n\t\tfmt.Println("Success")\n\t}\n\n}\n'})}),"\n",(0,o.jsx)(t.h2,{id:"example-word-count-reducer-program-written-in-go",children:"Example Word Count Reducer program written in Go"}),"\n",(0,o.jsx)(t.pre,{children:(0,o.jsx)(t.code,{className:"language-go",children:'package main\n\nimport (\n\t"bytes"\n\t"encoding/json"\n\t"fmt"\n\t"log"\n\t"net"\n\t"os"\n\t"time"\n)\n\ntype KV struct {\n\tKey   any   `json:"key"`\n\tValue []any `json:"value"`\n}\n\ntype KVPair struct {\n\tKey   any `json:"key"`\n\tValue any `json:"value"`\n}\n\nfunc main() {\n\targs := os.Args[1:]\n\tif len(args) < 1 {\n\t\tlog.Fatal("not enough args")\n\t}\n\torder := args[0]\n\tsocket, err := net.Listen("unix", fmt.Sprintf("/tmp/reduce-input-%v.sock", order))\n\tif err != nil {\n\t\tlog.Fatalf("Couldn\'t connect to the input socket %v", err.Error())\n\t}\n\tdefer socket.Close()\n\tfd, err := socket.Accept()\n\tif err != nil {\n\t\tlog.Fatalf("Couldn\'t accept the connection to the input socket %v", err.Error())\n\t}\n\tbuf := make([]byte, 1024)\n\t_, err = fd.Read(buf)\n\tif err != nil {\n\t\tlog.Fatalf("Couldn\'t read %v", err.Error())\n\t}\n\tbuf = bytes.Trim(buf, "\\x00")\n\tvar pair KV\n\terr = json.Unmarshal(buf, &pair)\n\tif err != nil {\n\t\tlog.Fatalf("Couldn\'t unmarshal the json %v", err)\n\t}\n\tvalues := pair.Value\n\tsum := 0\n\tfor _, value := range values {\n\t\tsum += int(value.(float64))\n\t}\n\tres := KVPair{\n\t\tKey:   pair.Key,\n\t\tValue: sum,\n\t}\n\tjsonRes, err := json.Marshal(res)\n\tif err != nil {\n\t\tlog.Fatalf("json marshalling error: %v", err.Error())\n\t}\n\t// Connect to the unix socket and send the data to it\n\tfd, err = net.Dial("unix", "/tmp/reduce.sock")\n\tdefer fd.Close()\n\tretryCount := 0\n\tfor err != nil && retryCount < 3 {\n\t\tfd, err = net.Dial("unix", "/tmp/reduce.sock")\n\t\tlog.Printf("couldn\'t connect to reduce.sock %v", err.Error())\n\t\ttime.Sleep(5 * time.Second)\n\t\tretryCount++\n\t}\n\tif err != nil {\n\t\tlog.Fatalf(err.Error())\n\t}\n\t_, err = fd.Write(jsonRes)\n\tretryCount = 0\n\tfor err != nil && retryCount < 3 {\n\t\t_, err = fd.Write(jsonRes)\n\t\tlog.Printf("couldn\'t write result to reduce.sock %v", err.Error())\n\t\ttime.Sleep(2 * time.Second)\n\t\tretryCount++\n\t}\n\tif err != nil {\n\t\tlog.Fatalf(err.Error())\n\t} else {\n\t\tfmt.Println("Success")\n\t}\n\n}\n'})})]})}function h(e={}){const{wrapper:t}={...(0,a.R)(),...e.components};return t?(0,o.jsx)(t,{...e,children:(0,o.jsx)(c,{...e})}):c(e)}},8453:(e,t,n)=>{n.d(t,{R:()=>i,x:()=>s});var r=n(6540);const o={},a=r.createContext(o);function i(e){const t=r.useContext(a);return r.useMemo((function(){return"function"==typeof e?e(t):{...t,...e}}),[t,e])}function s(e){let t;return t=e.disableParentContext?"function"==typeof e.components?e.components(o):e.components||o:i(e.components),r.createElement(a.Provider,{value:t},e.children)}}}]);