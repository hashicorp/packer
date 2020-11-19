package classic

import (
	"fmt"

	"github.com/hashicorp/packer/helper/communicator"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer/packer-plugin-sdk/template/interpolate"
)

const imageListDefault = "/oracle/public/OL_7.2_UEKR4_x86_64"
const imageListEntryDefault = 5
const usernameDefault = "opc"
const shapeDefault = "oc3"
const uploadImageCommandDefault = `
# https://www.oracle.com/webfolder/technetwork/tutorials/obe/cloud/objectstorage/upload_files_gt_5GB_REST_API/upload_files_gt_5GB_REST_API.html

# Split diskimage in to 100mb chunks
split -b 100m diskimage.tar.gz segment_
printf "Split diskimage into %s segments\n" $(ls segment_* | wc -l)

# Download jq tool
curl -OLs https://github.com/stedolan/jq/releases/download/jq-1.5/jq-linux64
mv jq-linux64 jq
chmod u+x jq

# Create manifest file
(
for i in segment_*; do
  ./jq -n --arg path "{{.SegmentPath}}/$i" \
          --arg etag $(md5sum $i | cut -f1 -d' ') \
          --arg size_bytes $(stat --printf "%s" $i) \
          '{path: $path, etag: $etag, size_bytes: $size_bytes}'
done
) | ./jq -s . > manifest.json

# Authenticate
curl -D auth-headers -s -X GET \
    -H "X-Storage-User: Storage-{{.AccountID}}:{{.Username}}" \
    -H "X-Storage-Pass: {{.Password}}" \
    https://{{.AccountID}}.storage.oraclecloud.com/auth/v1.0

export AUTH_TOKEN=$(awk 'BEGIN {FS=": "; RS="\r\n"}/^X-Auth-Token/{print $2}' auth-headers)
export STORAGE_URL=$(awk 'BEGIN {FS=": "; RS="\r\n"}/^X-Storage-Url/{print $2}' auth-headers)

# Upload segments
for i in segment_*; do
    echo "Uploading segment $i"

    curl -s -X PUT -T $i \
        -H "X-Auth-Token: $AUTH_TOKEN" ${STORAGE_URL}/{{.SegmentPath}}/$i;
done

# Create machine image from manifest
curl -s -X PUT \
    -H "X-Auth-Token: $AUTH_TOKEN" \
    "${STORAGE_URL}/compute_images/{{.ImageFile}}?multipart-manifest=put" \
    -T ./manifest.json

# Get uploaded image description
echo "Details of ${STORAGE_URL}/compute_images/{{.ImageFile}}:"
curl -I -X HEAD \
    -H "X-Auth-Token: $AUTH_TOKEN" \
    "${STORAGE_URL}/compute_images/{{.ImageFile}}"
`

type PVConfig struct {
	// PersistentVolumeSize lets us control the volume size by using persistent boot storage
	PersistentVolumeSize      int    `mapstructure:"persistent_volume_size"`
	BuilderUploadImageCommand string `mapstructure:"builder_upload_image_command"`

	// Builder Image
	BuilderShape          string `mapstructure:"builder_shape"`
	BuilderImageList      string `mapstructure:"builder_image_list"`
	BuilderImageListEntry *int   `mapstructure:"builder_image_list_entry"`

	BuilderComm communicator.Config `mapstructure:"builder_communicator"`
}

// IsPV tells us if we're using a persistent volume for this build
func (c *PVConfig) IsPV() bool {
	return c.PersistentVolumeSize > 0
}

func (c *PVConfig) Prepare(ctx *interpolate.Context) (errs *packersdk.MultiError) {
	if !c.IsPV() {
		if c.BuilderShape != "" {
			errs = packersdk.MultiErrorAppend(errs,
				fmt.Errorf("`builder_shape` has no meaning when `persistent_volume_size` is not set."))
		}
		if c.BuilderImageList != "" {
			errs = packersdk.MultiErrorAppend(errs,
				fmt.Errorf("`builder_image_list` has no meaning when `persistent_volume_size` is not set."))
		}

		if c.BuilderImageListEntry != nil {
			errs = packersdk.MultiErrorAppend(errs,
				fmt.Errorf("`builder_image_list_entry` has no meaning when `persistent_volume_size` is not set."))
		}
		return errs
	}

	c.BuilderComm.SSHPty = true
	if c.BuilderComm.Type == "winrm" {
		errs = packersdk.MultiErrorAppend(errs,
			fmt.Errorf("`ssh` is the only valid builder communicator type."))
	}

	if c.BuilderShape == "" {
		c.BuilderShape = shapeDefault
	}

	if c.BuilderComm.SSHUsername == "" {
		c.BuilderComm.SSHUsername = usernameDefault
	}

	if c.BuilderImageList == "" {
		c.BuilderImageList = imageListDefault
	}

	// Set to known working default if this is unset and we're using the
	// default image list
	if c.BuilderImageListEntry == nil {
		var entry int
		if c.BuilderImageList == imageListDefault {
			entry = imageListEntryDefault
		}
		c.BuilderImageListEntry = &entry
	}

	if c.BuilderUploadImageCommand == "" {
		c.BuilderUploadImageCommand = uploadImageCommandDefault
	}

	if es := c.BuilderComm.Prepare(ctx); len(es) > 0 {
		errs = packersdk.MultiErrorAppend(errs, es...)
	}

	return
}
