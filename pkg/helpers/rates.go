package helpers

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	vegeta "github.com/tsenart/vegeta/v12/lib"
)

// ParseRate addapted `rateFlag` from tsenart/vegeta
// since the `rateFlag` is not exported it is not possible
// to reuse the same method, but to match and correctly
// generate the rate we decided to use the same function.
// https://github.com/tsenart/vegeta/blob/d73edf2bc2663d83848da2a97a8401a7ed1440bc/flags.go#L68
func ParseRate(rate string) (vegeta.Rate, error) {
	if rate == "infinity" {
		return vegeta.Rate{}, nil
	}

	ps := strings.SplitN(rate, "/", 2)
	switch len(ps) {
	case 1:
		ps = append(ps, "1s")
	case 0:
		return vegeta.Rate{}, fmt.Errorf("-rate format %q doesn't match the \"freq/duration\" format (i.e. 50/1s)", rate)
	}

	f, err := strconv.Atoi(ps[0])
	if err != nil {
		return vegeta.Rate{}, err
	}

	if f == 0 {
		return vegeta.Rate{}, nil
	}

	switch ps[1] {
	case "ns", "us", "µs", "ms", "s", "m", "h":
		ps[1] = "1" + ps[1]
	}

	p, err := time.ParseDuration(ps[1])
	if err != nil {
		return vegeta.Rate{}, err
	}
	return vegeta.Rate{Freq: f, Per: p}, nil
}
