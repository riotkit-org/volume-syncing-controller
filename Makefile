include examples.mk

.PHONY: build
build:
	@echo "\n🔧  Building Go binaries..."
	mkdir -p .build
	CGO_ENABLED=0 GO111MODULE=on GOOS=linux GOARCH=amd64 go build -o .build/volume-syncer .

.PHONY: build_docker
build_docker:
	docker build . -t volume-syncer

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
