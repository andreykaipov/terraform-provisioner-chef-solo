VERSION=0.0.1
NAME=terraform-provisioner-chef-solo_v${VERSION}
TF_DIR=$$(dirname $$(which terraform))

.PHONY: build install clean test release

build:
	@echo Building ${NAME} plugin...
	@mkdir -p bin
	@go build -o bin/${NAME}

install:
	@echo Installing ${NAME} to ${TF_DIR}
	@install -c bin/${NAME} /usr/local/bin/${NAME}

clean:
	@rm -rf bin/

test:
	@echo Running tests...
	@go test $$(go list ./...)

release:
	@echo Building for each OS with GOARCH=amd64
	@for os in darwin linux windows; do\
		if [ $$os == windows ]; then extension=.exe; fi; \
		GOOS=$$os GOARCH=amd64 go build -o bin/${NAME}$$extension;\
		pushd bin;\
		zip ${NAME}-$$os-amd64.zip ${NAME}$$extension;\
		rm ${NAME}$$extension;\
		popd;\
	done
