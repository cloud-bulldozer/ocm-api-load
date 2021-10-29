package ramping

import (
	"testing"
)

func TestExponential_NextRate_scenario1(t *testing.T) {
	e := NewExponentialRamp(2, 20, 8)
	tests := []struct {
		name string
		want int
	}{
		{"step1", 3},
		{"step2", 4},
		{"step3", 5},
		{"step4", 6},
		{"step5", 8},
		{"step6", 11},
		{"step7", 15},
		{"step8", 20},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := e.NextRate(); got != tt.want {
				t.Errorf("Exponential.NextRate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExponential_NextRate_scenario2(t *testing.T) {
	e := NewExponentialRamp(1, 100, 8)
	tests := []struct {
		name string
		want int
	}{
		{"step1", 2},
		{"step2", 3},
		{"step3", 6},
		{"step4", 10},
		{"step5", 18},
		{"step6", 32},
		{"step7", 56},
		{"step8", 100},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := e.NextRate(); got != tt.want {
				t.Errorf("Exponential.NextRate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExponential_GetSteps(t *testing.T) {
	t.Run("testing GetSteps", func(t *testing.T) {
		e := NewExponentialRamp(1, 10, 5)
		if got := e.GetSteps(); got != 5 {
			t.Errorf("Exponential.GetSteps() = %v, want %v", got, 5)
		}
	})
}

func TestExponential_GetType(t *testing.T) {
	t.Run("testing GetType", func(t *testing.T) {
		e := NewExponentialRamp(1, 10, 5)
		if got := e.GetType(); got != "Exponential ramp" {
			t.Errorf("Exponential.GetType() = %v, want %v", got, "Exponential ramp")
		}
	})
}
