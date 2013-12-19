package vagrant

import (
	"github.com/mitchellh/packer/packer"
	"testing"
)

func testConfig() map[string]interface{} {
	return map[string]interface{}{}
}

func TestPostProcessor_ImplementsPostProcessor(t *testing.T) {
	var _ packer.PostProcessor = new(PostProcessor)
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

func TestProviderForName(t *testing.T) {
	if v, ok := providerForName("virtualbox").(*VBoxProvider); !ok {
		t.Fatalf("bad: %#v", v)
	}

	if providerForName("nope") != nil {
		t.Fatal("should be nil if bad provider")
	}
}
