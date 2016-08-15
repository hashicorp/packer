package googlecomputeexport

var StartupScript string = `#!/bin/sh

GetMetadata () {
  echo "$(curl -f -H "Metadata-Flavor: Google" http://metadata/computeMetadata/v1/instance/attributes/$1 2> /dev/null)"
}
IMAGENAME=$(GetMetadata image_name)
NAME=$(GetMetadata name)
DISKNAME=${NAME}-toexport
PATHS=$(GetMetadata paths)
ZONE=$(GetMetadata zone)

Exit () {
  for i in ${PATHS}; do
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

echo "Dumping disk..."
if ! dd if=/dev/disk/by-id/google-toexport of=disk.raw bs=4096 conv=sparse; then
  echo "Failed to dump disk to image."
  Exit 1
fi

echo "Compressing and tar'ing disk image..."
if ! tar -czf root.tar.gz disk.raw; then
  echo "Failed to tar disk image."
  Exit 1
fi

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

for i in ${PATHS}; do
  echo "Uploading tar'ed disk image to ${i}..."
  if ! gsutil -o GSUtil:parallel_composite_upload_threshold=100M cp root.tar.gz ${i}; then
    echo "Failed to upload image to ${i}."
    FAIL=1
  fi
done

Exit ${FAIL}
`
