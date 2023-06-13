package elastic

import (
	vegeta "github.com/tsenart/vegeta/v12/lib"
)

type doc struct {
	Uuid    string         `json:"uuid"`
	Version string         `json:"version"`
	Attack  string         `json:"attack"`
	Metrics vegeta.Metrics `json:"metrics"`
}
