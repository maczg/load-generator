FROM ubuntu:18.04

WORKDIR /tmp/
ADD http://download.tsi.telecom-paristech.fr/gpac/release/1.0.1/gpac_1.0.1-rev0-gd8538e8a-master_amd64.deb ./
RUN apt-get update && apt-get install -y ./gpac*.deb && apt-get clean && rm -rf ./gpac*.deb

RUN apt-get update && apt-get install -y git make gcc tar rsync curl wget vim golang golang-go && apt-get clean

ENV GO111MODULE on
ARG MODULE_PATH=load-generator
ARG MODULE_NAME=load-generator
ENV MODULE_NAME=$MODULE_NAME
ENV MODULE_PATH=$MODULE_PATH
ENV GOPATH=/go
WORKDIR /$GOPATH/src/$MODULE_PATH

# Live-Reload Go project
RUN go get github.com/cespare/reflex

# Add utils scripts for the container
COPY docker/root/ /

COPY go.* ./
RUN go mod download

ENTRYPOINT ["entrypoint.sh"]
CMD ["reflex", "-c", "/etc/reflex/reflex.conf"]
