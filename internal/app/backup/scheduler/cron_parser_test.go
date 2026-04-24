package scheduler

import (
	"testing"
)

func TestConvertCronToSystemd(t *testing.T) {
	tests := []struct {
		name     string
		cronExpr string
		expected string
		wantErr  bool
	}{
		{
			name:     "Daily at 2 AM",
			cronExpr: "0 2 * * *",
			expected: "*-*-* 02:00:00",
			wantErr:  false,
		},
		{
			name:     "Every 3 days at 1 AM",
			cronExpr: "0 1 */3 * *",
			expected: "*-*-1/3 01:00:00",
			wantErr:  false,
		},
		{
			name:     "Every Monday at 3:30",
			cronExpr: "30 3 * * Mon",
			expected: "Mon *-*-* 03:30:00",
			wantErr:  false,
		},
		{
			name:     "Invalid cron",
			cronExpr: "0 2 * *",
			expected: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ConvertCronToSystemd(tt.cronExpr)
			if (err != nil) != tt.wantErr {
				t.Errorf("ConvertCronToSystemd() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.expected {
				t.Errorf("ConvertCronToSystemd() = %v, want %v", got, tt.expected)
			}
		})
	}
}
