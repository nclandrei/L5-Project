package stats

import (
	"errors"
	"math"
)

var (
	ErrSampleSize        = errors.New("sample is too small")
	ErrZeroVariance      = errors.New("sample has zero variance")
	ErrMismatchedSamples = errors.New("samples have different lengths")
)

// A TTestSample is a sample that can be used for a one or two sample
// t-test.
type TTestSample interface {
	Weight() float64
	Mean() float64
	Variance() float64
}

// A TTestResult is the result of a t-test.
type TTestResult struct {
	// N1 and N2 are the sizes of the input samples. For a
	// one-sample t-test, N2 is 0.
	N1, N2 int

	// T is the value of the t-statistic for this t-test.
	T float64

	// DoF is the degrees of freedom for this t-test.
	DoF float64

	// P is p-value for this t-test for the given null hypothesis.
	P float64
}

// A TDist is a Student's t-distribution with V degrees of freedom.
type TDist struct {
	V float64
}

func (t TDist) CDF(x float64) float64 {
	if x == 0 {
		return 0.5
	} else if x > 0 {
		return 1 - 0.5*mathBetaInc(t.V/(t.V+x*x), t.V/2, 0.5)
	} else if x < 0 {
		return 1 - t.CDF(-x)
	}
	return math.NaN()
}

func lgamma(x float64) float64 {
	y, _ := math.Lgamma(x)
	return y
}

func mathBetaInc(x, a, b float64) float64 {
	if x < 0 || x > 1 {
		return math.NaN()
	}
	bt := 0.0
	if 0 < x && x < 1 {
		// Compute the coefficient before the continued
		// fraction.
		bt = math.Exp(lgamma(a+b) - lgamma(a) - lgamma(b) +
			a*math.Log(x) + b*math.Log(1-x))
	}
	if x < (a+1)/(a+b+2) {
		// Compute continued fraction directly.
		return bt * betacf(x, a, b) / a
	}
	return 1 - bt*betacf(1-x, b, a)/b
}

// betacf is the continued fraction component of the regularized
// incomplete beta function Iₓ(a, b).
func betacf(x, a, b float64) float64 {
	const maxIterations = 200
	const epsilon = 3e-14

	raiseZero := func(z float64) float64 {
		if math.Abs(z) < math.SmallestNonzeroFloat64 {
			return math.SmallestNonzeroFloat64
		}
		return z
	}

	c := 1.0
	d := 1 / raiseZero(1-(a+b)*x/(a+1))
	h := d
	for m := 1; m <= maxIterations; m++ {
		mf := float64(m)

		// Even step of the recurrence.
		numer := mf * (b - mf) * x / ((a + 2*mf - 1) * (a + 2*mf))
		d = 1 / raiseZero(1+numer*d)
		c = raiseZero(1 + numer/c)
		h *= d * c

		// Odd step of the recurrence.
		numer = -(a + mf) * (a + b + mf) * x / ((a + 2*mf) * (a + 2*mf + 1))
		d = 1 / raiseZero(1+numer*d)
		c = raiseZero(1 + numer/c)
		hfac := d * c
		h *= hfac

		if math.Abs(hfac-1) < epsilon {
			return h
		}
	}
	panic("betainc: a or b too big; failed to converge")
}

func newTTestResult(n1, n2 int, t, dof float64) *TTestResult {
	dist := TDist{dof}
	p := 2 * (1 - dist.CDF(math.Abs(t)))
	return &TTestResult{N1: n1, N2: n2, T: t, DoF: dof, P: p}
}

func TwoSampleWelchTTest(x1, x2 TTestSample) (*TTestResult, error) {
	n1, n2 := x1.Weight(), x2.Weight()
	if n1 <= 1 || n2 <= 1 {
		return nil, ErrSampleSize
	}
	v1, v2 := x1.Variance(), x2.Variance()
	if v1 == 0 && v2 == 0 {
		return nil, ErrZeroVariance
	}

	dof := math.Pow(v1/n1+v2/n2, 2) /
		(math.Pow(v1/n1, 2)/(n1-1) + math.Pow(v2/n2, 2)/(n2-1))
	s := math.Sqrt(v1/n1 + v2/n2)
	t := (x1.Mean() - x2.Mean()) / s
	return newTTestResult(int(n1), int(n2), t, dof), nil
}
