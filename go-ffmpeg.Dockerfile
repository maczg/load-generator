FROM golang:1.15-buster

#WORKDIR /tmp/
#ADD http://download.tsi.telecom-paristech.fr/gpac/release/1.0.1/gpac_1.0.1-rev0-gd8538e8a-master_amd64.deb ./
#RUN apt-get update && apt-get install -y ./gpac*.deb && apt-get clean && rm -rf ./gpac*.deb
RUN apt-get update && apt-get install -y git make gcc tar rsync curl wget vim \
        build-essential yasm nasm  libxml2-dev libxml2-utils libsdl2-dev \
        autoconf automake cmake git-core libass-dev libfreetype6-dev libgnutls28-dev \
        libtool libva-dev libvdpau-dev libvorbis-dev libxcb1-dev libxcb-shm0-dev \
        libxcb-xfixes0-dev pkg-config texinfo zlib1g-dev libx264-dev libx265-dev libnuma-dev \
        libvpx-dev libmp3lame-dev libopus-dev libaom-dev \
        ninja-build meson doxygen \
       && apt-get clean


RUN wget "https://github.com/Netflix/vmaf/archive/v1.5.3.tar.gz" -O vmaf.tar.gz && tar -xvf vmaf.tar.gz && \
    rm -rf vmaf.tar.gz && cd vmaf-1.5.3/libvmaf &&  meson build --buildtype release && \
    ninja -vC build && ninja -vC build install && cd ../../ && rm -rf ./vmaf*
# ENV LD_LIBRARY_PATH=$LD_LIBRARY_PATH:/usr/local/lib

WORKDIR /root
RUN wget "https://ffmpeg.org/releases/ffmpeg-snapshot.tar.bz2" &&\
    tar -xvf ffmpeg-snapshot.tar.bz2 && rm -rf ffmpeg-snapshot.tar.bz2 && \
    cd ./ffmpeg/ && ./configure --enable-demuxer=dash --enable-libxml2 --enable-libmp3lame \
    --enable-libopus --enable-libx265 --enable-gpl --enable-libx264 --enable-nonfree \
    --enable-libaom --enable-gnutls --enable-libvmaf --enable-version3 && make -j$(nproc) && make install && \
    cd ../ && rm -rf ffmpeg && cp /usr/local/lib/x86_64-linux-gnu/libvmaf.* /usr/local/lib/


RUN cp /usr/local/lib/x86_64-linux-gnu/libvmaf.* /usr/local/lib/
ENV LD_LIBRARY_PATH=$LD_LIBRARY_PATH:/usr/local/lib
#RUN mkdir -p /usr/local/share/model
#RUN curl https://raw.githubusercontent.com/Netflix/vmaf/master/model/vmaf_v0.6.1.pkl.model -o /usr/local/share/model/vmaf_v0.6.1.pkl.model
#RUN curl https://raw.githubusercontent.com/Netflix/vmaf/master/model/vmaf_v0.6.1.pkl -o /usr/local/share/model/vmaf_v0.6.1.pkl

# ./ffmpeg -rtbufsize 100M -report -re -i "http://www.bok.net/dash/tears_of_steel/cleartext/stream.mpd" o.mp4


#ENV GO111MODULE on
#ARG MODULE_PATH=load-generator
#ARG MODULE_NAME=load-generator
#ENV MODULE_NAME=$MODULE_NAME
#ENV MODULE_PATH=$MODULE_PATH
#ENV GOPATH=/go
#WORKDIR /$GOPATH/src/$MODULE_PATH
#
## Live-Reload Go project
#RUN go get github.com/cespare/reflex
#
## Add utils scripts for the container
#COPY docker/root/ /
#
#COPY go.* ./
#RUN go mod download
#
#
#
#
## RUN git clone https://github.com/FFmpeg/FFmpeg.git
#ENTRYPOINT ["entrypoint.sh"]
#CMD ["reflex", "-c", "/etc/reflex/reflex.conf"]
#
