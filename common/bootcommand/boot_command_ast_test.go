package bootcommand

import (
	"fmt"
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_parse(t *testing.T) {
	in := "<wait><wait20><wait3s><wait4m2ns>"
	in += "foo/bar > one 界"
	in += "<fon> b<fOff>"
	in += "<foo><f3><f12><spacebar><leftalt><rightshift><rightsuper>"
	expected := []string{
		"Wait<1s>",
		"Wait<20s>",
		"Wait<3s>",
		"Wait<4m0.000000002s>",
		"LIT-Press(f)",
		"LIT-Press(o)",
		"LIT-Press(o)",
		"LIT-Press(/)",
		"LIT-Press(b)",
		"LIT-Press(a)",
		"LIT-Press(r)",
		"LIT-Press( )",
		"LIT-Press(>)",
		"LIT-Press( )",
		"LIT-Press(o)",
		"LIT-Press(n)",
		"LIT-Press(e)",
		"LIT-Press( )",
		"LIT-Press(界)",
		"LIT-On(f)",
		"LIT-Press( )",
		"LIT-Press(b)",
		"LIT-Off(f)",
		"LIT-Press(<)",
		"LIT-Press(f)",
		"LIT-Press(o)",
		"LIT-Press(o)",
		"LIT-Press(>)",
		"Spec-Press(f3)",
		"Spec-Press(f12)",
		"Spec-Press(spacebar)",
		"Spec-Press(leftalt)",
		"Spec-Press(rightshift)",
		"Spec-Press(rightsuper)",
	}

	seq, err := GenerateExpressionSequence(in)
	if err != nil {
		log.Fatal(err)
	}
	for i, exp := range seq {
		assert.Equal(t, expected[i], fmt.Sprintf("%s", exp))
		log.Printf("%s\n", exp)
	}
}

func Test_special(t *testing.T) {
	var specials = []struct {
		in  string
		out string
	}{
		{
			"<rightShift><rightshift><RIGHTSHIFT>",
			"Spec-Press(rightshift)",
		},
		{
			"<delon><delON><deLoN><DELON>",
			"Spec-On(del)",
		},
		{
			"<enteroff><enterOFF><eNtErOfF><ENTEROFF>",
			"Spec-Off(enter)",
		},
	}
	for _, tt := range specials {
		seq, err := GenerateExpressionSequence(tt.in)
		if err != nil {
			log.Fatal(err)
		}
		for _, exp := range seq {
			assert.Equal(t, tt.out, exp.(*specialExpression).String())
		}
	}
}

func Test_validation(t *testing.T) {
	var expressions = []struct {
		in    string
		valid bool
	}{
		{
			"<wait1m>",
			true,
		},
		{
			"<wait-1m>",
			false,
		},
		{
			"<f1>",
			true,
		},
		{
			"<",
			true,
		},
	}
	for _, tt := range expressions {
		exp, err := GenerateExpressionSequence(tt.in)
		if err != nil {
			log.Fatal(err)
		}

		assert.Len(t, exp, 1)
		err = exp[0].Validate()
		if tt.valid {
			assert.NoError(t, err)
		} else {
			assert.Error(t, err)
		}
	}
}

func Test_empty(t *testing.T) {
	exp, err := GenerateExpressionSequence("")
	assert.NoError(t, err, "should have parsed an empty input okay.")
	assert.Len(t, exp, 0)
}
