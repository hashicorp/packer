// Copyright IBM Corp. 2024, 2026
// SPDX-License-Identifier: BUSL-1.1

package hcl2template

import (
	"strconv"
	"testing"

	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
	"github.com/zclconf/go-cty/cty"
)

func TestRenderLegacyTimeFunctions(t *testing.T) {
	timestamp := strconv.FormatInt(interpolate.InitTime.Unix(), 10)

	tests := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "timestamp",
			in:   "image-{{timestamp}}",
			want: "image-" + timestamp,
		},
		{
			name: "timestamp with spaces",
			in:   "{{ timestamp }}",
			want: timestamp,
		},
		{
			name: "isotime",
			in:   `{{isotime "2006-01-02"}}`,
			want: interpolate.InitTime.Format("2006-01-02"),
		},
		{
			name: "strftime",
			in:   `{{strftime "%Y"}}`,
			want: interpolate.InitTime.Format("2006"),
		},
		{
			name: "other template functions are kept",
			in:   "{{build `ID`}}-{{timestamp}}",
			want: "{{build `ID`}}-" + timestamp,
		},
		{
			name: "no template expression",
			in:   "plain",
			want: "plain",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := renderLegacyTimeFunctions(test.in); got != test.want {
				t.Fatalf("wrong result\ngot:  %q\nwant: %q", got, test.want)
			}
		})
	}
}

type timeFuncTestDecodable struct{}

func (timeFuncTestDecodable) ConfigSpec() hcldec.ObjectSpec {
	return hcldec.ObjectSpec{
		"value": &hcldec.AttrSpec{
			Name:     "value",
			Type:     cty.String,
			Required: true,
		},
	}
}

func TestDecodeHCL2SpecRendersTimeFunctions(t *testing.T) {
	file, diags := hclparse.NewParser().ParseHCL([]byte(`value = "image-{{timestamp}}"`), "test.hcl")
	if diags.HasErrors() {
		t.Fatalf("unexpected error: %s", diags.Error())
	}

	decoded, diags := decodeHCL2Spec(file.Body, nil, timeFuncTestDecodable{})
	if diags.HasErrors() {
		t.Fatalf("unexpected error: %s", diags.Error())
	}

	want := "image-" + strconv.FormatInt(interpolate.InitTime.Unix(), 10)
	if got := decoded.GetAttr("value").AsString(); got != want {
		t.Fatalf("wrong result\ngot:  %q\nwant: %q", got, want)
	}
}
