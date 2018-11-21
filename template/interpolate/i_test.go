package interpolate

import (
	"reflect"
	"testing"
)

func TestIRender(t *testing.T) {
	cases := map[string]struct {
		Ctx    *Context
		Value  string
		Result string
	}{
		"basic": {
			nil,
			"foo",
			"foo",
		},
	}

	for k, tc := range cases {
		i := &I{Value: tc.Value}
		result, err := i.Render(tc.Ctx)
		if err != nil {
			t.Fatalf("%s\n\ninput: %s\n\nerr: %s", k, tc.Value, err)
		}
		if result != tc.Result {
			t.Fatalf(
				"%s\n\ninput: %s\n\nexpected: %s\n\ngot: %s",
				k, tc.Value, tc.Result, result)
		}
	}
}

func TestContext_ParseArgs(t *testing.T) {
	type fields Context
	type args struct {
		commands []string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantRes [][]string
		wantErr bool
	}{
		{"empty", fields{}, args{}, [][]string{}, false},
		{"commands", fields{}, args{[]string{"a b c", " d ' e' f "}}, [][]string{
			{"a", "b", "c"},
			{"d", " e", "f"},
		}, false},
		{"quoted interpolation", fields{Data: "x"}, args{[]string{`"Professor {{.}}"`}}, [][]string{
			{"Professor x"},
		}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Context{
				Data:               tt.fields.Data,
				Funcs:              tt.fields.Funcs,
				UserVariables:      tt.fields.UserVariables,
				SensitiveVariables: tt.fields.SensitiveVariables,
				EnableEnv:          tt.fields.EnableEnv,
				BuildName:          tt.fields.BuildName,
				BuildType:          tt.fields.BuildType,
				TemplatePath:       tt.fields.TemplatePath,
			}
			gotRes, err := c.ParseArgs(tt.args.commands)
			if (err != nil) != tt.wantErr {
				t.Errorf("Context.ParseArgs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotRes, tt.wantRes) {
				t.Errorf("Context.ParseArgs() = %v, want %v", gotRes, tt.wantRes)
			}
		})
	}
}
