package plot

import (
	"fmt"

	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"

	"gonum.org/v1/gonum/floats"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/vg"
)

// JiraPlotter defines the plotter we use throughout the project
type JiraPlotter struct {
	*plot.Plot
	path string
}

// NewPlotter returns a new plot.Plotter
func NewPlotter() (*JiraPlotter, error) {
	p, err := plot.New()
	return &JiraPlotter{
		Plot: p,
		path: "resources/graphs",
	}, err
}

// DrawAttachmentsBarchart computes a barchart given two slices of floats
func (p *JiraPlotter) DrawAttachmentsBarchart(plotName, yLabel string, sl1, sl2 []float64) error {
	barA := plotter.Values{floats.Sum(sl1) / float64(len(sl1))}
	barB := plotter.Values{floats.Sum(sl2) / float64(len(sl2))}

	p.Title.Text = plotName
	p.Y.Label.Text = yLabel

	w := vg.Points(20)
	barChartA, err := plotter.NewBarChart(barA, w)
	if err != nil {
		return err
	}
	barChartA.LineStyle.Width = vg.Length(0)
	barChartA.Color = plotutil.Color(0)
	barChartA.Offset = -w

	barChartB, err := plotter.NewBarChart(barB, w)
	if err != nil {
		return err
	}
	barChartB.LineStyle.Width = vg.Length(0)
	barChartB.Color = plotutil.Color(1)

	p.Add(barChartA, barChartB)
	p.Legend.Add("With Attachments", barChartA)
	p.Legend.Add("Without Attachments", barChartB)

	return p.Save(8*vg.Inch, 8*vg.Inch, fmt.Sprintf("%s/%s.png", p.path, plotName))
}

func computeAverage(els []float64) float64 {
	return floats.Sum(els) / float64(len(els))
}
