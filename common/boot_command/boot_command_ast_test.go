package bootcommand

import (
	"log"
	"strings"
	"testing"
)

func toIfaceSlice(v interface{}) []interface{} {
	if v == nil {
		return nil
	}
	return v.([]interface{})
}

func TestParse(t *testing.T) {
	in := "<wait><wait20><wait3s><wait4m2ns>"
	in += "foo/bar > one 界"
	in += "<fOn> b<fOff>"
	in += "<foo><f3><f12><spacebar><leftalt><rightshift><rightsuper>"
	got, err := ParseReader("", strings.NewReader(in))
	if err != nil {
		log.Fatal(err)
	}
	gL := toIfaceSlice(got)
	for _, g := range gL {
		log.Printf("%s\n", g)
	}

}
