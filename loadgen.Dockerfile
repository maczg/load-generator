FROM aleskandro/golang-ffmpeg:1.15-buster
ENV GO111MODULE on
ARG MODULE_PATH=load-generator
ARG MODULE_NAME=load-generator
ENV MODULE_NAME=$MODULE_NAME
ENV MODULE_PATH=$MODULE_PATH
ENV GOPATH=/go

WORKDIR /$GOPATH/src/$MODULE_PATH
RUN apt-get update && apt-get install -y netcat less
# Live-Reload Go project
RUN go get github.com/cespare/reflex

# Add utils scripts for the container
COPY docker/root/ /

COPY go.* ./
RUN go mod download

# ./ffmpeg -rtbufsize 100M -report -re -i "http://www.bok.net/dash/tears_of_steel/cleartext/stream.mpd" o.mp4

ENTRYPOINT ["entrypoint.sh"]
CMD ["reflex", "-c", "/etc/reflex/reflex.conf"]
