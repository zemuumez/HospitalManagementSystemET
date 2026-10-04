package domain

import (
	"math"
	"testing"
	"time"
)

func TestVitalsValidation(t *testing.T) {
	now := time.Now()
	oxygen := 99.0
	valid := VitalsInput{ObservedAt: now.Add(-time.Minute), Measurements: Measurements{OxygenPercent: &oxygen}}
	if e := valid.Validate(now); e != nil {
		t.Fatal(e)
	}
	for _, v := range []float64{math.NaN(), math.Inf(1), -1, 101} {
		i := valid
		i.Measurements.OxygenPercent = &v
		if i.Validate(now) == nil {
			t.Fatal("invalid oxygen accepted", v)
		}
	}
	i := valid
	i.Measurements = Measurements{}
	if i.Validate(now) == nil {
		t.Fatal("empty observation")
	}
	i = valid
	i.ObservedAt = now.Add(time.Second)
	if i.Validate(now) == nil {
		t.Fatal("future observation")
	}
	i = valid
	i.Measurements.SystolicMmHg = &oxygen
	if i.Validate(now) == nil {
		t.Fatal("incomplete blood pressure")
	}
}
