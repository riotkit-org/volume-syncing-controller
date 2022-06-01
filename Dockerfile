# =====================================================
# Take rclone from a selected released version artifact
# =====================================================
FROM rclone/rclone:1.58.1 as rcloneSrc


# ================
# Build entrypoint
# ================
FROM alpine:3.16 AS workspaceBuilder

RUN mkdir -p /etc/volume-syncing-operator /mnt \
    && chown 65312:65312 /etc/volume-syncing-operator /mnt \
    && chmod 777 /etc/volume-syncing-operator /mnt


# =========================
# Create a distroless image
# =========================
FROM scratch

# copy a versioned artifact from official released image
COPY --from=rcloneSrc /usr/local/bin/rclone /usr/bin/rclone
# copy already built artifact by CI
COPY ./.build/volume-syncing-operator /usr/bin/volume-syncing-operator
# copy a directory with prepared permissions
COPY --from=workspaceBuilder /etc/volume-syncing-operator /etc/volume-syncing-operator

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
ENTRYPOINT ["/usr/bin/volume-syncing-operator"]
