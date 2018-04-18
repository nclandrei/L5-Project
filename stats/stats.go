package stats

import (
	"errors"
	"github.com/dgryski/go-onlinestats"
	"math"
)

var (
	errSampleSize   = errors.New("sample is too small")
	errZeroVariance = errors.New("sample has zero variance")
)

// Stats defines the basic slice of float64 used for statistical tests.
type Stats []float64

// Weight calculates the default weight of a Stats value.
func (s Stats) Weight() float64 {
	return float64(len(s))
}

// Len returns the length of the underlying slice.
func (s Stats) Len() int {
	return len(s)
}

// Mean calculates the mean of the variables inside the underlying slice of a Stats value.
func (s Stats) Mean() float64 {
	var total float64
	for _, n := range s {
		total += n
	}
	return total / float64(len(s))
}

// Variance returns the variance of the underlying slice of a Stats value.
func (s Stats) Variance() float64 {
	var total float64
	mean := s.Mean()
	for _, number := range s {
		total += math.Pow(number-mean, 2)
	}
	return total / float64(len(s)-1)
}

// TwoSampleSpearmanRTest returns the rank correlation coefficient and p value given two samples.
func TwoSampleSpearmanRTest(xs, ys Stats) *SpearmanResult {
	rs, p := onlinestats.Spearman(xs, ys)
	return &SpearmanResult{
		Rs: rs,
		P:  p,
	}
}

// TwoSampleWelchTTest computes the result of a Welch T Test given two samples.
func TwoSampleWelchTTest(x1, x2 Sample) (*TTestResult, error) {
	n1, n2 := x1.Weight(), x2.Weight()
	if n1 <= 1 || n2 <= 1 {
		return nil, errSampleSize
	}
	v1, v2 := x1.Variance(), x2.Variance()
	if v1 == 0 && v2 == 0 {
		return nil, errZeroVariance
	}

	dof := math.Pow(v1/n1+v2/n2, 2) /
		(math.Pow(v1/n1, 2)/(n1-1) + math.Pow(v2/n2, 2)/(n2-1))
	s := math.Sqrt(v1/n1 + v2/n2)
	t := (x1.Mean() - x2.Mean()) / s
	return newTTestResult(int(n1), int(n2), t, dof), nil
}

// A Sample can be used to compute various statistical tests.
type Sample interface {
	Weight() float64
	Mean() float64
	Variance() float64
}

// A TTestResult is the result of a t-test.
type TTestResult struct {
	N1, N2 int
	T      float64
	DoF    float64
	P      float64
}

// A SpearmanResult is the result of Spearman's rank correlation coefficient.
type SpearmanResult struct {
	Rs float64
	P  float64
}

type tDist struct {
	V float64
}

func (t tDist) cdf(x float64) float64 {
	if x == 0 {
		return 0.5
	} else if x > 0 {
		return 1 - 0.5*mathBetaInc(t.V/(t.V+x*x), t.V/2, 0.5)
	} else if x < 0 {
		return 1 - t.cdf(-x)
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
		bt = math.Exp(lgamma(a+b) - lgamma(a) - lgamma(b) +
			a*math.Log(x) + b*math.Log(1-x))
	}
	if x < (a+1)/(a+b+2) {
		return bt * betacf(x, a, b) / a
	}
	return 1 - bt*betacf(1-x, b, a)/b
}

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

		numer := mf * (b - mf) * x / ((a + 2*mf - 1) * (a + 2*mf))
		d = 1 / raiseZero(1+numer*d)
		c = raiseZero(1 + numer/c)
		h *= d * c

		numer = -(a + mf) * (a + b + mf) * x / ((a + 2*mf) * (a + 2*mf + 1))
		d = 1 / raiseZero(1+numer*d)
		c = raiseZero(1 + numer/c)
		hfac := d * c
		h *= hfac

		if math.Abs(hfac-1) < epsilon {
			return h
		}
	}
	panic("failed to converge")
}

func newTTestResult(n1, n2 int, t, dof float64) *TTestResult {
	dist := tDist{dof}
	p := 2 * (1 - dist.cdf(math.Abs(t)))
	return &TTestResult{
		N1:  n1,
		N2:  n2,
		T:   t,
		DoF: dof,
		P:   p,
	}
}
