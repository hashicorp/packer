package yandeximport

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_s3URLToBucketKey(t *testing.T) {
	tests := []struct {
		name       string
		storageURL string
		wantBucket string
		wantKey    string
		wantErr    bool
	}{
		{
			name:       "path-style url #1",
			storageURL: "https://storage.yandexcloud.net/bucket1/key1/foobar.txt",
			wantBucket: "bucket1",
			wantKey:    "key1/foobar.txt",
			wantErr:    false,
		},
		{
			name:       "path-style url #2",
			storageURL: "https://storage.yandexcloud.net/bucket1.with.dots/key1/foobar.txt",
			wantBucket: "bucket1.with.dots",
			wantKey:    "key1/foobar.txt",
			wantErr:    false,
		},
		{
			name:       "host-style url #1",
			storageURL: "https://bucket1.with.dots.storage.yandexcloud.net/key1/foobar.txt",
			wantBucket: "bucket1.with.dots",
			wantKey:    "key1/foobar.txt",
			wantErr:    false,
		},
		{
			name:       "host-style url #2",
			storageURL: "https://bucket-with-dash.storage.yandexcloud.net/key2/foobar.txt",
			wantBucket: "bucket-with-dash",
			wantKey:    "key2/foobar.txt",
			wantErr:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotBucket, gotKey, err := s3URLToBucketKey(tt.storageURL)
			if (err != nil) != tt.wantErr {
				t.Errorf("s3URLToBucketKey() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.wantBucket, gotBucket)
			assert.Equal(t, tt.wantKey, gotKey)
		})
	}
}
