FROM golang:1.15-buster

ENV GO111MODULE on
ARG MODULE_PATH=load-generator
ARG MODULE_NAME=load-generator
ENV MODULE_NAME=$MODULE_NAME
ENV MODULE_PATH=$MODULE_PATH

WORKDIR /$GOPATH/src/$MODULE_PATH
RUN apt-get update && apt-get install -y netcat less libnspr4 libnss3 \
    libexpat1 libfontconfig1 libuuid1 wget

RUN wget -q -O - https://dl-ssl.google.com/linux/linux_signing_key.pub | apt-key add -
RUN echo "deb [arch=amd64] http://dl.google.com/linux/chrome/deb/ stable main" >> /etc/apt/sources.list.d/google.list
RUN apt-get update && apt-get install -y google-chrome-stable

  #     && apt-get clean && rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*
# Live-Reload Go project
RUN go get github.com/cespare/reflex

# Add utils scripts for the container
COPY docker/root/ /

COPY go.* ./
COPY . .
RUN go mod download

# ./ffmpeg -rtbufsize 100M -report -re -i "http://www.bok.net/dash/tears_of_steel/cleartext/stream.mpd" o.mp4
EXPOSE 9222-10222
VOLUME /go/src/$MODULE_PATH

ENTRYPOINT ["entrypoint.sh"]
CMD ["reflex", "-c", "/etc/reflex/reflex.conf"]
