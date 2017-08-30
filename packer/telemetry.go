package packer

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"time"

	checkpoint "github.com/hashicorp/go-checkpoint"
	packerVersion "github.com/hashicorp/packer/version"
)

const TelemetryVersion string = "beta/packer/4"
const TelemetryPanicVersion string = "beta/packer_panic/4"

var CheckpointReporter CheckpointTelemetry

func init() {
	CheckpointReporter.startTime = time.Now().UTC()
}

type PackerReport struct {
	Spans    []*TelemetrySpan `json:"spans"`
	ExitCode int              `json:"exit_code"`
	Error    string           `json:"error"`
	Command  string           `json:"command"`
}

type CheckpointTelemetry struct {
	enabled       bool
	spans         []*TelemetrySpan
	signatureFile string
	startTime     time.Time
}

func (c *CheckpointTelemetry) Enable(disableSignature bool) {
	configDir, err := ConfigDir()
	if err != nil {
		log.Printf("[WARN] (telemetry) setup error: %s", err)
		return
	}

	signatureFile := ""
	if disableSignature {
		log.Printf("[INFO] (telemetry) Checkpoint signature disabled")
	} else {
		signatureFile = filepath.Join(configDir, "checkpoint_signature")
	}

	c.signatureFile = signatureFile
	c.enabled = true
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
	if !c.enabled {
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

func (c *CheckpointTelemetry) AddSpan(name, pluginType string) *TelemetrySpan {
	log.Printf("[INFO] (telemetry) Starting %s %s", pluginType, name)
	ts := &TelemetrySpan{
		Name:      name,
		Type:      pluginType,
		StartTime: time.Now().UTC(),
	}
	c.spans = append(c.spans, ts)
	return ts
}

func (c *CheckpointTelemetry) Finalize(command string, errCode int, err error) error {
	if !c.enabled {
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

	ctx, cancel := context.WithTimeout(context.Background(), 750*time.Millisecond)
	defer cancel()

	return checkpoint.Report(ctx, params)
}

type TelemetrySpan struct {
	Name      string    `json:"name"`
	Type      string    `json:"type"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Error     string    `json:"error"`
}

func (s *TelemetrySpan) End(err error) {
	s.EndTime = time.Now().UTC()
	log.Printf("[INFO] (telemetry) ending %s", s.Name)
	if err != nil {
		s.Error = err.Error()
		log.Printf("[INFO] (telemetry) found error: %s", err.Error())
	}
}
