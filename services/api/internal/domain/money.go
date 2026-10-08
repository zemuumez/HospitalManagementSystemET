package domain

import "math"

const (
	// MaxMoneyMinor defines the maximum supported monetary amount in minor units (9 quadrillion = 90 trillion).
	// This fits comfortably within JavaScript Number.MAX_SAFE_INTEGER (9,007,199,254,740,991) and prevents
	// int64 overflow even when multiplied by 100 during percentage calculations.
	MaxMoneyMinor = int64(9_000_000_000_000_000)
	MaxQuantity   = 1_000_000
	MaxLines      = 500
)

func safeAdd(a, b int64) (int64, error) {
	if b > 0 && a > math.MaxInt64-b {
		return 0, ErrValidation
	}
	if b < 0 && a < math.MinInt64-b {
		return 0, ErrValidation
	}
	res := a + b
	if res > MaxMoneyMinor || res < 0 {
		return 0, ErrValidation
	}
	return res, nil
}

func safeMul(a, b int64) (int64, error) {
	if a == 0 || b == 0 {
		return 0, nil
	}
	if a > 0 && b > 0 && a > math.MaxInt64/b {
		return 0, ErrValidation
	}
	if a > 0 && b < 0 && b < math.MinInt64/a {
		return 0, ErrValidation
	}
	if a < 0 && b > 0 && a < math.MinInt64/b {
		return 0, ErrValidation
	}
	if a < 0 && b < 0 && a < math.MaxInt64/b {
		return 0, ErrValidation
	}
	res := a * b
	if res > MaxMoneyMinor || res < 0 {
		return 0, ErrValidation
	}
	return res, nil
}
