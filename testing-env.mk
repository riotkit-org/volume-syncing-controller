IMAGE_TAG=snapshot
DEV_LOCAL_IMAGE_REPOSITORY=127.0.0.1:30050/volume-syncing-operator

k3d@cluster: ## Run local empty Kubernetes cluster
	k3d cluster create volume-syncing-operator-sandbox --agents 1 -p "30080:30080@agent:0" -p "30081:30081@agent:0" -p "30050:30050@agent:0"

k3d@registry: k3d@make-sure
	helm repo add twuni https://helm.twun.io
	helm upgrade --install registry twuni/docker-registry -n default --set ingress.enabled=true --set ingress.hosts[0]=registry.ingress.cluster.local
	kubectl apply -f tests/.helpers/local-registry.yaml

k3d@publish-image: ## Publish to local Kubernetes registry
	docker build . -t ${DEV_LOCAL_IMAGE_REPOSITORY}:snapshot
	docker push ${DEV_LOCAL_IMAGE_REPOSITORY}:snapshot

k3d@make-sure:
	@kubectl cluster-info | grep https://0.0.0.0 > /dev/null || (echo "KUBECONFIG does not point to test k3d cluster" && exit 1)
