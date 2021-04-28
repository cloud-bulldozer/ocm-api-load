package helpers

import (
	"strconv"
	"strings"
	"time"

	vegeta "github.com/tsenart/vegeta/v12/lib"
)

func ParseRate(rate string) (vegeta.Rate, error) {
	r := strings.Split(rate, "/")
	f, err := strconv.ParseInt(r[0], 10, 0)
	if err != nil {
		return vegeta.Rate{}, err
	}
	p, err := time.ParseDuration("1" + r[1])
	if err != nil {
		return vegeta.Rate{}, err
	}
	return vegeta.Rate{
		Freq: int(f),
		Per:  p,
	}, nil
}
