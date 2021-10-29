package ramping

type RampType int64

const (
	LinearRamp RampType = iota
	ExponentialRamp
)

type Ramper interface {
	NextRate() int
	GetSteps() int
	GetType() string
}

// NewRampingService when using None ramping
// send the rate in the minRate it will be the used
// to initialize NoneRamp
func NewRampingService(rampType RampType, startRate, endRate, steps int) Ramper {
	switch rampType {
	case LinearRamp:
		return NewLinearRamp(startRate, endRate, steps)
	case ExponentialRamp:
		return NewExponentialRamp(startRate, endRate, steps)
	}
	return nil
}
