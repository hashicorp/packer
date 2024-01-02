// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package sleep

import (
	"context"
	"testing"
	"time"
)

func test1sConfig() map[string]interface{} {
	return map[string]interface{}{
		"duration": "1s",
	}
}

func TestConfigPrepare_1s(t *testing.T) {
	raw := test1sConfig()
	var p Provisioner
	err := p.Prepare(raw)
	if err != nil {
		t.Fatalf("prerare failed: %v", err)
	}

	if p.Duration != time.Second {
		t.Fatal("wrong duration")
	}
}

func TestProvisioner_Provision(t *testing.T) {
	ctxCancelled, cancel := context.WithCancel(context.Background())
	cancel()
	type fields struct {
		Duration time.Duration
	}
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{"valid sleep", fields{1 * time.Millisecond}, args{context.Background()}, false},
		{"timeout", fields{1 * time.Millisecond}, args{ctxCancelled}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Provisioner{
				Duration: tt.fields.Duration,
			}
			if err := p.Provision(tt.args.ctx, nil, nil, make(map[string]interface{})); (err != nil) != tt.wantErr {
				t.Errorf("Provisioner.Provision() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
