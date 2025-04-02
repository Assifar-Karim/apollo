FROM golang:1.21-alpine AS Build

RUN apk add --no-cache make

RUN mkdir -p protoc
RUN cd protoc && wget https://github.com/protocolbuffers/protobuf/releases/download/v27.1/protoc-27.1-linux-x86_64.zip
RUN unzip protoc/protoc-27.1-linux-x86_64.zip
ENV PATH="$PATH:/go/protoc/bin"

RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.36.5
RUN go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.5.1


WORKDIR /apollo
COPY go.mod .
COPY go.sum .

RUN go mod download
RUN go mod verify

COPY cmd/coordinator cmd/coordinator
COPY proto/msg.proto proto/msg.proto
COPY internal internal
COPY Makefile .

RUN make build_coordinator

FROM alpine:3.20

RUN addgroup --gid 4010 apollo && \
    adduser \
    --disabled-password \
    --gecos "" \
    --home /apollo \
    --no-create-home \
    --ingroup apollo \
    --uid 4010 \
    apollo

USER apollo:apollo
WORKDIR /apollo
RUN mkdir -p data
COPY  --chown=apollo:apollo --from=Build /apollo/bin/coordinator coordinator
EXPOSE 4750

ENTRYPOINT ./coordinator $COORDINATOR_OPTS