// CODE GENERATED. DO NOT EDIT
package yandexexport

var (
	CloudInitScript = `#!/usr/bin/env bash

GetMetadata() {
    curl -f -H "Metadata-Flavor: Google" http://169.254.169.254/computeMetadata/v1/instance/attributes/$1 2>/dev/null
}

GetInstanceId() {
    curl -f -H "Metadata-Flavor: Google" http://169.254.169.254/computeMetadata/v1/instance/id 2>/dev/null
}

GetServiceAccountId() {
    yc compute instance get "${INSTANCE_ID}" | grep service_account | cut -f2 -d' '
}

InstallYc() {
    curl -s "${S3_ENDPOINT}/yandexcloud-yc/install.sh" | sudo bash -s -- -n -i /usr/local
}

InstallAwsCli() {
    curl "https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip" -o "awscliv2.zip"
    unzip -o awscliv2.zip >/dev/null
    sudo ./aws/install
}

InstallPackages() {
    sudo apt-get update -qq && sudo apt-get install -y unzip jq qemu-utils
}

InstallTools() {
    InstallPackages
    InstallYc
    InstallAwsCli
}

INSTANCE_ID=$(GetInstanceId)
PATHS=$(GetMetadata paths)
S3_ENDPOINT="https://storage.yandexcloud.net"

Exit() {
    for i in ${PATHS}; do
        LOGDEST="${i}.exporter.log"
        echo "Uploading exporter log to ${LOGDEST}..."
        aws s3 --region ru-central1 --endpoint-url="${S3_ENDPOINT}" cp /var/log/syslog "${LOGDEST}"
    done

    echo "Delete static access key..."
    if ! yc iam access-key delete "${YC_SK_ID}"; then
        echo "Failed to delete static access key."
        FAIL=1
    fi

    if [ $1 -ne 0 ]; then
        echo "Set metadata key 'cloud-init-status' to 'cloud-init-error' value"
        if ! yc compute instance update "${INSTANCE_ID}" --metadata cloud-init-status=cloud-init-error; then
            echo "Failed to update metadata key 'cloud-init-status'."
            exit 111
        fi
    fi

    exit $1
}

InstallTools

echo "####### Export configuration #######"
echo "Instance ID - ${INSTANCE_ID}"
echo "Export paths - ${PATHS}"
echo "####################################"

echo "Detect Service Account ID..."
SERVICE_ACCOUNT_ID=$(GetServiceAccountId)
echo "Use Service Account ID: ${SERVICE_ACCOUNT_ID}"

echo "Create static access key..."
SEC_json=$(yc iam access-key create --service-account-id "${SERVICE_ACCOUNT_ID}" \
    --description "this key is for export image to storage" --format json)

if [ $? -ne 0 ]; then
    echo "Failed to create static access key."
    Exit 1
fi

echo "Setup env variables to access storage..."
eval "$(jq -r '@sh "export YC_SK_ID=\(.access_key.id); export AWS_ACCESS_KEY_ID=\(.access_key.key_id); export AWS_SECRET_ACCESS_KEY=\(.secret)"' <<<${SEC_json})"

for i in ${PATHS}; do
    bucket=$(echo "${i}" | sed 's/\(s3:\/\/[^\/]*\).*/\1/')
    echo "Check access to storage: '${bucket}'..."
    if ! aws s3 --region ru-central1 --endpoint-url="${S3_ENDPOINT}" ls "${bucket}" >/dev/null; then
        echo "Failed to access storage: '${bucket}'."
        Exit 1
    fi
done

echo "Dumping disk..."
if ! qemu-img convert -O qcow2 -o cluster_size=2M /dev/disk/by-id/virtio-doexport disk.qcow2; then
    echo "Failed to dump disk to qcow2 image."
    Exit 1
fi

for i in ${PATHS}; do
    echo "Uploading qcow2 disk image to ${i}..."
    if ! aws s3 --region ru-central1 --endpoint-url="${S3_ENDPOINT}" cp disk.qcow2 "${i}"; then
        echo "Failed to upload image to ${i}."
        FAIL=1
    fi
done

echo "Set metadata key 'cloud-init-status' to 'cloud-init-done' value"
if ! yc compute instance update "${INSTANCE_ID}" --metadata cloud-init-status=cloud-init-done; then
    echo "Failed to update metadata key to 'cloud-init-status'."
    Exit 1
fi

Exit ${FAIL}
`
)
