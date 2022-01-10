package ramping

import "math"

type Exponential struct {
	startRate   int
	endRate     int
	steps       int
	currentRate float64
	currentStep int
	delta       float64
}

func NewExponentialRamp(startRate, endRate, steps int) *Exponential {
	d := math.Pow((float64(endRate) / float64(startRate)), (1 / float64(steps)))
	return &Exponential{
		startRate:   startRate,
		endRate:     endRate,
		steps:       steps,
		currentStep: 1,
		currentRate: float64(startRate),
		delta:       d,
	}
}

func (e *Exponential) NextRate() int {
	if e.currentStep == e.steps {
		return int(e.endRate)
	}
	e.currentRate = float64(e.startRate) * math.Pow(e.delta, float64(e.currentStep))
	e.currentStep += 1
	return int(math.Round(e.currentRate))
}

func (e *Exponential) GetSteps() int {
	return int(e.steps)
}

func (e *Exponential) GetType() string {
	return "Exponential ramp"
}
