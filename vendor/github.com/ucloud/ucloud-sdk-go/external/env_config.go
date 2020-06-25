package external

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

const (
	UCloudPublicKeyEnvVar = "UCLOUD_PUBLIC_KEY"

	UCloudPrivateKeyEnvVar = "UCLOUD_PRIVATE_KEY"

	UCloudProjectIdEnvVar = "UCLOUD_PROJECT_ID"

	UCloudRegionEnvVar = "UCLOUD_REGION"

	UCloudZoneEnvVar = "UCLOUD_ZONE"

	UCloudAPIBaseURLEnvVar = "UCLOUD_API_BASE_URL"

	UCloudTimeoutSecondEnvVar = "UCLOUD_TIMEOUT_SECOND"

	UCloudSharedProfileEnvVar = "UCLOUD_PROFILE"

	UCloudSharedConfigFileEnvVar = "UCLOUD_SHARED_CONFIG_FILE"

	UCloudSharedCredentialFileEnvVar = "UCLOUD_SHARED_CREDENTIAL_FILE"
)

func loadEnvConfig() (*config, error) {
	cfg := &config{
		PublicKey:            os.Getenv(UCloudPublicKeyEnvVar),
		PrivateKey:           os.Getenv(UCloudPrivateKeyEnvVar),
		ProjectId:            os.Getenv(UCloudProjectIdEnvVar),
		Region:               os.Getenv(UCloudRegionEnvVar),
		Zone:                 os.Getenv(UCloudZoneEnvVar),
		BaseUrl:              os.Getenv(UCloudAPIBaseURLEnvVar),
		SharedConfigFile:     os.Getenv(UCloudSharedConfigFileEnvVar),
		SharedCredentialFile: os.Getenv(UCloudSharedCredentialFileEnvVar),
		Profile:              os.Getenv(UCloudSharedProfileEnvVar),
	}

	durstr, ok := os.LookupEnv(UCloudTimeoutSecondEnvVar)
	if ok {
		durnum, err := strconv.Atoi(durstr)
		if err != nil {
			return nil, fmt.Errorf("parse environment variable UCLOUD_TIMEOUT_SECOND [%s] error : %v", durstr, err)
		}
		cfg.Timeout = time.Second * time.Duration(durnum)
	}
	return cfg, nil
}
