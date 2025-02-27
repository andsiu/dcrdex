#!/bin/sh

set -e

VER="0.5.0"

rm -rf bin
mkdir -p bin/dexc-windows-amd64-v${VER}
mkdir -p bin/dexc-linux-amd64-v${VER}
mkdir -p bin/dexc-linux-arm64-v${VER}
mkdir -p bin/dexc-darwin-amd64-v${VER}
mkdir -p bin/dexc-darwin-arm64-v${VER}

export CGO_ENABLED=0

LDFLAGS="-s -w -X main.Version=${VER}+release"

pushd client/cmd/dexc
GOOS=linux GOARCH=amd64 go build -trimpath -o ../../../bin/dexc-linux-amd64-v${VER} -ldflags "$LDFLAGS"
GOOS=linux GOARCH=arm64 go build -trimpath -o ../../../bin/dexc-linux-arm64-v${VER} -ldflags "$LDFLAGS"
GOOS=windows GOARCH=amd64 go build -trimpath -o ../../../bin/dexc-windows-amd64-v${VER} -ldflags "$LDFLAGS"
GOOS=darwin GOARCH=amd64 go build -trimpath -o ../../../bin/dexc-darwin-amd64-v${VER} -ldflags "$LDFLAGS"
GOOS=darwin GOARCH=arm64 go build -trimpath -o ../../../bin/dexc-darwin-arm64-v${VER} -ldflags "$LDFLAGS"
popd

LDFLAGS="-s -w -X main.Version=${VER}+release"

pushd client/cmd/dexcctl
GOOS=linux GOARCH=amd64 go build -trimpath -o ../../../bin/dexc-linux-amd64-v${VER} -ldflags "$LDFLAGS"
GOOS=linux GOARCH=arm64 go build -trimpath -o ../../../bin/dexc-linux-arm64-v${VER} -ldflags "$LDFLAGS"
GOOS=windows GOARCH=amd64 go build -trimpath -o ../../../bin/dexc-windows-amd64-v${VER} -ldflags "$LDFLAGS"
GOOS=darwin GOARCH=amd64 go build -trimpath -o ../../../bin/dexc-darwin-amd64-v${VER} -ldflags "$LDFLAGS"
GOOS=darwin GOARCH=arm64 go build -trimpath -o ../../../bin/dexc-darwin-arm64-v${VER} -ldflags "$LDFLAGS"
popd

pushd client/webserver/site
go generate # should be a no-op
npm ci
npm run build
popd

rm -rf bin/site
mkdir -p bin/site/src
pushd client/webserver/site
cp -R dist ../../../bin/site
cp -R src/font src/html src/localized_html src/img ../../../bin/site/src
popd

pushd bin
cp -R site dexc-windows-amd64-v${VER}
cp -R site dexc-darwin-amd64-v${VER}
cp -R site dexc-darwin-arm64-v${VER}
cp -R site dexc-linux-amd64-v${VER}
cp -R site dexc-linux-arm64-v${VER}
zip -9 -r -q dexc-windows-amd64-v${VER}.zip dexc-windows-amd64-v${VER}
tar -I 'gzip -9' --owner=0 --group=0 -cf dexc-linux-amd64-v${VER}.tar.gz dexc-linux-amd64-v${VER}
tar -I 'gzip -9' --owner=0 --group=0 -cf dexc-linux-arm64-v${VER}.tar.gz dexc-linux-arm64-v${VER}
tar -I 'gzip -9' --owner=0 --group=0 -cf dexc-darwin-amd64-v${VER}.tar.gz dexc-darwin-amd64-v${VER}
tar -I 'gzip -9' --owner=0 --group=0 -cf dexc-darwin-arm64-v${VER}.tar.gz dexc-darwin-arm64-v${VER}
sha256sum *.gz *.zip > dexc-v${VER}-manifest.txt
popd
