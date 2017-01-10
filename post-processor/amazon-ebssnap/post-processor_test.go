package ebssnap

import (
	"bytes"
	"github.com/mitchellh/packer/packer"
	"testing"
)

func testConfig() map[string]interface{} {
	return map[string]interface{}{}
}

func testPP(t *testing.T) *PostProcessor {
	var p PostProcessor
	if err := p.Configure(testConfig()); err != nil {
		t.Fatalf("err: %s", err)
	}

	return &p
}

func testUi() *packer.BasicUi {
	return &packer.BasicUi{
		Reader: new(bytes.Buffer),
		Writer: new(bytes.Buffer),
	}
}

func TestPostProcessor_ImplementsPostProcessor(t *testing.T) {
	var _ packer.PostProcessor = new(PostProcessor)
}

// Test Helpers

// See checksum and compress
//func setup(t *testing.T) (packer.Ui, packer.Artifact, error) {
//}
//packer/post_processor_mock.go:// MockPostProcessor is an implementation of PostProcessor that can be
//packer/post_processor_mock.go:type MockPostProcessor struct {
//packer/post_processor_mock.go:func (t *MockPostProcessor) Configure(configs ...interface{}) error {
//packer/post_processor_mock.go:func (t *MockPostProcessor) PostProcess(ui Ui, a Artifact) (Artifact, bool, error) {
//packer/post_processor_mock.go:  return &MockArtifact{
