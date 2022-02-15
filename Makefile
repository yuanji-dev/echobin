BIN := ./echobin
DOCKER_IMAGE ?= gimo/echobin
DOCKER_TAG ?= latest
DOCKER_REF := $(DOCKER_IMAGE):$(DOCKER_TAG)

.PHONY: test
test:
	go test -v ./...

.PHONY: docs
docs:
	@hash swag > /dev/null 2>&1; if [ $$? -ne 0 ]; then \
		go install github.com/swaggo/swag/cmd/swag@latest; \
	fi
	swag fmt && swag init -o static/swagger-ui -ot json

.PHONY: watch
watch:
	@hash CompileDaemon > /dev/null 2>&1; if [ $$? -ne 0 ]; then \
		go install github.com/githubnemo/CompileDaemon@latest; \
	fi
	CompileDaemon -exclude-dir=.git -command $(BIN)

.PHONY: clean
clean:
	rm -rf $(BIN)

.PHONY: docker-build
docker-build: clean
	docker build --no-cache -t $(DOCKER_REF) .

.PHONY: docker-push
docker-push:
	docker push $(DOCKER_REF)

.PHONY: docker-run
docker-run:
	docker run --rm -p 8080:8080 $(DOCKER_REF)
