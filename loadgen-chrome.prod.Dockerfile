FROM golang:1.15-buster as buildstage
ENV GO111MODULE on
ARG MODULE_PATH=load-generator
ARG MODULE_NAME=load-generator
ENV MODULE_NAME=$MODULE_NAME
ENV MODULE_PATH=$MODULE_PATH

WORKDIR /$GOPATH/src/$MODULE_PATH

COPY go.* ./
RUN go mod download
COPY . .

# Compile with static linking

RUN make build distMode=dir ignoreMissing=yes
RUN mv ./dist/service/service /service


# ################## #
# Microservice stage #
# ################## #
FROM debian:buster-slim

RUN apt-get update && apt-get install -y wget gnupg \
    && apt-get clean && rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

RUN wget -q -O - https://dl-ssl.google.com/linux/linux_signing_key.pub | apt-key add -
RUN echo "deb [arch=amd64] http://dl.google.com/linux/chrome/deb/ stable main" >> /etc/apt/sources.list.d/google.list
RUN apt-get update && apt-get install -y google-chrome-stable \
    && apt-get clean && rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

COPY --from=buildstage /service /usr/bin/load-generator
EXPOSE 9222-10222

CMD ["/usr/bin/load-generator"]
