package googlecomputeexport

import (
	"fmt"

	"github.com/hashicorp/packer-plugin-googlecompute/builder/googlecompute"
)

var StartupScript string = fmt.Sprintf(`#!/bin/bash

GetMetadata () {
  echo "$(curl -f -H "Metadata-Flavor: Google" http://metadata/computeMetadata/v1/instance/attributes/$1 2> /dev/null)"
}

ZONE=$(basename $(GetMetadata zone))

SetMetadata () {
  gcloud compute instances add-metadata ${HOSTNAME} --metadata ${1}=${2} --zone ${ZONE}
}

STARTUPSCRIPT=$(GetMetadata attributes/%s)
STARTUPSCRIPTPATH=/packer-wrapped-startup-script
if [ -f "/var/log/startupscript.log" ]; then
  STARTUPSCRIPTLOGPATH=/var/log/startupscript.log
else
  STARTUPSCRIPTLOGPATH=/var/log/daemon.log
fi
STARTUPSCRIPTLOGDEST=$(GetMetadata attributes/startup-script-log-dest)

IMAGENAME=$(GetMetadata image_name)
NAME=$(GetMetadata name)
DISKNAME=${NAME}-toexport
PATHS=($(GetMetadata paths))

Exit () {
  for i in ${PATHS[@]}; do
    LOGDEST="${i}.exporter.log"
    echo "Uploading exporter log to ${LOGDEST}..."
    gsutil -h "Content-Type:text/plain" cp /var/log/daemon.log ${LOGDEST}
  done
  exit $1
}

echo "####### Export configuration #######"
echo "Image name - ${IMAGENAME}"
echo "Instance name - ${NAME}"
echo "Instance zone - ${ZONE}"
echo "Disk name - ${DISKNAME}"
echo "Export paths - ${PATHS}"
echo "####################################"

echo "Creating disk from image to be exported..."
if ! gcloud compute disks create ${DISKNAME} --image ${IMAGENAME} --zone ${ZONE}; then
  echo "Failed to create disk."
  Exit 1
fi

echo "Attaching disk..."
if ! gcloud compute instances attach-disk ${NAME} --disk ${DISKNAME} --device-name toexport --zone ${ZONE}; then
  echo "Failed to attach disk."
  Exit 1
fi

echo "GCEExport: Running export tool."
gce_export -gcs_path "${PATHS[0]}" -disk /dev/disk/by-id/google-toexport -y
if [ $? -ne 0 ]; then
  echo "ExportFailed: Failed to export disk source to ${PATHS[0]}."
  Exit 1
fi

echo "ExportSuccess"
sync

echo "Detaching disk..."
if ! gcloud compute instances detach-disk ${NAME} --disk ${DISKNAME} --zone ${ZONE}; then
  echo "Failed to detach disk."
fi

FAIL=0
echo "Deleting disk..."
if ! gcloud compute disks delete ${DISKNAME} --zone ${ZONE}; then
  echo "Failed to delete disk."
  FAIL=1
fi

for i in ${PATHS[@]:1}; do
  echo "Copying archive image to ${i}..."
  if ! gsutil -o GSUtil:parallel_composite_upload_threshold=100M cp ${PATHS[0]} ${i}; then
    echo "Failed to copy image to ${i}."
    FAIL=1
  fi
done

SetMetadata %s %s

Exit ${FAIL}
`, googlecompute.StartupWrappedScriptKey, googlecompute.StartupScriptStatusKey, googlecompute.StartupScriptStatusDone)
