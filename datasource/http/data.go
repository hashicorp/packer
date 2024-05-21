// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

//go:generate packer-sdc struct-markdown
//go:generate packer-sdc mapstructure-to-hcl2 -type DatasourceOutput,Config
package http

import (
	"context"
	"fmt"
	"io"
	"mime"
	"net/http"
	"regexp"
	"strings"

	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/packer-plugin-sdk/common"
	"github.com/hashicorp/packer-plugin-sdk/hcl2helper"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/template/config"
	"github.com/zclconf/go-cty/cty"
)

type Config struct {
	common.PackerConfig `mapstructure:",squash"`
	// The URL to request data from. This URL must respond with a `200 OK` response and a `text/*` or `application/json` Content-Type.
	Url string `mapstructure:"url" required:"true"`
	// A map of strings representing additional HTTP headers to include in the request.
	RequestHeaders map[string]string `mapstructure:"request_headers" required:"false"`
}

type Datasource struct {
	config Config
}

type DatasourceOutput struct {
	// The URL the data was requested from.
	Url string `mapstructure:"url"`
	// The raw body of the HTTP response.
	ResponseBody string `mapstructure:"body"`
	// A map of strings representing the response HTTP headers.
	// Duplicate headers are concatenated with, according to [RFC2616](https://www.w3.org/Protocols/rfc2616/rfc2616-sec4.html#sec4.2).
	ResponseHeaders map[string]string `mapstructure:"request_headers"`
}

func (d *Datasource) ConfigSpec() hcldec.ObjectSpec {
	return d.config.FlatMapstructure().HCL2Spec()
}

func (d *Datasource) Configure(raws ...interface{}) error {
	err := config.Decode(&d.config, nil, raws...)
	if err != nil {
		return err
	}

	var errs *packersdk.MultiError

	if d.config.Url == "" {
		errs = packersdk.MultiErrorAppend(
			errs,
			fmt.Errorf("the `url` must be specified"))
	}

	if errs != nil && len(errs.Errors) > 0 {
		return errs
	}
	return nil
}

func (d *Datasource) OutputSpec() hcldec.ObjectSpec {
	return (&DatasourceOutput{}).FlatMapstructure().HCL2Spec()
}

// This is to prevent potential issues w/ binary files
// and generally unprintable characters
// See https://github.com/hashicorp/terraform/pull/3858#issuecomment-156856738
func isContentTypeText(contentType string) bool {

	parsedType, params, err := mime.ParseMediaType(contentType)
	if err != nil {
		return false
	}

	allowedContentTypes := []*regexp.Regexp{
		regexp.MustCompile("^text/.+"),
		regexp.MustCompile("^application/json$"),
		regexp.MustCompile("^application/samlmetadata\\+xml"),
	}

	for _, r := range allowedContentTypes {
		if r.MatchString(parsedType) {
			charset := strings.ToLower(params["charset"])
			return charset == "" || charset == "utf-8" || charset == "us-ascii"
		}
	}

	return false
}

// Most of this code comes from http terraform provider data source
// https://github.com/hashicorp/terraform-provider-http/blob/main/internal/provider/data_source.go
func (d *Datasource) Execute() (cty.Value, error) {
	ctx := context.TODO()
	url, headers := d.config.Url, d.config.RequestHeaders
	client := &http.Client{}

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	// TODO: How to make a test case for this?
	if err != nil {
		fmt.Println("Error creating http request")
		return cty.NullVal(cty.EmptyObject), err
	}

	for name, value := range headers {
		req.Header.Set(name, value)
	}

	resp, err := client.Do(req)
	// TODO: How to make test case for this
	if err != nil {
		fmt.Println("Error making performing http request")
		return cty.NullVal(cty.EmptyObject), err
	}

	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return cty.NullVal(cty.EmptyObject), fmt.Errorf("HTTP request error. Response code: %d", resp.StatusCode)
	}

	contentType := resp.Header.Get("Content-Type")
	if contentType == "" || isContentTypeText(contentType) == false {
		fmt.Println(fmt.Sprintf(
			"Content-Type is not recognized as a text type, got %q",
			contentType))
		fmt.Println("If the content is binary data, Packer may not properly handle the contents of the response.")
	}

	bytes, err := io.ReadAll(resp.Body)
	// TODO: How to make test case for this?
	if err != nil {
		fmt.Println("Error processing response body of call")
		return cty.NullVal(cty.EmptyObject), err
	}

	responseHeaders := make(map[string]string)
	for k, v := range resp.Header {
		// Concatenate according to RFC2616
		// cf. https://www.w3.org/Protocols/rfc2616/rfc2616-sec4.html#sec4.2
		responseHeaders[k] = strings.Join(v, ", ")
	}

	output := DatasourceOutput{
		Url:             d.config.Url,
		ResponseHeaders: responseHeaders,
		ResponseBody:    string(bytes),
	}
	return hcl2helper.HCL2ValueFromConfig(output, d.OutputSpec()), nil
}
