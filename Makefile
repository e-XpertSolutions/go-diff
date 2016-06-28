PKG  = e-xpert_solutions/go-diff/diff

all: dev-deps check test cover

dev-deps:
	@go get -u -v golang.org/x/tools/cmd/gotype
	@go get -u -v github.com/golang/lint/golint

test:
	go test ${PKG}

check:
	go vet ${PKG}
	golint ${PKG}
	gotype ${GOPATH}/src/${PKG}

cover:
	go test -cover ${PKG}

.PHONY: check, test, dev-deps, cover
