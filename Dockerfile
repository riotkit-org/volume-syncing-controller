# =====================================================
# Take rclone from a selected released version artifact
# =====================================================
FROM rclone/rclone:1.58.1 as rcloneSrc


# ================
# Build entrypoint
# ================
FROM golang:1.18.2-alpine AS operatorBuilder

ADD ./ /build
WORKDIR /build

RUN apk add make
RUN make build
RUN mkdir -p /etc/volume-syncing-operator /mnt \
    && chown 65312:65312 /etc/volume-syncing-operator /mnt \
    && chmod 777 /etc/volume-syncing-operator /mnt


# =========================
# Create a distroless image
# =========================
FROM scratch
COPY --from=rcloneSrc /usr/local/bin/rclone /usr/bin/rclone
COPY --from=operatorBuilder /build/.build/volume-syncing-operator /usr/bin/volume-syncing-operator
COPY --from=operatorBuilder /etc/volume-syncing-operator /etc/volume-syncing-operator

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
