volume-syncing-controller
=========================

Docker container and Kubernetes controller for periodically synchronizing volumes to cloud-native storage, and restoring their state from cloud-native storage.

Example use cases
-----------------

- Service restored from cloud-storage in other data center with its own CNI
- Kubernetes service state kept in cloud-storage, restored outside Kubernetes (e.g. in docker on desktop for development)
- Snapshot type backup (does not support versioning on client side - the cloud remote storage could do this)
- Ability to move services between machines in tiny K3s clusters, where network CNI is not available - only local storage is used

Roadmap
-------

**Features for first version:**
- [x] Supports all storage kinds that are supported by Rclone
- [x] Rclone configuration using environment variables
- [x] End-To-End encryption support
- [x] Periodical synchronization with built-in cron-like scheduler
- [x] `volume-syncing-controller sync-to-remote` command to synchronize local files to remote
- [x] `volume-syncing-controller remote-to-local-sync` command to sync files back from remote to local
- [x] Support for Kubernetes: **initContainer** to `restore` files, and **side-car** to back up files to remote
- [x] Extra security layer preventing from accidental file deletion in comparison to plain `rclone` or `rsync` usage :100:
- [x] Non-root container
- [x] Allow to disable synchronization or restore in CRD
- [x] Jinja2 templating support inside `kind: PodFilesystemSync` to allow using single definition for multiple `kind: Pod` objects
- [x] Termination hook to synchronize Pod before it gets terminated
- [x] Allow to decide about the order of initContainer in CRD

**v1.1:**
- [ ] Health check: If N-synchronization fails, then mark Pod as unhealthy (optionally)
- [ ] Periodical synchronization using filesystem events instead of cron-like scheduler (both should be available)

**v1.2:**
- [ ] Watch synchronization progress and update status field. Pods can notify controller using webhooks with a token granted on Pod creation
- [ ] Multiple application replicas support: One Pod can be marked as primary, the rest will have sync-to-remote on hold
- [ ] Expose Prometheus metrics at /metrics, add configuration for Prometheus and Victoria Metrics in Helm


Kubernetes controller architecture
----------------------------------

The solution architecture is designed to be Pod-centric and live together with the application, not on the underlying infrastructural level.

Above means, that when Pod starts - the volume is **restored from remote**, then all data is **synchronized to remote periodically** during Pod lifetime to keep an external storage up-to-date.

There are added two containers:
- **init-volume-restore** (initContainer - executes before application starts)
- **volume-syncing-sidecar** (sidecar - lives together with application)

Security
--------

- The images are scanned for vulnerabilities: [Open report](https://artifacthub.io/packages/helm/riotkit-org/volume-syncing-controller?modal=security-report)
- Distroless-type image reduces surface of attack
- No bash scripts. The container contains only `rclone` and `volume-syncing-controller` binaries
- Built-in E2E encryption automation
- Semantic Versioning is used to protect end users from breaking changes as much as it is possible

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
volume-syncing-controller sync-to-remote -s ./ -d testbucket/some-directory

# synchronize back from remote storage to local directory
volume-syncing-controller remote-to-local-sync -v -s testbucket -d ./.build/testing-restore
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

`volume-syncing-controller` has built-in crontab-like scheduler to optionally enable periodical syncing.

```
volume-syncing-controller --schedule '@every 1m'    # ...
volume-syncing-controller --schedule '0 30 * * * *' # ...
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

Operator have to be installed using Helm first. It will react for every Pod labelled with `riotkit.org/volume-syncing-controller: true` and matching CRD of `PodFilesystemSync` kind.

The `riotkit.org/volume-syncing-controller: true` is there for performance and cluster safety by limiting the scope of Admission Webhook.

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

    initContainerPlacement:
        containerReference: some-container-name # defaults to empty
        placement: before # before, after, first or last (first & last requires the `containerReference` to be empty). Defaults to "last"

    syncOptions:
        # NOTICE: every next synchronization will be cancelled if previous one was not finished
        method: "scheduler"  # or "fsnotify"
        schedule: "@every 5m"
        maxOneSyncPerMinutes: "15"  # When "fsnotify" used, then perform only max one sync per N minutes. Allows to decrease network/cpu/disk usage with a little risk factor

        # Optional "RunAs"
        permissions:
            # Can be overridden by Pod annotation `riotkit.org/volume-user-id`
            uid: "1001"
            # Can be overridden by Pod annotation `riotkit.org/volume-group-id`
            gid: "1001"
            
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
        allowedDirections:
            # Decides if a side-car container should be spawned to synchronize changes periodically TO REMOTE
            toRemote: true
            # Decides if an init container should be placed to restore files FROM REMOTE on startup
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
    envFromSecrets:
        - name: cloud-press-secret-envs

    # Optional
    # Will generate a key, store it in `kind: Secret` and setup End-To-End encryption
    # if existing secret exists and is valid, then will be reused
    #
    # NOTICE: REMEMBER TO BACKUP THIS secret. End-To-End encryption means your data will be unreadable on the server
    automaticEncryption:
        enabled: true
        secretName: cloud-press-remote-sync
```

SIGTERM support
---------------

Pod's sidecar container named `init-volume-restore` is having implemented a signal handling. 
When Kubernetes wants to terminate our Pod, then a `volume-syncing-controller interrupt` command is invoked, next it sends a kill signal to the `volume-syncing-controller` main process.

`volume-syncing-controller` main process is stopping the cron-like scheduler and invokes last synchronization to remote before exit.

You may want to adjust your Pod's `terminationGracePeriodSeconds` to a value that makes sure the Kubernetes will wait longer before terminating containers.

Permissions
-----------

UID and GID can be specified in the `PodFilesystemSync` resource as well as in `Pod` annotations.

**PodFilesystemSync example**

```yaml
# ...
spec:
    syncOptions:
        # ...
        permissions:
            # Can be overridden by Pod annotation `riotkit.org/volume-user-id`
            uid: 1001
            # Can be overridden by Pod annotation `riotkit.org/volume-group-id`
            gid: 1001
```

**Pod example**

```yaml
# ...
metadata:
    annotations:
        riotkit.org/volume-user-id: 161
        riotkit.org/volume-group-id: 1312
```

The annotations have precedence over PodFilesystemSync resource settings.

PodFilesystemSync deletion
--------------------------

Deletion if a definition does not recursively delete secrets that were automatically created for encryption.

Triggering synchronization manually
-----------------------------------

To perform a synchronization in Kubernetes when **automatic encryption** is turned on, type:

```bash
kubectl exec -n my-namespace-name my-pod-name -c volume-syncing-sidecar -it -- /usr/bin/volume-syncing-controller sync-to-remote -s /path/to/local/directory -c /etc/volume-syncing-controller/rclone.conf
```

In case, when **automatic encryption is not used**, then you additionally need to specify the remote target directory.
