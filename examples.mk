.PHONY: test

# <test pipelines>
test: test_sync_without_encryption test_sync_using_envs test_remote_to_local
test_k8s: test_k8s_without_encryption_scheduler_permissions test_k8s_dynamic_directory_name test_k8s_with_encryption
# <end of test pipelines>

#
# Uses commandline switches to configure rclone
#
.PHONY: test_sync_without_encryption
test_sync_without_encryption:
	.build/volume-syncing-controller sync-to-remote -d testbucket -p 'type=s3' -p 'provider=Minio' -p 'access_key_id=AKIAIOSFODNN7EXAMPLE' -p 'secret_access_key = wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY' -p 'endpoint = http://localhost:9000' -p 'acl = private'

#
# Environment variables are more handy when using Docker, docker-compose and Kubernetes
#
.PHONY: test_sync_using_envs
test_sync_using_envs:
	export REMOTE_TYPE=s3; \
	export REMOTE_PROVIDER=Minio; \
	export REMOTE_ACCESS_KEY_ID=AKIAIOSFODNN7EXAMPLE; \
	export REMOTE_SECRET_ACCESS_KEY=wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY; \
	export REMOTE_ENDPOINT=http://localhost:9000; \
	export REMOTE_ACL=private; \
	.build/volume-syncing-controller sync-to-remote -d testbucket

.PHONY: test_sync_every_1_minute
test_sync_every_1_minute:
	export REMOTE_TYPE=s3; \
	export REMOTE_PROVIDER=Minio; \
	export REMOTE_ACCESS_KEY_ID=AKIAIOSFODNN7EXAMPLE; \
	export REMOTE_SECRET_ACCESS_KEY=wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY; \
	export REMOTE_ENDPOINT=http://localhost:9000; \
	export REMOTE_ACL=private; \
	.build/volume-syncing-controller sync-to-remote -d testbucket --schedule "@every 1m"

test_remote_to_local:
	# upload
	.build/volume-syncing-controller sync-to-remote -d testbucket -s ./pkg -p 'type=s3' -p 'provider=Minio' -p 'access_key_id=AKIAIOSFODNN7EXAMPLE' -p 'secret_access_key = wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY' -p 'endpoint = http://localhost:9000' -p 'acl = private'

	# then download into different directory
	.build/volume-syncing-controller remote-to-local-sync -v -s testbucket -d ./.build/testing-restore -p 'type=s3' -p 'provider=Minio' -p 'access_key_id=AKIAIOSFODNN7EXAMPLE' -p 'secret_access_key = wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY' -p 'endpoint = http://localhost:9000' -p 'acl = private'


_test_k8s_variant:
	kubectl delete -f tests/.examples/${VARIANT}/sync.yaml || true
	kubectl delete -f tests/.examples/${VARIANT}/pod.yaml || true
	sleep 1
	kubectl apply -f tests/.examples/${VARIANT}/sync.yaml
	kubectl apply -f tests/.examples/${VARIANT}/pod.yaml

test_k8s_without_encryption_scheduler_permissions:
	make _test_k8s_variant VARIANT=minio-without-encryption-scheduler-permissions

test_k8s_dynamic_directory_name:
	make _test_k8s_variant VARIANT=with-dynamic-directory-name

test_k8s_with_encryption:
	make _test_k8s_variant VARIANT=with-encryption
