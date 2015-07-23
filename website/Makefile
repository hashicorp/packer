all: build

init:
	bundle

dev: init
	PACKER_DISABLE_DOWNLOAD_FETCH=true PACKER_VERSION=1.0 bundle exec middleman server

build: init
	PACKER_DISABLE_DOWNLOAD_FETCH=true PACKER_VERSION=1.0 bundle exec middleman build

format:
	bundle exec htmlbeautifier -t 2 source/*.erb
	bundle exec htmlbeautifier -t 2 source/layouts/*.erb
	pandoc -v > /dev/null && find . -iname "*.html.markdown" | xargs -I{} bash -c "pandoc -r markdown -w markdown --tab-stop=2 --atx-headers -s --columns=80 {} > {}.new"\; || true
	pandoc -v > /dev/null && find . -iname "*.html.markdown" | xargs -I{} bash -c "mv {}.new {}"\; || true

