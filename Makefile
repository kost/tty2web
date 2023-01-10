OUTPUT_DIR = ./builds
GIT_COMMIT = `git rev-parse HEAD | cut -c1-7`
VERSION = 2.6.5
BUILD_OPTIONS = -ldflags "-X main.Version=$(VERSION) -X main.CommitID=$(GIT_COMMIT)"
STATIC_OPTIONS = -ldflags "-extldflags='-static' -X main.Version=$(VERSION) -X main.CommitID=$(GIT_COMMIT)"


tty2web: main.go server/*.go webtty/*.go backend/*.go Makefile
	go get -u ./...
	go build ${BUILD_OPTIONS}

tty2web-static: main.go server/*.go webtty/*.go backend/*.go Makefile
	go get -u ./...
	CGO_ENABLED=0 go build ${STATIC_OPTIONS}

.PHONY: asset
asset: bindata/static/js/tty2web-bundle.js bindata/static/index.html bindata/static/favicon.png bindata/static/css/index.css bindata/static/css/xterm.css bindata/static/css/xterm_customize.css bindata/static/js/sidenav.js
	go-bindata -prefix bindata -pkg server -ignore=\\.gitkeep -o server/asset.go bindata/...
	gofmt -w server/asset.go

.PHONY: all
all: asset tty2web

bindata:
	mkdir -p bindata

bindata/static: bindata
	mkdir -p bindata/static

bindata/static/index.html: bindata/static resources/index.html
	cp resources/index.html bindata/static/index.html

bindata/static/favicon.png: bindata/static resources/favicon.png
	cp resources/favicon.png bindata/static/favicon.png

bindata/static/js: bindata/static
	mkdir -p bindata/static/js

bindata/static/js/tty2web-bundle.js: bindata/static/js js/dist/tty2web-bundle.js
	cp js/dist/tty2web-bundle.js bindata/static/js/tty2web-bundle.js

bindata/static/js/sidenav.js: bindata/static/js resources/js/sidenav.js
	cp resources/js/sidenav.js bindata/static/js/sidenav.js

bindata/static/css: bindata/static
	mkdir -p bindata/static/css

bindata/static/css/index.css: bindata/static/css resources/index.css
	cp resources/index.css bindata/static/css/index.css

bindata/static/css/xterm_customize.css: bindata/static/css resources/xterm_customize.css
	cp resources/xterm_customize.css bindata/static/css/xterm_customize.css

bindata/static/css/xterm.css: bindata/static/css js/node_modules/xterm/css/xterm.css
	cp js/node_modules/xterm/css/xterm.css bindata/static/css/xterm.css

js/node_modules/xterm/dist/xterm.css:
	cd js && \
	npm install

js/dist/tty2web-bundle.js: js/src/* js/node_modules/webpack
	cd js && \
	`npm bin`/webpack

js/node_modules/webpack:
	cd js && \
	npm install

tools:
	go get github.com/mitchellh/gox
	go get github.com/tcnksm/ghr
	go get github.com/jteeuwen/go-bindata/...

test:
	if [ `go fmt $(go list ./... | grep -v /vendor/) | wc -l` -gt 0 ]; then echo "go fmt error"; exit 1; fi

cross_compile:
	GOARM=5 gox -os="darwin linux freebsd netbsd openbsd" -arch="386 amd64 arm" -osarch="!darwin/arm" -output "${OUTPUT_DIR}/pkg/{{.OS}}_{{.Arch}}/{{.Dir}}"

targz:
	mkdir -p ${OUTPUT_DIR}/dist
	cd ${OUTPUT_DIR}/pkg/; for osarch in *; do (cd $$osarch; tar zcvf ../../dist/tty2web_${VERSION}_$$osarch.tar.gz ./*); done;

shasums:
	cd ${OUTPUT_DIR}/dist; sha256sum * > ./SHA256SUMS

rel:
	mkdir -p release
	CGO_ENABLED=0 gox -osarch="!darwin/386" -ldflags="-s -w -X main.Version=$(VERSION) -X main.CommitID=$(GIT_COMMIT)" -output="release/{{.Dir}}_{{.OS}}_{{.Arch}}"

draft:
	ghr -draft v$(VERSION) release/

release:
	ghr -c ${GIT_COMMIT} --delete --prerelease -u kost -r tty2web pre-release ${OUTPUT_DIR}/dist
