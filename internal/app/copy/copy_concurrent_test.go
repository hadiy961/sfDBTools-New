package copy

import (
	"testing"
)

func TestSanitizeArgsForData(t *testing.T) {
	s := &Service{}

	tests := []struct {
		name string
		args string
		want string
	}{
		{
			name: "remove routines and triggers",
			args: "--routines --triggers --single-transaction --quick",
			want: "--single-transaction --quick",
		},
		{
			name: "remove short form flags",
			args: "-r -t -e --quick",
			want: "--quick",
		},
		{
			name: "remove gtid purged",
			args: "--set-gtid-purged=OFF --routines --quick",
			want: "--quick",
		},
		{
			name: "mixed case and spaces",
			args: "--ROUTINES   --Triggers --events  --databases db1",
			want: "",
		},
		{
			name: "keep unknown flags",
			args: "--hex-blob --opt --routines",
			want: "--hex-blob --opt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := s.sanitizeArgsForData(tt.args)
			if got != tt.want {
				t.Errorf("sanitizeArgsForData() = %q, want %q", got, tt.want)
			}
		})
	}
}
