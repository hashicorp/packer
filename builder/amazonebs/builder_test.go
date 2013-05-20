package amazonebs

import (
	"cgl.tideland.biz/asserts"
	"github.com/mitchellh/packer/packer"
	"testing"
)

func TestBuilder_ImplementsBuilder(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

	var actual packer.Builder
	assert.Implementor(&Builder{}, &actual, "should be a Builder")
}

func TestBuilder_Prepare_BadType(t *testing.T) {
	b := &Builder{}
	c := map[string]interface{}{
		"access_key": []string{},
	}

	err := b.Prepare(c)
	if err == nil {
		t.Fatalf("prepare should fail")
	}
}

func TestBuilder_Prepare_Good(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

	b := &Builder{}
	c := map[string]interface{}{
		"access_key": "foo",
		"secret_key": "bar",
		"source_ami": "123456",
	}

	err := b.Prepare(c)
	assert.Nil(err, "should not have an error")
	assert.Equal(b.config.AccessKey, "foo", "should be valid access key")
	assert.Equal(b.config.SecretKey, "bar", "should be valid secret key")
	assert.Equal(b.config.SourceAmi, "123456", "should have source AMI")
}
