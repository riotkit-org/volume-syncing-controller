.PHONY: test
test: test_sync_without_encryption test_sync_using_envs

.PHONY: test_sync_without_encryption
test_sync_without_encryption:
	.build/volume-syncer sync-to-remote -d testbucket -p 'type=s3' -p 'provider=Minio' -p 'access_key_id=AKIAIOSFODNN7EXAMPLE' -p 'secret_access_key = wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY' -p 'endpoint = http://localhost:9000' -p 'acl = private'

.PHONY: test_sync_using_envs
test_sync_using_envs:
	export REMOTE_TYPE=s3; \
	export REMOTE_PROVIDER=Minio; \
	export REMOTE_ACCESS_KEY_ID=AKIAIOSFODNN7EXAMPLE; \
	export REMOTE_SECRET_ACCESS_KEY=wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY; \
	export REMOTE_ENDPOINT=http://localhost:9000; \
	export REMOTE_ACL=private; \
	.build/volume-syncer sync-to-remote -d testbucket

.PHONY: test_sync_every_1_minute
test_sync_every_1_minute:
	export REMOTE_TYPE=s3; \
	export REMOTE_PROVIDER=Minio; \
	export REMOTE_ACCESS_KEY_ID=AKIAIOSFODNN7EXAMPLE; \
	export REMOTE_SECRET_ACCESS_KEY=wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY; \
	export REMOTE_ENDPOINT=http://localhost:9000; \
	export REMOTE_ACL=private; \
	.build/volume-syncer sync-to-remote -d testbucket --schedule "@every 1m"
