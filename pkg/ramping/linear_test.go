package ramping

import (
	"testing"
)

func TestLinear_NextRate_scenario1(t *testing.T) {
	l := NewLinearRamp(2, 50, 6)
	tests := []struct {
		name string
		want int
	}{
		{"step1", 2},
		{"step2", 12},
		{"step3", 21},
		{"step4", 31},
		{"step5", 40},
		{"step6", 50},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := l.NextRate(); got != tt.want {
				t.Errorf("Linear.NextRate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLinear_NextRate_scenario2(t *testing.T) {
	l := NewLinearRamp(1, 15, 9)
	tests := []struct {
		name string
		want int
	}{
		{"step1", 1},
		{"step2", 3},
		{"step3", 5},
		{"step4", 6},
		{"step5", 8},
		{"step6", 10},
		{"step7", 12},
		{"step8", 13},
		{"step9", 15},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := l.NextRate(); got != tt.want {
				t.Errorf("Linear.NextRate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLinear_NextRate_scenario3(t *testing.T) {
	l := NewLinearRamp(1, 10, 10)
	tests := []struct {
		name string
		want int
	}{
		{"step1", 1},
		{"step2", 2},
		{"step3", 3},
		{"step4", 4},
		{"step5", 5},
		{"step6", 6},
		{"step7", 7},
		{"step8", 8},
		{"step9", 9},
		{"step10", 10},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := l.NextRate(); got != tt.want {
				t.Errorf("Linear.NextRate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLinear_GetSteps(t *testing.T) {
	t.Run("testing GetSteps", func(t *testing.T) {
		l := NewLinearRamp(1, 10, 7)
		if got := l.GetSteps(); got != 7 {
			t.Errorf("Linear.GetSteps() = %v, want %v", got, 7)
		}
	})
}

func TestLinear_GetType(t *testing.T) {
	t.Run("testing GetType", func(t *testing.T) {
		l := NewLinearRamp(1, 10, 7)
		if got := l.GetType(); got != "Linear ramp" {
			t.Errorf("Linear.GetType() = %v, want %v", got, "Linear ramp")
		}
	})
}
