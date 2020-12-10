#!/usr/bin/env bash

GetMetadata() {
    curl -f -H "Metadata-Flavor: Google" http://169.254.169.254/computeMetadata/v1/instance/attributes/$1 2>/dev/null
}

[[ "$(GetMetadata debug)" == "1"  || "$(GetMetadata debug)" == "true" ]] && set -x

InstallPackages() {
    sudo apt-get update -qq && sudo apt-get install -y qemu-utils awscli
}

WaitFile() {
    local RETRIES=60
    while [[ ${RETRIES} -gt 0 ]]; do
        echo "Wait ${1}"
        if [ -e "${1}" ]; then
            echo "[${1}] has been found"
            return 0
        fi
        RETRIES=$((RETRIES-1))
        sleep 5
    done
    echo "[${1}] not found"
    return 1
}

PATHS=$(GetMetadata paths)
S3_ENDPOINT="https://storage.yandexcloud.net"
DISK_EXPORT_PATH="/dev/disk/by-id/virtio-doexport"
export AWS_SHARED_CREDENTIALS_FILE="/tmp/aws-credentials"
export AWS_REGION=ru-central1

Exit() {
    for i in ${PATHS}; do
        LOGDEST="${i}.exporter.log"
        echo "Uploading exporter log to ${LOGDEST}..."
        aws s3 --endpoint-url="${S3_ENDPOINT}" cp /var/log/syslog "${LOGDEST}"
    done

    exit $1
}

InstallPackages

echo "####### Export configuration #######"
echo "Export paths - ${PATHS}"
echo "####################################"

if ! WaitFile "${AWS_SHARED_CREDENTIALS_FILE}"; then
    echo "Failed wait credentials"
    Exit 1
fi
udevadm trigger || true

if ! WaitFile "${DISK_EXPORT_PATH}"; then
    echo "Failed wait attach disk"
    Exit 1
fi

echo "Dumping disk..."
if ! qemu-img convert -p -O qcow2 -o cluster_size=2M "${DISK_EXPORT_PATH}" disk.qcow2; then
    echo "Failed to dump disk to qcow2 image."
    Exit 1
fi

for i in ${PATHS}; do
    echo "Uploading qcow2 disk image to ${i}..."
    if ! aws s3 --endpoint-url="${S3_ENDPOINT}" cp disk.qcow2 "${i}"; then
        echo "Failed to upload image to ${i}."
        FAIL=1
    fi
done

Exit ${FAIL}
