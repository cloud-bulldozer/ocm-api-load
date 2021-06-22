package providers

import (
	"math/rand"
	"time"
)

type RandomProvider struct {
	Items      []interface{}
	Randomizer *rand.Rand
}

func (r *RandomProvider) GetItem() interface{} {
	return r.Items[r.Randomizer.Intn(len(r.Items))]
}

func NewRandomProvider(items ...interface{}) *RandomProvider {
	return &RandomProvider{
		Items:      items,
		Randomizer: rand.New(rand.NewSource(time.Now().UTC().UnixNano())),
	}
}
