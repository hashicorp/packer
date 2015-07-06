all: build

init:
	bundle

dev: init
	PACKER_DISABLE_DOWNLOAD_FETCH=true PACKER_VERSION=1.0 bundle exec middleman server

build: init
	PACKER_DISABLE_DOWNLOAD_FETCH=true PACKER_VERSION=1.0 bundle exec middleman build
