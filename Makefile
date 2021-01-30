SHELL = /bin/bash -o pipefail
PKGS := github.com/tullo/password/password
VETTERS := "asmdecl,assign,atomic,bools,buildtag,cgocall,composites,copylocks,errorsas,httpresponse,loopclosure,lostcancel,nilfunc,printf,shift,stdmethods,structtag,tests,unmarshal,unreachable,unsafeptr,unusedresult"
SRCDIRS := $(shell go list -f '{{.Dir}}' ./...)

check: test vet gofmt ineffassign misspell staticcheck unconvert unparam

pedantic: check errcheck

gofmt:  
	@echo Checking code is gofmted
	@test -z "$(shell gofmt -s -l -d -e $(SRCDIRS) | tee /dev/stderr)"

test: 
	@go test -race -timeout=1m -vet="${VETTERS}" $(PKGS)

vet: test
	@go vet $(PKGS)

errcheck:
	@cd && GO111MODULE=on go get github.com/kisielk/errcheck
	@$(shell go env GOPATH)/bin/errcheck $(PKGS)

ineffassign:
	@cd && GO111MODULE=on go get github.com/gordonklaus/ineffassign
	@find $(SRCDIRS) -name '*.go' | xargs $(shell go env GOPATH)/bin/ineffassign

misspell:
	@cd && GO111MODULE=on go get github.com/client9/misspell/cmd/misspell
	@$(shell go env GOPATH)/bin/misspell -locale="US" -error -source="text" **/*

staticcheck:
	@cd && GO111MODULE=on go get honnef.co/go/tools/cmd/staticcheck@2020.2.1
	@$(shell go env GOPATH)/bin/staticcheck -go 1.15 -checks all -tests $(PKGS)

unconvert:
	@cd && GO111MODULE=on go get github.com/mdempsky/unconvert
	@$(shell go env GOPATH)/bin/unconvert -v $(PKGS)

unparam:
	@cd && GO111MODULE=on go get mvdan.cc/unparam
	@$(shell go env GOPATH)/bin/unparam ./...
