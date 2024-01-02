// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package packer

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"time"

	checkpoint "github.com/hashicorp/go-checkpoint"
	"github.com/hashicorp/packer-plugin-sdk/pathing"
	packerVersion "github.com/hashicorp/packer/version"
	"github.com/zclconf/go-cty/cty"
)

type PackerTemplateType string

const (
	UnknownTemplate PackerTemplateType = "Unknown"
	HCL2Template    PackerTemplateType = "HCL2"
	JSONTemplate    PackerTemplateType = "JSON"
)

const TelemetryVersion string = "beta/packer/7"
const TelemetryPanicVersion string = "beta/packer_panic/4"

var CheckpointReporter *CheckpointTelemetry

type PackerReport struct {
	Spans        []*TelemetrySpan   `json:"spans"`
	ExitCode     int                `json:"exit_code"`
	Error        string             `json:"error"`
	Command      string             `json:"command"`
	TemplateType PackerTemplateType `json:"template_type"`
	UseBundled   bool               `json:"use_bundled"`
}

type CheckpointTelemetry struct {
	spans         []*TelemetrySpan
	signatureFile string
	startTime     time.Time
	templateType  PackerTemplateType
	useBundled    bool
}

func NewCheckpointReporter(disableSignature bool) *CheckpointTelemetry {
	if disabled := os.Getenv("CHECKPOINT_DISABLE"); disabled != "" {
		return nil
	}

	configDir, err := pathing.ConfigDir()
	if err != nil {
		log.Printf("[WARN] (telemetry) setup error: %s", err)
		return nil
	}

	signatureFile := ""
	if disableSignature {
		log.Printf("[INFO] (telemetry) Checkpoint signature disabled")
	} else {
		signatureFile = filepath.Join(configDir, "checkpoint_signature")
	}

	return &CheckpointTelemetry{
		signatureFile: signatureFile,
		startTime:     time.Now().UTC(),
		templateType:  UnknownTemplate,
	}
}

func (c *CheckpointTelemetry) baseParams(prefix string) *checkpoint.ReportParams {
	version := packerVersion.Version
	if packerVersion.VersionPrerelease != "" {
		version += "-" + packerVersion.VersionPrerelease
	}

	return &checkpoint.ReportParams{
		Product:       "packer",
		SchemaVersion: prefix,
		StartTime:     c.startTime,
		Version:       version,
		RunID:         os.Getenv("PACKER_RUN_UUID"),
		SignatureFile: c.signatureFile,
	}
}

func (c *CheckpointTelemetry) ReportPanic(m string) error {
	if c == nil {
		return nil
	}
	panicParams := c.baseParams(TelemetryPanicVersion)
	panicParams.Payload = m
	panicParams.EndTime = time.Now().UTC()

	// This timeout can be longer because it runs in the real main.
	// We're also okay waiting a bit longer to collect panic information
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	return checkpoint.Report(ctx, panicParams)
}

func (c *CheckpointTelemetry) AddSpan(name, pluginType string, options interface{}) *TelemetrySpan {
	if c == nil {
		return nil
	}
	log.Printf("[INFO] (telemetry) Starting %s %s", pluginType, name)

	ts := &TelemetrySpan{
		Name:      name,
		Options:   flattenConfigKeys(options),
		StartTime: time.Now().UTC(),
		Type:      pluginType,
	}
	c.spans = append(c.spans, ts)
	return ts
}

// SetTemplateType registers the template type being processed for a Packer command
func (c *CheckpointTelemetry) SetTemplateType(t PackerTemplateType) {
	if c == nil {
		return
	}

	c.templateType = t
}

// SetBundledUsage marks the template as using bundled plugins
func (c *CheckpointTelemetry) SetBundledUsage() {
	if c == nil {
		return
	}
	c.useBundled = true
}

func (c *CheckpointTelemetry) Finalize(command string, errCode int, err error) error {
	if c == nil {
		return nil
	}

	params := c.baseParams(TelemetryVersion)
	params.EndTime = time.Now().UTC()

	extra := &PackerReport{
		Spans:    c.spans,
		ExitCode: errCode,
		Command:  command,
	}
	if err != nil {
		extra.Error = err.Error()
	}

	extra.UseBundled = c.useBundled
	extra.TemplateType = c.templateType
	params.Payload = extra
	// b, _ := json.MarshalIndent(params, "", "    ")
	// log.Println(string(b))

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	log.Printf("[INFO] (telemetry) Finalizing.")
	return checkpoint.Report(ctx, params)
}

type TelemetrySpan struct {
	EndTime   time.Time `json:"end_time"`
	Error     string    `json:"error"`
	Name      string    `json:"name"`
	Options   []string  `json:"options"`
	StartTime time.Time `json:"start_time"`
	Type      string    `json:"type"`
}

func (s *TelemetrySpan) End(err error) {
	if s == nil {
		return
	}
	s.EndTime = time.Now().UTC()
	log.Printf("[INFO] (telemetry) ending %s", s.Name)
	if err != nil {
		s.Error = err.Error()
	}
}

func flattenConfigKeys(options interface{}) []string {
	var flatten func(string, interface{}) []string

	flatten = func(prefix string, options interface{}) (strOpts []string) {
		switch opt := options.(type) {
		case map[string]interface{}:
			return flattenJSON(prefix, options)
		case cty.Value:
			return flattenHCL(prefix, opt)
		default:
			return nil
		}
	}

	flattened := flatten("", options)
	sort.Strings(flattened)
	return flattened
}

func flattenJSON(prefix string, options interface{}) (strOpts []string) {
	if m, ok := options.(map[string]interface{}); ok {
		for k, v := range m {
			if prefix != "" {
				k = prefix + "/" + k
			}
			if n, ok := v.(map[string]interface{}); ok {
				strOpts = append(strOpts, flattenJSON(k, n)...)
			} else {
				strOpts = append(strOpts, k)
			}
		}
	}
	return
}

func flattenHCL(prefix string, v cty.Value) (args []string) {
	if v.IsNull() {
		return []string{}
	}
	t := v.Type()
	switch {
	case t.IsObjectType(), t.IsMapType():
		if !v.IsKnown() {
			return []string{}
		}
		it := v.ElementIterator()
		for it.Next() {
			key, val := it.Element()
			keyStr := key.AsString()

			if val.IsNull() {
				continue
			}

			if prefix != "" {
				keyStr = fmt.Sprintf("%s/%s", prefix, keyStr)
			}

			if val.Type().IsObjectType() || val.Type().IsMapType() {
				args = append(args, flattenHCL(keyStr, val)...)
			} else {
				args = append(args, keyStr)
			}
		}
	}
	return args
}
