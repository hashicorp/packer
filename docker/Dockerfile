FROM golang:1.3.3

# build gox toolchhain for crosscompile
RUN go get github.com/mitchellh/gox \
	&& gox -build-toolchain

# get golang dependencies
RUN go get -d github.com/mitchellh/packer \
	&& cd /go/src/github.com/mitchellh/packer \
	&& make updatedeps

# get the patch from ekle/goamz (see: https://github.com/mitchellh/goamz/pull/154)
RUN cd /go/src/github.com/mitchellh/goamz \
	&& git remote add hack  https://github.com/ekle/goamz.git \
	&& git fetch hack \
	&& git checkout -b fix \
	&& git reset --hard hack/master \
	&& git config --global user.email "no@name.com" \
	&& git rebase origin/master \
	&& git checkout master \
	&& git reset --hard fix

ENV XC_ARCH amd64
ENV XC_OS darwin

RUN cd /go/src/github.com/mitchellh/packer/ \
	&& make dev \
	&& tar czvf packer.tgz pkg/darwin_amd64/*
    
