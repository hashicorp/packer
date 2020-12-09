package packer

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"sort"
	"time"

	checkpoint "github.com/hashicorp/go-checkpoint"
	"github.com/hashicorp/packer-plugin-sdk/pathing"
	packerVersion "github.com/hashicorp/packer/version"
)

const TelemetryVersion string = "beta/packer/5"
const TelemetryPanicVersion string = "beta/packer_panic/4"

var CheckpointReporter *CheckpointTelemetry

type PackerReport struct {
	Spans    []*TelemetrySpan `json:"spans"`
	ExitCode int              `json:"exit_code"`
	Error    string           `json:"error"`
	Command  string           `json:"command"`
}

type CheckpointTelemetry struct {
	spans         []*TelemetrySpan
	signatureFile string
	startTime     time.Time
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
		if m, ok := options.(map[string]interface{}); ok {
			for k, v := range m {
				if prefix != "" {
					k = prefix + "/" + k
				}
				if n, ok := v.(map[string]interface{}); ok {
					strOpts = append(strOpts, flatten(k, n)...)
				} else {
					strOpts = append(strOpts, k)
				}
			}
		}
		return
	}

	flattened := flatten("", options)
	sort.Strings(flattened)
	return flattened
}
