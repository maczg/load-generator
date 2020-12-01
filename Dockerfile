#----
# Build stage
#----

FROM golang:1.15-buster as buildstage
WORKDIR /$GOPATH/src/mailer

# Install git
RUN apt-get update && apt-get install -y git make gcc tar rsync

# Enable go modules
ENV GO111MODULE on
ARG MODULE_PATH
ARG MODULE_NAME
ENV MODULE_NAME=$MODULE_NAME
ENV MODULE_PATH=$MODULE_PATH

# Populate the module cache based on the go.{mod,sum} files for maas
COPY go.mod go.mod
COPY go.sum go.sum
RUN go mod download

COPY . .

# Compile with static linking
RUN make build distMode=dir ignoreMissing=yes
RUN mv ./dist/service/service /service

#----
# Microservice stage
#----
FROM scratch
# COPY ./conf ./conf
# Copy built executable
COPY --from=buildstage /service ./

CMD ["./service"]
