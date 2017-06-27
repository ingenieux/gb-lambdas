clean:
	-rm -rf bin/ pkg/ vendor/pkg/ src/shim/assets/*

sources:
	gb vendor restore

shim:
	-mkdir -p src/shim/assets
	docker run --rm eawsy/aws-lambda-go-shim /bin/bash -c 'cd /shim ; tar czvf - *.pyc *.so' | tar xzvf - -C src/shim/assets
	chmod 755 src/shim/assets/*.so
	upx src/shim/assets/*.so
	docker run --rm eawsy/aws-lambda-go-shim /bin/bash -c '/usr/local/go/bin/go version' > src/shim/assets/goversion
	GOPATH=${PWD}:${PWD}/vendor rice embed-go -i shim

all: shim sources
	gb build
