FROM golang:1.7.1
ARG GLIDE_VERSION=0.12.3

RUN export DEBIAN_FRONTEND=noninteractive && \
apt-get -qq update && \
apt-get -y install zip

RUN wget -nv https://github.com/Masterminds/glide/releases/download/v$GLIDE_VERSION/glide-v$GLIDE_VERSION-linux-amd64.tar.gz -O- | \
tar xzf - -C /usr/local/bin/ --strip-components=1 linux-amd64/glide
