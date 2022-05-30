volume-syncing-operator
=======================

Docker container and Kubernetes operator for periodically synchronizing volumes to cloud-native storage, and restoring their state from cloud-native storage.

**Features:**
- [x] Supports all storage kinds that are supported by Rclone
- [x] Rclone configuration using environment variables
- [x] End-To-End encryption support
- [ ] Periodical synchronization with built-in cron-like scheduler
- [x] `volume-syncing-operator sync-to-remote` command to synchronize local files to remote
- [ ] `volume-syncing-operator restore` command to sync files back from remote to local
- [ ] Support for Kubernetes: **initContainer** to `restore` files, and **side-car** to back up files to remote

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

volume-syncing-operator sync-to-remote -s ./ -d testbucket/some-directory
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

Scheduling periodically
-----------------------

`volume-syncing-operator` has built-in crontab-like scheduler to optionally enable periodical syncing.

```
volume-syncing-operator --schedule '@every 1m'    # ...
volume-syncing-operator --schedule '0 30 * * * *' # ...
```
