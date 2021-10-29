package ramping

import (
	"math"
)

type Linear struct {
	startRate   int
	endRate     int
	steps       int
	currentRate float64
	currentStep int
	delta       float64
}

func NewLinearRamp(startRate, endRate, steps int) *Linear {
	d := float64(endRate-startRate) / float64(steps-1)
	return &Linear{
		startRate:   startRate,
		endRate:     endRate,
		steps:       steps,
		currentStep: 1,
		currentRate: float64(startRate),
		delta:       d,
	}
}

func (l *Linear) NextRate() int {
	if l.currentStep == l.steps {
		return int(l.endRate)
	}
	if l.currentStep != 1 {
		l.currentRate = l.currentRate + l.delta
	}
	l.currentStep += 1
	return int(math.Round(l.currentRate))
}

func (l *Linear) GetSteps() int {
	return int(l.steps)
}

func (l *Linear) GetType() string {
	return "Linear ramp"
}
