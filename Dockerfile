# =====================================================
# Take rclone from a selected released version artifact
# =====================================================
FROM rclone/rclone:1.62.2 as rcloneSrc


# ================
# Build entrypoint
# ================
FROM alpine:3.17 AS workspaceBuilder

RUN mkdir -p /etc/volume-syncing-controller /mnt /run \
    && touch /etc/volume-syncing-controller/rclone.conf /run/volume-syncing-controller.pid \
    && chown -R 65312:65312 /etc/volume-syncing-controller /mnt /run \
    && chmod -R 777 /etc/volume-syncing-controller /mnt /run

# make sure the permissions are correct
COPY ./.build/volume-syncing-controller /usr/bin/volume-syncing-controller
RUN chmod +x /usr/bin/volume-syncing-controller && chown root:root /usr/bin/volume-syncing-controller && chmod 755 /usr/bin/volume-syncing-controller


# =========================
# Create a distroless image
# =========================
FROM scratch

# copy ca-certificates to work with some apis
COPY --from=rcloneSrc /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
# copy a versioned artifact from official released image
COPY --from=rcloneSrc /usr/local/bin/rclone /usr/bin/rclone
# copy already built artifact by CI
COPY --from=workspaceBuilder /usr/bin/volume-syncing-controller /usr/bin/volume-syncing-controller
# copy a directory with prepared permissions
COPY --from=workspaceBuilder /etc/volume-syncing-controller /etc/volume-syncing-controller
COPY --from=workspaceBuilder /run /run

ENV REMOTE_TYPE="s3"
ENV PATH="/usr/bin"

# Environment variables are DYNAMIC, depending on desired rclone configuration
# example:
# ENV REMOTE_PROVIDER=Minio
# ENV REMOTE_ACCESS_KEY_ID=AKIAIOSFODNN7EXAMPLE
# ENV REMOTE_SECRET_ACCESS_KEY=wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY
# ENV REMOTE_ENDPOINT=http://localhost:9000
# ENV REMOTE_ACL=private
#
# Read more: https://rclone.org/overview/

USER 65312
WORKDIR /mnt
ENTRYPOINT ["/usr/bin/volume-syncing-controller"]
