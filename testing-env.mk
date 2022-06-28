IMAGE_TAG=snapshot
DEV_LOCAL_IMAGE_REPOSITORY=127.0.0.1:30050/volume-syncing-operator
CHART_NAME=volume-syncing-operator

k3d@cluster: ## Run local empty Kubernetes cluster
	k3d cluster create volume-syncing-operator-sandbox --agents 1 -p "30080:30080@agent:0" -p "30081:30081@agent:0" -p "30050:30050@agent:0"

k3d@registry: k3d@make-sure
	helm repo add twuni https://helm.twun.io
	helm upgrade --install registry twuni/docker-registry -n default --set ingress.enabled=true --set ingress.hosts[0]=registry.ingress.cluster.local
	kubectl apply -f tests/.helpers/local-registry.yaml

k3d@minio: k3d@make-sure
	helm repo add minio https://helm.min.io/
	helm upgrade --install minio minio/minio --values ./tests/.helpers/local-minio.yaml --wait --timeout 2m0s

k3d@publish-image: ## Publish to local Kubernetes registry
	docker build . -t ${DEV_LOCAL_IMAGE_REPOSITORY}:snapshot
	docker push ${DEV_LOCAL_IMAGE_REPOSITORY}:snapshot

k3d@deploy: k3d@make-sure
	cd helm/${CHART_NAME} && helm upgrade --install vso . --values ../../tests/.helpers/local-release.values.yaml --debug

k3d@release: k3d@make-sure
	kubectl delete deployment vso-volume-syncing-operator || true
	make build-binary k3d@publish-image k3d@deploy
	sleep 5; kubectl logs -f deployment/vso-volume-syncing-operator

k3d@template:
	cd helm/${CHART_NAME} && helm template vso . --values ../../tests/.helpers/local-release.values.yaml --debug

k3d@make-sure:
	@kubectl cluster-info | grep https://0.0.0.0 > /dev/null || (echo "KUBECONFIG does not point to test k3d cluster" && exit 1)
