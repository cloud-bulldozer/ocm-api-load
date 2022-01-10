package ramping

import (
	"reflect"
	"testing"
)

func TestNewRampingService(t *testing.T) {
	tests := []struct {
		name      string
		rampType  RampType
		startRate int
		endRate   int
		steps     int
		want      Ramper
	}{
		{"Linear", LinearRamp, 2, 10, 4, NewLinearRamp(2, 10, 4)},
		{"Exponential", ExponentialRamp, 2, 10, 4, NewExponentialRamp(2, 10, 4)},
		{"Error", 2, 2, 10, 4, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewRampingService(tt.rampType, tt.startRate, tt.endRate, tt.steps); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewRampingService() = %v, want %v", got, tt.want)
			}
		})
	}
}
