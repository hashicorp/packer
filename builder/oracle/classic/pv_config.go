package classic

import (
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/template/interpolate"
)

type PVConfig struct {
	// PersistentVolumeSize lets us control the volume size by using persistent boot storage
	PersistentVolumeSize      int    `mapstructure:"persistent_volume_size"`
	BuilderImageList          string `mapstructure:"builder_image_list"`
	BuilderUploadImageCommand string `mapstructure:"builder_upload_image_command"`
	/* TODO:
	default to OL image
	make sure if set then PVS is above
	some way to choose which connection to use for master
	possible ignore everything for builder and always use SSH keys
	*/
}

func (c *PVConfig) IsPV() bool {
	return c.PersistentVolumeSize > 0
}

func (c *PVConfig) Prepare(ctx *interpolate.Context) (errs *packer.MultiError) {
	if !c.IsPV() {
		return nil
	}

	if c.BuilderUploadImageCommand == "" {
		c.BuilderUploadImageCommand = `
# https://www.oracle.com/webfolder/technetwork/tutorials/obe/cloud/objectstorage/upload_files_gt_5GB_REST_API/upload_files_gt_5GB_REST_API.html

# Split diskimage in to 100mb chunks
split -b 100m diskimage.tar.gz segment_

# Authenticate
curl -D auth-headers -s -X GET \
	-H "X-Storage-User: Storage-{{.AccountID}}:{{.Username}}" \
	-H "X-Storage-Pass: {{.Password}}" \
	https://{{.AccountID}}.storage.oraclecloud.com/auth/v1.0

export AUTH_TOKEN=$(awk 'BEGIN {FS=": "; RS="\r\n"}/^X-Auth-Token/{print $2}' auth-headers)
export STORAGE_URL=$(awk 'BEGIN {FS=": "; RS="\r\n"}/^X-Storage-Url/{print $2}' auth-headers)

# Create {{.Container}}
curl -v -X PUT -H "X-Auth-Token: $AUTH_TOKEN" ${STORAGE_URL}/{{.Container}}

# Upload segments
for i in segment_*; do
	curl -v -X PUT -T $i \
		-H "X-Auth-Token: $AUTH_TOKEN" \
		${STORAGE_URL}/{{.Container}}/$i;
done

# Download jq tool
curl -OL https://github.com/stedolan/jq/releases/download/jq-1.5/jq-linux64
mv jq-linux64 jq
chmod u+x jq

# Create manifest file
(
for i in segment_*; do
  ./jq -n --arg path "{{.Container}}/$i" \
          --arg etag $(md5sum $i | cut -f1 -d' ') \
		  --arg size_bytes $(stat --printf "%s" $i) \
		  '{path: $path, etag: $etag, size_bytes: $size_bytes}'
done
) | ./jq -s . > manifest.json

# Create machine image from manifest
curl -v -X PUT \
	-H "X-Auth-Token: $AUTH_TOKEN" \
	"${STORAGE_URL}/compute_images/{{.ImageName}}.tar.gz?multipart-manifest=put" \
	-T ./manifest.json

# Get uploaded image description
curl -I -X HEAD \
	-H "X-Auth-Token: $AUTH_TOKEN" \
	"${STORAGE_URL}/compute_images/{{.ImageName}}.tar.gz"


# Delete {{.Container}}
# following https://docs.oracle.com/en/cloud/iaas/storage-cloud/cssto/deleting-multiple-objects-single-operation.html

# Collect objects we need to clean up
printf "{{.Container}}/%s\n" segment_* > objects_to_delete.txt

# Delete objects
curl -v -s -X DELETE \
	-H "X-Auth-Token: $AUTH_TOKEN" \
	-H "Content-Type: text/plain" \
	-T objects_to_delete.txt \
	"${STORAGE_URL}/?bulk-delete"

# Delete container
curl -v -s -X DELETE \
	-H "X-Auth-Token: $AUTH_TOKEN" \
	"${STORAGE_URL}/{{.Container}}"
`
	}
	/*
		errs = packer.MultiErrorAppend(errs,
			fmt.Errorf("Persistent storage volumes are only supported on unix, and must use the ssh communicator."))
	*/
	return
}
