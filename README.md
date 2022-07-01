volume-syncing-controller
=========================

Docker container and Kubernetes controller for periodically synchronizing volumes to cloud-native storage, and restoring their state from cloud-native storage.

**Features:**
- [x] Supports all storage kinds that are supported by Rclone
- [x] Rclone configuration using environment variables
- [x] End-To-End encryption support
- [x] Periodical synchronization with built-in cron-like scheduler
- [x] `volume-syncing-operator sync-to-remote` command to synchronize local files to remote
- [x] `volume-syncing-operator remote-to-local-sync` command to sync files back from remote to local
- [x] Support for Kubernetes: **initContainer** to `restore` files, and **side-car** to back up files to remote
- [x] Extra security layer preventing from accidental file deletion in comparison to plain `rclone` or `rsync` usage :100:
- [x] Non-root container
- [ ] Periodical synchronization using filesystem events instead of cron-like scheduler (both should be available)
- [x] Jinja2 templating support inside `kind: PodFilesystemSync` to allow using single definition for multiple `kind: Pod` objects
- [ ] Termination hook to synchronize Pod before it gets terminated
- [ ] Health check: If N-synchronization fails, then mark Pod as unhealthy
- [ ] Allow to decide about the order of initContainer in CRD + annotation
- [ ] Allow to disable synchronization or restore in CRD + annotation

Kubernetes operator architecture
--------------------------------

The solution architecture is designed to be Pod-centric and live together with the application, not on the underlying infrastructural level.

Above means, that when Pod starts - the volume is **restored from remote**, then all data is **synchronized to remote periodically** during Pod lifetime to keep an external storage up-to-date.


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
    # Follows K8s convention: https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/#resources-that-support-set-based-requirements
    podSelector:
        matchLabels:
            my-pod-label: test

    localPath: /var/www/riotkit/wp-content
    remotePath: /example-org-bucket          # Can be also a JINJA2 template with access to Pod. Example: '/stalin-was-a-dickhead/{{ pod.ObjectMeta.Annotations["subdir"] }}'

    syncOptions:
        # NOTICE: every next synchronization will be cancelled if previous one was not finished
        method: "scheduler"  # or "fsnotify"
        schedule: "@every 5m"
        maxOneSyncPerMinutes: "15"  # When "fsnotify" used, then perform only max one sync per N minutes. Allows to decrease network/cpu/disk usage with a little risk factor

        # Optional "RunAs"
        permissions:
            # Can be overridden by Pod annotation `riotkit.org/volume-user-id`
            uid: 1001
            # Can be overridden by Pod annotation `riotkit.org/volume-group-id`
            gid: 1001
            
        # Optional
        cleanUp:
            # Decides if files are synchronized or copied. Synchronization means that redundant files are deleted. 
            # When in directory A you delete a file, then it gets deleted in directory B
            remote: true
            local: true

            # Disables "security valves"
            forceRemote: false    
            forceLocal: false

        # Optional
        allowedDirectories:
            # Decides if a side-car container should be spawned to synchronize changes periodically
            toRemote: true
            # Decides if an init container should be placed to restore files from remote on startup
            fromRemote: true
            
        # Allows to decide on a case, when the directory is to be synchronized first time. 
        # Should the initContainer be placed and restore from remote (it may be potentially empty if is new?), or should it be skipped first time?
        # [!!!] This counts not for Pods, not for whole PodFilesystemSync but PER .spec.remotePath
        restoreRemoteOnFirstRun: true
    env:
        REMOTE_TYPE: s3
        REMOTE_PROVIDER: Minio
        REMOTE_ENDPOINT: http://localhost:9000
        REMOTE_ACL: private

        # Best practice is to move sensitive information into `kind: Secret`
        # and reference that secret in `envFromSecret`
        # to keep your secret in GIT you can try e.g. SealedSecrets or ExternalSecrets
        #REMOTE_ACCESS_KEY_ID: ...
        #REMOTE_SECRET_ACCESS_KEY: ...
    envFromSecret:
        - ref: cloud-press-secret-envs

    # Optional
    # Will generate a key, store it in `kind: Secret` and setup End-To-End encryption
    # if existing secret exists and is valid, then will be reused
    automaticEncryption:
        enabled: true
        secretName: cloud-press-remote-sync
```
