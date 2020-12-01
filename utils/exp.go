package utils

import "math/rand"

type ExponentialDistribution struct {
	rng    *rand.Rand
	lambda float64
}

func NewExponentialDistribution(rng *rand.Rand, lambda float64) *ExponentialDistribution {
	return &ExponentialDistribution{rng: rng, lambda: lambda}
}

func (d *ExponentialDistribution) ExpFloat64() float64 {
	return d.rng.ExpFloat64() / d.lambda
}
