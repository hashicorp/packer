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
		c.BuilderUploadImageCommand = `split -b 10m diskimage.tar.gz segment_ 

curl -D auth-headers -s -X GET -H "X-Storage-User: Storage-a459477:jake@hashicorp.com" -H "X-Storage-Pass: ***REMOVED***" https://a459477.storage.oraclecloud.com/auth/v1.0

export AUTH_TOKEN=$(awk 'BEGIN {FS=": "; RS="\r\n"}/^X-Auth-Token/{print $2}' auth-headers)
export STORAGE_URL=$(awk 'BEGIN {FS=": "; RS="\r\n"}/^X-Storage-Url/{print $2}' auth-headers)

curl -v -X PUT -H "X-Auth-Token: $AUTH_TOKEN" ${STORAGE_URL}/mwhooker-test-1

for i in segment_*; do 
	curl -v -X PUT -T $i \
		-H "X-Auth-Token: $AUTH_TOKEN" \
		${STORAGE_URL}/mwhooker-test-1/$i;
done

curl -OL https://github.com/stedolan/jq/releases/download/jq-1.5/jq-linux64
mv jq-linux64 jq
chmod u+x jq

(
for i in segment_*; do
  ./jq -n --arg path "mwhooker-test-1/$i" --arg etag $(md5sum $i | cut -f1 -d' ') --arg size_bytes $(stat --printf "%s" $i) '{path: $path, etag: $etag, size_bytes: $size_bytes}'
done
) | ./jq -s . > manifest.json

curl -v -X PUT -H "X-Auth-Token: $AUTH_TOKEN"  "${STORAGE_URL}/compute_images/mwhooker-diskimage-01.tar.gz?multipart-manifest=put" -T ./manifest.json

# curl -I -X HEAD -H "X-Auth-Token: $AUTH_TOKEN"  "${STORAGE_URL}/mwhooker-test-1/diskimage.tar.gz"
'...'`
	}
	/*
		errs = packer.MultiErrorAppend(errs,
			fmt.Errorf("Persistent storage volumes are only supported on unix, and must use the ssh communicator."))
	*/
	return
}
