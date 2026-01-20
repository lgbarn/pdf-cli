package ocr

import "testing"

func TestParseBackendType(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  BackendType
	}{
		{"native", "native", BackendNative},
		{"wasm", "wasm", BackendWASM},
		{"auto", "auto", BackendAuto},
		{"empty string defaults to auto", "", BackendAuto},
		{"invalid defaults to auto", "invalid", BackendAuto},
		{"uppercase is not recognized", "NATIVE", BackendAuto},
		{"mixed case is not recognized", "Native", BackendAuto},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ParseBackendType(tt.input); got != tt.want {
				t.Errorf("ParseBackendType(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestBackendTypeString(t *testing.T) {
	tests := []struct {
		name  string
		input BackendType
		want  string
	}{
		{"native", BackendNative, "native"},
		{"wasm", BackendWASM, "wasm"},
		{"auto", BackendAuto, "auto"},
		{"invalid value defaults to auto", BackendType(99), "auto"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.input.String(); got != tt.want {
				t.Errorf("BackendType.String() = %q, want %q", got, tt.want)
			}
		})
	}
}
