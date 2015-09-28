NAME            = discover
DOCKER_IP      ?= docker.local
SRCS            = $(shell find . -name '*.go')
RELEASE_I       = tutum.co/agrarianlabs/$(NAME)

all             : build

.build          : $(SRCS)
		@docker build -t $(RELEASE_I) .
		@touch $@
build           : .build

push            : test build
		@docker push $(RELEASE_I)

test            : .build
		@echo "checking go vet..."
		@docker run --rm $(RELEASE_I) bash -c '[ -z "$$(go vet ./... |& \grep -v old/ | \grep -v Godeps/ | \grep -v "exit status" | tee /dev/stderr || true)" ]' || (echo "go vet issue"; exit 1)
		@echo "checking golint..."
		@docker run --rm $(RELEASE_I) bash -c '[ -z "$$(golint ./... |& \grep -v old/ | \grep -v Godeps/ | tee /dev/stderr || true)" ]' || (echo "golint issue"; exit 1)
		@echo "checking gofmt -s..."
		@docker run --rm $(RELEASE_I) bash -c '[ -z "$$(gofmt -s -l . |& \grep -v old/ | \grep -v Godeps/ | tee /dev/stderr || true)" ]' || (echo "gofmt -s issue"; exit 1)
		@echo "checking goimports..."
		@docker run --rm $(RELEASE_I) bash -c '[ -z "$$(goimports -l . |& \grep -v old/ | \grep -v Godeps/ | tee /dev/stderr || true)" ]' || (echo "goimports issue"; exit 1)
		@echo "running the tests..."
		@docker run --rm $(RELEASE_I) godep go test -v -cover -coverprofile=/tmp/c .

clean           :
		@rm -f .build

.PHONY          : build clean all push test
