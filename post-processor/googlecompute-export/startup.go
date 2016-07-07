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

echo "####### Export configuration #######"
echo "Image name - ${IMAGENAME}"
echo "Instance name - ${NAME}"
echo "Instance zone - ${ZONE}"
echo "Disk name - ${DISKNAME}"
echo "Export paths - ${PATHS}"
echo "####################################"

echo "Creating disk from image to be exported..."
gcloud compute disks create ${DISKNAME} --image ${IMAGENAME} --zone ${ZONE}
echo "Attaching disk..."
gcloud compute instances attach-disk ${NAME} --disk ${DISKNAME} --device-name toexport --zone ${ZONE}

echo "Dumping disk..."
dd if=/dev/disk/by-id/google-toexport of=disk.raw bs=4096 conv=sparse
echo "Compressing and tar'ing disk image..."
tar -czf root.tar.gz disk.raw

echo "Detaching disk..."
gcloud compute instances detach-disk ${NAME} --disk ${DISKNAME} --zone ${ZONE}
echo "Deleting disk..."
gcloud compute disks delete ${DISKNAME} --zone ${ZONE}

for i in ${PATHS}; do
  echo "Uploading tar'ed disk image to ${i}..."
  gsutil -o GSUtil:parallel_composite_upload_threshold=100M cp root.tar.gz ${i}
  LOGDEST="${i}.exporter.log"
  echo "Uploading exporter log to ${LOGDEST}..."
  gsutil -h "Content-Type:text/plain" cp /var/log/daemon.log ${LOGDEST}
done
`
