package api_test

import (
	"testing"

	"github.com/tk3fftk/tfustomize/api"
)

func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  string
		wantErr bool
	}{
		{
			name:    "valid config",
			config:  "../test/overlay/tfustomization.hcl",
			wantErr: false,
		},
		{
			name:    "invalid config",
			config:  "../test/invalid/broken_schema_tfustomization.hcl",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := api.LoadConfig(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
