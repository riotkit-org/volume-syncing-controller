include examples.mk
include testing-env.mk

HELM_DOCS_IMAGE = jnorwood/helm-docs:v1.8.1

.PHONY: gen-api
gen-api:
	./hack/update-codegen.sh
	git add pkg/apis pkg/client

CONTROLLER_GEN := $(GOPATH)/bin/controller-gen
$(CONTROLLER_GEN):
	pushd /tmp; $(GO) get -u sigs.k8s.io/controller-tools/cmd/controller-gen@v0.4.0; popd

crd-manifests: $(CONTROLLER_GEN)
	$(CONTROLLER_GEN) crd:maxDescLen=0 paths="./pkg/apis/riotkit.org/v1alpha1/..." output:crd:artifacts:config=crds
	cp crds/* helm/volume-syncing-controller/templates/
	git add crds helm/volume-syncing-controller/templates/

.PHONY: build-all
build-all: gen-api build-binary crd-manifests

.PHONY: build
build: build-binary

.PHONY: build-binary
build-binary:
	@echo "\n🔧  Building Go binaries..."
	mkdir -p .build
	CGO_ENABLED=0 GO111MODULE=on GOOS=linux GOARCH=amd64 go build -o .build/volume-syncing-controller .

.PHONY: build-docker
build-docker:
	docker build . -t volume-syncer

.PHONY: helm
helm:
	cp README.md helm/volume-syncing-controller/
	cd helm/volume-syncing-controller/ && helm lint ./

.PHONY: helm-docs
helm-docs:
	docker run --rm --name helm-docs  \
		--user $(shell id -u):$(shell id -g) \
		--mount type=bind,src="$(shell pwd)/helm/volume-syncing-controller",dst=/helm-charts \
		-w /helm-charts \
		$(HELM_DOCS_IMAGE) \
		helm-docs

.PHONY: coverage
coverage:
	go test -v ./... -covermode=count -coverprofile=coverage.out


.PHONY: minio
minio:
	docker rm -f minio1 || true
	docker run \
      -d \
      -p 9000:9000 \
      -p 9001:9001 \
      --name minio1 \
      -v /tmp/minio/data:/data \
      -e "MINIO_ROOT_USER=AKIAIOSFODNN7EXAMPLE" \
      -e "MINIO_ROOT_PASSWORD=wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY" \
      quay.io/minio/minio server /data --console-address ":9001"
