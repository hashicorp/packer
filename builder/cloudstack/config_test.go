package cloudstack

import "testing"

func TestNewConfig(t *testing.T) {
	cases := map[string]struct {
		Config  map[string]interface{}
		Nullify string
		Err     bool
	}{
		"no_api_url": {
			Config: map[string]interface{}{
				"disk_size":       "20",
				"source_template": "d31e6af5-94a8-4756-abf3-6493c38db7e5",
			},
			Nullify: "api_url",
			Err:     true,
		},
		"no_api_key": {
			Config: map[string]interface{}{
				"disk_size":       "20",
				"source_template": "d31e6af5-94a8-4756-abf3-6493c38db7e5",
			},
			Nullify: "api_key",
			Err:     true,
		},
		"no_secret_key": {
			Config: map[string]interface{}{
				"disk_size":       "20",
				"source_template": "d31e6af5-94a8-4756-abf3-6493c38db7e5",
			},
			Nullify: "secret_key",
			Err:     true,
		},
		"no_cidr_list": {
			Config: map[string]interface{}{
				"disk_size":       "20",
				"source_template": "d31e6af5-94a8-4756-abf3-6493c38db7e5",
			},
			Nullify: "cidr_list",
			Err:     true,
		},
		"no_cidr_list_with_use_local_ip_address": {
			Config: map[string]interface{}{
				"disk_size":            "20",
				"source_template":      "d31e6af5-94a8-4756-abf3-6493c38db7e5",
				"use_local_ip_address": true,
			},
			Nullify: "cidr_list",
			Err:     false,
		},
		"no_network": {
			Config: map[string]interface{}{
				"disk_size":       "20",
				"source_template": "d31e6af5-94a8-4756-abf3-6493c38db7e5",
			},
			Nullify: "network",
			Err:     true,
		},
		"no_service_offering": {
			Config: map[string]interface{}{
				"disk_size":       "20",
				"source_template": "d31e6af5-94a8-4756-abf3-6493c38db7e5",
			},
			Nullify: "service_offering",
			Err:     true,
		},
		"no_template_os": {
			Config: map[string]interface{}{
				"disk_size":       "20",
				"source_template": "d31e6af5-94a8-4756-abf3-6493c38db7e5",
			},
			Nullify: "template_os",
			Err:     true,
		},
		"no_zone": {
			Config: map[string]interface{}{
				"disk_size":       "20",
				"source_template": "d31e6af5-94a8-4756-abf3-6493c38db7e5",
			},
			Nullify: "zone",
			Err:     true,
		},
		"no_source": {
			Err: true,
		},
		"both_sources": {
			Config: map[string]interface{}{
				"disk_offering":   "f043d193-242f-4941-a847-29408b998711",
				"disk_size":       "20",
				"hypervisor":      "KVM",
				"source_iso":      "fbd904dc-f46c-42e7-a467-f27480c667d5",
				"source_template": "d31e6af5-94a8-4756-abf3-6493c38db7e5",
			},
			Err: true,
		},
		"source_iso_good": {
			Config: map[string]interface{}{
				"disk_offering": "f043d193-242f-4941-a847-29408b998711",
				"hypervisor":    "KVM",
				"source_iso":    "fbd904dc-f46c-42e7-a467-f27480c667d5",
			},
			Err: false,
		},
		"source_iso_without_disk_offering": {
			Config: map[string]interface{}{
				"hypervisor": "KVM",
				"source_iso": "fbd904dc-f46c-42e7-a467-f27480c667d5",
			},
			Err: true,
		},
		"source_iso_without_hypervisor": {
			Config: map[string]interface{}{
				"disk_offering": "f043d193-242f-4941-a847-29408b998711",
				"source_iso":    "fbd904dc-f46c-42e7-a467-f27480c667d5",
			},
			Err: true,
		},
		"source_template_good": {
			Config: map[string]interface{}{
				"disk_size":       "20",
				"source_template": "d31e6af5-94a8-4756-abf3-6493c38db7e5",
			},
			Err: false,
		},
	}

	for desc, tc := range cases {
		raw := testConfig(tc.Config)

		if tc.Nullify != "" {
			raw[tc.Nullify] = nil
		}

		_, errs := NewConfig(raw)

		if tc.Err {
			if errs == nil {
				t.Fatalf("%q should error", desc)
			}
		} else {
			if errs != nil {
				t.Fatalf("%q should not error: %s", desc, errs)
			}
		}
	}
}

func testConfig(config map[string]interface{}) map[string]interface{} {
	raw := map[string]interface{}{
		"api_url":          "https://cloudstack.com/client/api",
		"api_key":          "some-api-key",
		"secret_key":       "some-secret-key",
		"cidr_list":        []interface{}{"0.0.0.0/0"},
		"network":          "c5ed8a14-3f21-4fa9-bd74-bb887fc0ed0d",
		"service_offering": "a29c52b1-a83d-4123-a57d-4548befa47a0",
		"template_os":      "52d54d24-cef1-480b-b963-527703aa4ff9",
		"zone":             "a3b594d9-25e9-47c1-9c03-7a5fc61e3f43",
	}

	for k, v := range config {
		raw[k] = v
	}

	return raw
}
