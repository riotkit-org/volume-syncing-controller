include examples.mk

.PHONY: gen-api
gen-api:
	./hack/update-codegen.sh
	git add pkg/apis pkg/client

CONTROLLER_GEN := $(GOPATH)/bin/controller-gen
$(CONTROLLER_GEN):
	pushd /tmp; $(GO) get -u sigs.k8s.io/controller-tools/cmd/controller-gen@v0.4.0; popd

crd-manifests: $(CONTROLLER_GEN)
	$(CONTROLLER_GEN) crd:maxDescLen=0 paths="./pkg/apis/riotkit.org/v1alpha1/..." output:crd:artifacts:config=crds
	cp crds/* helm/volume-syncing-operator/templates/
	git add crds helm/volume-syncing-operator/templates/

.PHONY: build
build: gen-api build-binary crd-manifests

.PHONY: build-binary
build-binary:
	@echo "\n🔧  Building Go binaries..."
	mkdir -p .build
	CGO_ENABLED=0 GO111MODULE=on GOOS=linux GOARCH=amd64 go build -o .build/volume-syncing-operator .

.PHONY: build-docker
build-docker:
	docker build . -t volume-syncer

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
