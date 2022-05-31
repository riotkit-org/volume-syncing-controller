.PHONY: test
test: test_sync_without_encryption test_sync_using_envs test_remote_to_local

#
# Uses commandline switches to configure rclone
#
.PHONY: test_sync_without_encryption
test_sync_without_encryption:
	.build/volume-syncing-operator sync-to-remote -d testbucket -p 'type=s3' -p 'provider=Minio' -p 'access_key_id=AKIAIOSFODNN7EXAMPLE' -p 'secret_access_key = wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY' -p 'endpoint = http://localhost:9000' -p 'acl = private'

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
	.build/volume-syncing-operator sync-to-remote -d testbucket

.PHONY: test_sync_every_1_minute
test_sync_every_1_minute:
	export REMOTE_TYPE=s3; \
	export REMOTE_PROVIDER=Minio; \
	export REMOTE_ACCESS_KEY_ID=AKIAIOSFODNN7EXAMPLE; \
	export REMOTE_SECRET_ACCESS_KEY=wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY; \
	export REMOTE_ENDPOINT=http://localhost:9000; \
	export REMOTE_ACL=private; \
	.build/volume-syncing-operator sync-to-remote -d testbucket --schedule "@every 1m"

test_remote_to_local:
	# upload
	.build/volume-syncing-operator sync-to-remote -d testbucket -s ./pkg -p 'type=s3' -p 'provider=Minio' -p 'access_key_id=AKIAIOSFODNN7EXAMPLE' -p 'secret_access_key = wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY' -p 'endpoint = http://localhost:9000' -p 'acl = private'

	# then download into different directory
	.build/volume-syncing-operator remote-to-local-sync -v -s testbucket -d ./.build/testing-restore -p 'type=s3' -p 'provider=Minio' -p 'access_key_id=AKIAIOSFODNN7EXAMPLE' -p 'secret_access_key = wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY' -p 'endpoint = http://localhost:9000' -p 'acl = private'
