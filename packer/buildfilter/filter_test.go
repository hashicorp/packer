// Copyright IBM Corp. 2013, 2025
// SPDX-License-Identifier: BUSL-1.1

package buildfilter

import (
	"testing"
)

type fakeBuild struct {
	tags   []string
	labels map[string]string
}

func (f fakeBuild) FilterTags() []string            { return f.tags }
func (f fakeBuild) FilterLabels() map[string]string { return f.labels }

func TestParse_Valid(t *testing.T) {
	cases := []struct {
		in     string
		op     Op
		key    string
		values []string
	}{
		{"tags=prod,x86", OpAll, "tags", []string{"prod", "x86"}},
		{"tags~=prod,staging", OpAny, "tags", []string{"prod", "staging"}},
		{"tags!=experimental", OpNone, "tags", []string{"experimental"}},
		{"region=us-east", OpAll, "region", []string{"us-east"}},
		{"region~=us-*", OpAny, "region", []string{"us-*"}},
		{" region = us-east ", OpAll, "region", []string{"us-east"}},
		{"tier.sub_group=a,b", OpAll, "tier.sub_group", []string{"a", "b"}},
	}
	for _, tc := range cases {
		t.Run(tc.in, func(t *testing.T) {
			exprs, err := Parse([]string{tc.in})
			if err != nil {
				t.Fatalf("Parse(%q) unexpected err: %v", tc.in, err)
			}
			if len(exprs) != 1 {
				t.Fatalf("Parse(%q) got %d exprs, want 1", tc.in, len(exprs))
			}
			e := exprs[0]
			if e.Key != tc.key {
				t.Errorf("key: got %q, want %q", e.Key, tc.key)
			}
			if e.Op != tc.op {
				t.Errorf("op: got %v, want %v", e.Op, tc.op)
			}
			if len(e.Raw) != len(tc.values) {
				t.Fatalf("values len: got %d (%v), want %d (%v)", len(e.Raw), e.Raw, len(tc.values), tc.values)
			}
			for i := range tc.values {
				if e.Raw[i] != tc.values[i] {
					t.Errorf("value[%d]: got %q, want %q", i, e.Raw[i], tc.values[i])
				}
			}
		})
	}
}

func TestParse_Invalid(t *testing.T) {
	cases := []string{
		"",
		"tags",
		"=value",
		"tags=",
		"1bad=x",
		"bad key=x",
	}
	for _, in := range cases {
		t.Run(in, func(t *testing.T) {
			if _, err := Parse([]string{in}); err == nil {
				t.Fatalf("Parse(%q) expected error, got nil", in)
			}
		})
	}
}

func TestParse_MultipleANDed(t *testing.T) {
	exprs, err := Parse([]string{"tags=prod", "region=us-east"})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(exprs) != 2 {
		t.Fatalf("got %d exprs, want 2", len(exprs))
	}
}

func TestMatch_TagsAll(t *testing.T) {
	exprs, _ := Parse([]string{"tags=prod,x86"})
	match := func(tags ...string) bool {
		return Match(fakeBuild{tags: tags}, exprs)
	}
	if !match("prod", "x86", "us-east") {
		t.Errorf("should match superset")
	}
	if match("prod") {
		t.Errorf("missing x86 should fail")
	}
	if match("dev", "x86") {
		t.Errorf("missing prod should fail")
	}
}

func TestMatch_TagsAny(t *testing.T) {
	exprs, _ := Parse([]string{"tags~=prod,staging"})
	match := func(tags ...string) bool {
		return Match(fakeBuild{tags: tags}, exprs)
	}
	if !match("prod", "x86") {
		t.Errorf("prod should match")
	}
	if !match("staging") {
		t.Errorf("staging should match")
	}
	if match("dev") {
		t.Errorf("dev should not match")
	}
}

func TestMatch_TagsNone(t *testing.T) {
	exprs, _ := Parse([]string{"tags!=experimental,broken"})
	match := func(tags ...string) bool {
		return Match(fakeBuild{tags: tags}, exprs)
	}
	if !match("prod") {
		t.Errorf("prod should match (no excluded tags)")
	}
	if match("prod", "broken") {
		t.Errorf("broken should exclude")
	}
}

func TestMatch_LabelsGlob(t *testing.T) {
	exprs, _ := Parse([]string{"region~=us-*"})
	if !Match(fakeBuild{labels: map[string]string{"region": "us-east"}}, exprs) {
		t.Errorf("us-east should match us-*")
	}
	if Match(fakeBuild{labels: map[string]string{"region": "eu-west"}}, exprs) {
		t.Errorf("eu-west should not match us-*")
	}
	if Match(fakeBuild{labels: nil}, exprs) {
		t.Errorf("missing label should not satisfy ~= selector")
	}
}

func TestMatch_LabelsNoneOnMissing(t *testing.T) {
	exprs, _ := Parse([]string{"region!=eu-*"})
	// build with no "region" label trivially satisfies != eu-*.
	if !Match(fakeBuild{labels: nil}, exprs) {
		t.Errorf("missing label should satisfy != selector")
	}
	if Match(fakeBuild{labels: map[string]string{"region": "eu-west"}}, exprs) {
		t.Errorf("eu-west should be excluded by !=eu-*")
	}
}

func TestMatch_ANDSemantics(t *testing.T) {
	exprs, _ := Parse([]string{"tags=prod", "region=us-east"})
	b1 := fakeBuild{tags: []string{"prod"}, labels: map[string]string{"region": "us-east"}}
	if !Match(b1, exprs) {
		t.Errorf("b1 should match both filters")
	}
	b2 := fakeBuild{tags: []string{"prod"}, labels: map[string]string{"region": "eu-west"}}
	if Match(b2, exprs) {
		t.Errorf("b2 should fail on region filter")
	}
}

func TestMatch_EmptyExprMatchesAll(t *testing.T) {
	if !Match(fakeBuild{}, nil) {
		t.Errorf("nil exprs should match any build")
	}
}
