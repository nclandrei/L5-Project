package stats

import (
	"errors"
	"github.com/dgryski/go-onlinestats"
	"github.com/nclandrei/ticketguru/jira"
	"math"
)

var (
	errSampleSize   = errors.New("sample is too small")
	errZeroVariance = errors.New("sample has zero variance")
)

// Stats defines the basic slice of float64 used for statistical tests.
type stats []float64

// Weight calculates the default weight of a Stats value.
func (s stats) Weight() float64 {
	return float64(len(s))
}

// Len returns the length of the underlying slice.
func (s stats) Len() int {
	return len(s)
}

// Mean calculates the mean of the variables inside the underlying slice of a Stats value.
func (s stats) Mean() float64 {
	var total float64
	for _, n := range s {
		total += n
	}
	return total / float64(len(s))
}

// Variance returns the variance of the underlying slice of a Stats value.
func (s stats) Variance() float64 {
	var total float64
	mean := s.Mean()
	for _, number := range s {
		total += math.Pow(number-mean, 2)
	}
	return total / float64(len(s)-1)
}

// CategoricalTest defines a function that takes a variadic number of tickets and computes Welch's T test
// on fields of them.
type CategoricalTest func(...jira.Ticket) (*TTestResult, error)

// ContinuousTest defines a function that takes a variadic number of tickets and computes a Spearman R test
// on fields of them.
type ContinuousTest func(...jira.Ticket) *SpearmanResult

// Attachments performs Welch's T Test on all tickets' attachments.
func Attachments(tickets ...jira.Ticket) (*TTestResult, error) {
	var withTimes stats
	var withoutTimes stats
	for _, t := range tickets {
		highPriority := jira.IsHighPriority(t)
		if t.TimeToClose <= 0 ||
			t.TimeToClose > jira.MaxTimeToCloseH ||
			!highPriority {
			continue
		}
		if len(t.Fields.Attachments) > 0 {
			withTimes = append(withTimes, t.TimeToClose)
		} else {
			withoutTimes = append(withoutTimes, t.TimeToClose)
		}
	}
	return twoSampleWelchTTest(withTimes, withoutTimes)
}

// StepsToReproduce performs Welch's T Test on steps to reproduce presence or not for all tickets.
func StepsToReproduce(tickets ...jira.Ticket) (*TTestResult, error) {
	var withTimes stats
	var withoutTimes stats
	for _, t := range tickets {
		highPriority := jira.IsHighPriority(t)
		if t.TimeToClose <= 0 ||
			t.TimeToClose > jira.MaxTimeToCloseH ||
			!highPriority {
			continue
		}
		if t.HasStepsToReproduce {
			withTimes = append(withTimes, t.TimeToClose)
		} else {
			withoutTimes = append(withoutTimes, t.TimeToClose)
		}
	}
	return twoSampleWelchTTest(withTimes, withoutTimes)
}

// Stacktraces performs Welch's T Test on stack traces presence or not for all tickets.
func Stacktraces(tickets ...jira.Ticket) (*TTestResult, error) {
	var withTimes stats
	var withoutTimes stats
	for _, t := range tickets {
		highPriority := jira.IsHighPriority(t)
		if t.TimeToClose <= 0 ||
			t.TimeToClose > jira.MaxTimeToCloseH ||
			!highPriority {
			continue
		}
		if t.HasStackTrace {
			withTimes = append(withTimes, t.TimeToClose)
		} else {
			withoutTimes = append(withoutTimes, t.TimeToClose)
		}
	}
	return twoSampleWelchTTest(withTimes, withoutTimes)
}

// CommentsComplexity performs Spearman R's test on the complexity of comments and times-to-close.
func CommentsComplexity(tickets ...jira.Ticket) *SpearmanResult {
	var comms stats
	var times stats
	for _, t := range tickets {
		highPriority := jira.IsHighPriority(t)
		if highPriority &&
			t.TimeToClose > 0 &&
			t.TimeToClose < jira.MaxTimeToCloseH &&
			t.CommentWordsCount > 0 &&
			t.CommentWordsCount < jira.MaxCommWordCount {
			comms = append(comms, float64(t.CommentWordsCount))
			times = append(times, t.TimeToClose)
		}
	}
	return twoSampleSpearmanRTest(comms, times)
}

// FieldsComplexity performs Spearman R's test on the complexity of summary&description and times-to-close.
func FieldsComplexity(tickets ...jira.Ticket) *SpearmanResult {
	var fields stats
	var times stats
	for _, t := range tickets {
		highPriority := jira.IsHighPriority(t)
		if highPriority &&
			t.TimeToClose > 0 &&
			t.TimeToClose <= jira.MaxTimeToCloseH &&
			t.SummaryDescWordsCount > 0 &&
			t.SummaryDescWordsCount < jira.MaxSummaryDescWordCount {
			fields = append(fields, float64(t.SummaryDescWordsCount))
			times = append(times, t.TimeToClose)
		}
	}
	return twoSampleSpearmanRTest(fields, times)
}

// Sentiment performs Spearman R's test on sentiment scores and times-to-close.
func Sentiment(tickets ...jira.Ticket) *SpearmanResult {
	var scores stats
	var times stats
	for _, t := range tickets {
		highPriority := jira.IsHighPriority(t)
		if highPriority &&
			t.TimeToClose > 0 &&
			t.TimeToClose <= jira.MaxTimeToCloseH &&
			t.Sentiment.HasScore {
			scores = append(scores, t.Sentiment.Score)
			times = append(times, t.TimeToClose)
		}
	}
	return twoSampleSpearmanRTest(scores, times)
}

// Grammar performs Spearman R's test on grammar correctness scores and times-to-close.
func Grammar(tickets ...jira.Ticket) *SpearmanResult {
	var scores stats
	var times stats
	for _, t := range tickets {
		highPriority := jira.IsHighPriority(t)
		if highPriority &&
			t.TimeToClose > 0 &&
			t.TimeToClose <= jira.MaxTimeToCloseH &&
			t.GrammarCorrectness.HasScore &&
			t.GrammarCorrectness.Score < jira.MaxGrammarErrCount {
			scores = append(scores, float64(t.GrammarCorrectness.Score))
			times = append(times, t.TimeToClose)
		}
	}
	return twoSampleSpearmanRTest(scores, times)
}

// twoSampleSpearmanRTest returns the rank correlation coefficient and p value given two samples.
func twoSampleSpearmanRTest(xs, ys stats) *SpearmanResult {
	rs, p := onlinestats.Spearman(xs, ys)
	return &SpearmanResult{
		Rs: rs,
		P:  p,
	}
}

// twoSampleWelchTTest computes the result of a Welch T Test given two samples.
func twoSampleWelchTTest(x1, x2 Sample) (*TTestResult, error) {
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
	return newTTestResult(int(n1), int(n2), t, dof, x1.Mean(), x2.Mean()), nil
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
	N1Mean float64
	N2Mean float64
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

func newTTestResult(n1, n2 int, t, dof, n1Mean, n2Mean float64) *TTestResult {
	dist := tDist{dof}
	p := 2 * (1 - dist.cdf(math.Abs(t)))
	return &TTestResult{
		N1:     n1,
		N2:     n2,
		T:      t,
		DoF:    dof,
		P:      p,
		N1Mean: n1Mean,
		N2Mean: n2Mean,
	}
}
