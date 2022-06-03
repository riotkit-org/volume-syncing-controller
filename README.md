volume-syncing-operator
=======================

Docker container and Kubernetes operator for periodically synchronizing volumes to cloud-native storage, and restoring their state from cloud-native storage.

**Features:**
- [x] Supports all storage kinds that are supported by Rclone
- [x] Rclone configuration using environment variables
- [x] End-To-End encryption support
- [x] Periodical synchronization with built-in cron-like scheduler
- [x] `volume-syncing-operator sync-to-remote` command to synchronize local files to remote
- [x] `volume-syncing-operator remote-to-local-sync` command to sync files back from remote to local
- [ ] Support for Kubernetes: **initContainer** to `restore` files, and **side-car** to back up files to remote
- [x] Extra security layer preventing from accidental file deletion in comparison to plain `rclone` or `rsync` usage :100:
- [x] Non-root container

Runtime compatibility
---------------------

| :penguin: Platform | Usage type                                                                                                                    | 
|--------------------|-------------------------------------------------------------------------------------------------------------------------------|
| Bare metal         | :heavy_check_mark: Environment variables                                                                                      |
| Docker             | :heavy_check_mark: Environment variables                                                                                      |
| Kubernetes         | :heavy_check_mark: Environment variables, Helm + operator (mutating `kind: Pod` by adding initContainer + side-car container) |
 

How it works?
-------------

Container and binary are configured with environment variables that are translated into rclone configuration file.

**Example usage:**

```bash
export REMOTE_TYPE=s3
export REMOTE_PROVIDER=Minio
export REMOTE_ACCESS_KEY_ID=AKIAIOSFODNN7EXAMPLE
export REMOTE_SECRET_ACCESS_KEY=wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY
export REMOTE_ENDPOINT=http://localhost:9000
export REMOTE_ACL=private

# synchronize to remote storage
volume-syncing-operator sync-to-remote -s ./ -d testbucket/some-directory

# synchronize back from remote storage to local directory
volume-syncing-operator remote-to-local-sync -v -s testbucket -d ./.build/testing-restore
```

**Will translate into configuration:**

```conf
[remote]
type = s3
provider = Minio
access_key_id = AKIAIOSFODNN7EXAMPLE
secret_access_key = wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY
endpoint = http://localhost:9000
acl = private
```

**And will run:**

```bash
rclone sync ./ remote:/testbucket/some-directory
```

### For more examples see our [Makefile](./examples.mk) full of examples that could be run in context of this repository

:watch: Scheduling periodically
-------------------------------

`volume-syncing-operator` has built-in crontab-like scheduler to optionally enable periodical syncing.

```
volume-syncing-operator --schedule '@every 1m'    # ...
volume-syncing-operator --schedule '0 30 * * * *' # ...
```

:black_circle: Safety valves
----------------------------

There are various "safety valves" that in default configuration would try to prevent misconfigured run from deleting your data.

### sync-to-remote

| :black_circle: Safety rule                                                               |
|------------------------------------------------------------------------------------------|
| Local directory cannot be empty (it would mean deleting all files from remote directory) |


### remote-to-local-sync

Validation of those rules can be intentionally skipped using commandline switch `--force-delete-local-dir`

| :black_circle: Safety rule                                                                                                         |
|------------------------------------------------------------------------------------------------------------------------------------|
| Remote directory cannot be empty (it would mean deleting all local files)                                                          |
| `/usr/bin`, `/bin`, `/`, `/home`, `/usr/lib` cannot be picked as synchronization root due to risk of erasing your operating system |
| Local target directory cannot be owned by root                                                                                     |
| Local target directory must be owned by same user id as the current process runs on                                                |


Kubernetes example
------------------

Operator have to be installed using Helm first. It will react for every Pod labelled with `riotkit.org/volume-syncing-operator: true` and matching CRD of `PodFilesystemSync` kind.

The `riotkit.org/volume-syncing-operator: true` is there for performance and cluster safety by limiting the scope of Admission Webhook.

```yaml
---
apiVersion: riotkit.org/v1alpha1
kind: PodFilesystemSync
metadata:
    name: cloud-press
spec:
    podSelector:
        my-pod-label: test

    localPath: /var/www/riotkit/wp-content
    remotePath: /example-org-bucket
    schedule: "@every 5m"
    env:
        REMOTE_TYPE: s3
        REMOTE_PROVIDER: Minio
        REMOTE_ENDPOINT: http://localhost:9000
        REMOTE_ACL: private

        # best practice is to move sensitive information into `kind: Secret`
        # and reference that secret in `envFromSecret`
        # to keep your secret in GIT you can try e.g. SealedSecrets or ExternalSecrets
        #REMOTE_ACCESS_KEY_ID: ...
        #REMOTE_SECRET_ACCESS_KEY: ...
    envFromSecret:
        - ref: cloud-press-secret-envs
    # will generate a key, store it in `kind: Secret` and setup End-To-End encryption
    automaticEncryption: true
```
