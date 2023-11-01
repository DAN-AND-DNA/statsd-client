package statsdclient

import (
	"github.com/dan-and-dna/statsd-client/internal"
)

func Init(prefix string, addresses []string) error {
	return internal.GetSingleInst().Init(prefix, addresses)
}

func Stop() {
	internal.GetSingleInst().Stop()
}

func ReCreate(prefix string, addresses []string) error {
	return internal.GetSingleInst().ReCreate(prefix, addresses)
}

func Increment(name string, count int, rate float64, needFlush bool) error {
	return internal.GetSingleInst().Increment(name, count, rate, needFlush)
}

func Gauge(name string, count int, needFlush bool) error {
	return internal.GetSingleInst().Gauge(name, count, needFlush)
}

func Histogram(name string, count int, needFlush bool) error {
	return internal.GetSingleInst().Histogram(name, count, needFlush)
}
