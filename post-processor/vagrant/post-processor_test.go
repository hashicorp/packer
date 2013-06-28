package vagrant

import (
	"github.com/mitchellh/packer/packer"
	"testing"
)

func testConfig() map[string]interface{} {
	return map[string]interface{}{}
}

func TestPostProcessor_ImplementsPostProcessor(t *testing.T) {
	var raw interface{}
	raw = &PostProcessor{}
	if _, ok := raw.(packer.PostProcessor); !ok {
		t.Fatalf("AWS PostProcessor should be a PostProcessor")
	}
}

func TestBuilderPrepare_OutputPath(t *testing.T) {
	var p PostProcessor

	// Default
	c := testConfig()
	delete(c, "output")
	err := p.Configure(c)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	// Bad template
	c["output"] = "bad {{{{.Template}}}}"
	err = p.Configure(c)
	if err == nil {
		t.Fatal("should have error")
	}
}

func TestBuilderPrepare_PPConfig(t *testing.T) {
	var p PostProcessor

	// Default
	c := testConfig()
	c["aws"] = map[string]interface{}{}
	err := p.Configure(c)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
}
