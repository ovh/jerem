

BUILDPATH=$(CURDIR)/bin
BINPATH=$(BUILDPATH)/jerem

# go parameters
GOCMD=$(shell which go)
GOBUILD=$(GOCMD) build -mod vendor
GOBUILDPATH=src

export GO111MODULE=on

dev: dep fmt lint vet test compile

dep:
	@echo "Ensuring dependencies"
	@cd $(GOBUILDPATH); $(GOCMD) mod vendor
	@echo "Done."

test: dep
	@echo "Testing binary"
	@cd $(GOBUILDPATH); go test  ./...
	@echo "Done."

lint: dep
	@echo "Linting binary"
	@cd $(GOBUILDPATH); golangci-lint run
	@echo "Done."

vet: dep
	@echo "Veting binary"
	@cd $(GOBUILDPATH); go vet $$(go list ./... | grep -v /vendor/)
	@echo "Done."

fmt: dep
	@echo "Fmting binary"
	@cd $(GOBUILDPATH); go fmt $$(go list ./... | grep -v /vendor/)
	@echo "Done."

compile: dep
	@echo "Compiling binary"
	@cd $(GOBUILDPATH); $(GOBUILD) -o $(BINPATH)
	@echo "Done."

clean:
	@echo "Cleaning binary..."
	@if [ -d $(BUILDPATH) ] ; then rm -r $(BUILDPATH) ; fi
	@echo "Done."
