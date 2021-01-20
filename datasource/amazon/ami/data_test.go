package ami

import (
	"testing"

	awscommon "github.com/hashicorp/packer/builder/amazon/common"
)

func TestDatasourceConfigure_FilterBlank(t *testing.T) {
	datasource := Datasource{
		config: Config{
			AmiFilterOptions: awscommon.AmiFilterOptions{},
		},
	}
	if err := datasource.Configure(nil); err == nil {
		t.Fatalf("Should error if filters map is empty or not specified")
	}
}

func TestRunConfigPrepare_SourceAmiFilterOwnersBlank(t *testing.T) {
	datasource := Datasource{
		config: Config{
			AmiFilterOptions: awscommon.AmiFilterOptions{
				Filters: map[string]string{"foo": "bar"},
			},
		},
	}
	if err := datasource.Configure(nil); err == nil {
		t.Fatalf("Should error if Owners is not specified)")
	}
}

func TestRunConfigPrepare_SourceAmiFilterGood(t *testing.T) {
	datasource := Datasource{
		config: Config{
			AmiFilterOptions: awscommon.AmiFilterOptions{
				Owners:  []string{"1234567"},
				Filters: map[string]string{"foo": "bar"},
			},
		},
	}
	if err := datasource.Configure(nil); err != nil {
		t.Fatalf("err: %s", err)
	}
}
