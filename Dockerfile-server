FROM golang:1.12-stretch as build
WORKDIR $GOPATH/src/github.com/cachecashproject/go-cachecash

COPY go.mod go.sum ./
RUN GO111MODULE=on go mod download

COPY . .
RUN make PREFIX=/artifacts all

FROM debian:stretch
RUN apt-get update && apt-get install -y ca-certificates
COPY --from=build /artifacts/bin/* /usr/local/bin/
ENTRYPOINT ["update-server"]
