TEST?=$$(go list ./... |grep -v 'vendor')
TEST_JDCLOUD?=$$(go list ./... |grep 'jdcloud/jdcloud')
GO_FMT_FILES?=$$(find . -name '*.go' |grep -v vendor)
WEBSITE_REPO=github.com/hashicorp/terraform-website


default: build

build: fmtcheck
	go install

test: fmtcheck
	go test $(TEST) -timeout=30s -parallel=4

testacc: fmtcheck
	TF_ACC=1 go test ./... -v -timeout 120m

vet:
	@echo "go vet ."
	@go vet $$(go list ./... | grep -v vendor/);

fmt:
	gofmt -w $(GOFMT_FILES)

fmtcheck:
	@sh -c "'$(CURDIR)/scripts/gofmtcheck.sh'"

errcheck:
	@sh -c "'$(CURDIR)/scripts/errcheck.sh'"

vendor-status:
	@govendor status

tools:
	go get -u github.com/kardianos/govendor
	go get -u github.com/alecthomas/gometalinter
	gometalinter --install

test-compile:
	@if ["$(TEST2)" = "./..."];then \
		echo "[ERROR]:Set TEST to a specific package. For example you can command like this: "; \
		echo "	make test-compile TEST=./$(PKG_NAME)"; \
		exit 1; \
	fi
	go test -c $(TEST_JDCLOUD) $(TESTARGS)


#
# Exit 1 is added in order to avoid unncessary cloning 
#
#website:
#ifeq(,$(wildcard $(GOPATH)/src/$(WEBSITE_REPO)))
#	echo "$(WEBSITE_REPO) not found in your GOPATH which is necessary for layouts and assets, cloning a raw one into your GOPATH"
#	exit 1
#	git clone https://$(WEBSITE_REPO) $(GOPATH)
#endif
#	@$(MAKE) -c $(GOPATH)/src/$(WEBSITE_REPO) website-provider PROVIDER_PATH=$(shell pwd) PROVIDER_NAME=$(PKG_NAME)
#
#website-test:
#ifeq(,$(wildcard $(GOPATH)/src/$(WEBSITE_REPO)))
#	echo "$(WEBSITE_REPO) not found in your GOPATH which is necessary for layouts and assets, cloning a raw one into your GOPATH"
#	git clone https://$(WEBSITE_REPO) $(GOPATH)
#endif
#	@$(MAKE) -c $(GOPATH)/src/$(WEBSITE_REPO) website-provider-test PROVIDER_PATH=$(shell pwd) PROVIDER_NAME=$(PKG_NAME)
#
