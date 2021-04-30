package helpers

import (
	"reflect"
	"testing"
	"time"

	vegeta "github.com/tsenart/vegeta/v12/lib"
)

func TestParseRate(t *testing.T) {
	tests := []struct {
		name    string
		rate    string
		wantF   int
		wantD   string
		wantErr bool
	}{
		{"1", "1/s", 1, "1s", false},
		{"2", "infinity", 0, "", false},
		{"3", "10", 10, "1s", false},
		{"4", "0", 0, "", false},
		{"5", "1/m", 1, "1m", false},
		{"6", "1/h", 1, "1h", false},
		{"7", "1/ms", 1, "1ms", false},
		{"8", "1/ns", 1, "1ns", false},
		{"9", "1/us", 1, "1us", false},
		{"10", "1/µs", 1, "1µs", false},
		{"11", "500/s", 500, "1s", false},
		{"12", "1/t", 0, "", true},
		{"13", "fast", 0, "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseRate(tt.rate)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseRate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			d, err := time.ParseDuration(tt.wantD)
			if (err != nil) != tt.wantErr && tt.wantD != "" {
				t.Errorf("ParseRate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			want := vegeta.ConstantPacer{Freq: tt.wantF, Per: d}
			if !reflect.DeepEqual(got, want) {
				t.Errorf("ParseRate() = %v, want %v", got, want)
			}
		})
	}
}
