package plot

import (
	"fmt"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/vg"
	"math/rand"
)

// JiraPlotter defines the plotter we use throughout the project
type JiraPlotter plot.Plot

// NewPlotter returns a new plot.Plotter
func NewPlotter() (*JiraPlotter, error) {
	p, err := plot.New()
	return (*JiraPlotter)(p), err
}

// Draw computes a plot given a slice of XY points
func (p *JiraPlotter) Draw(plotName, xLabel, yLabel string, pointSlice ...plotter.XYs) error {
	p.Title.Text = plotName
	p.X.Label.Text = xLabel
	p.Y.Label.Text = yLabel

	err := plotutil.AddLinePoints((*plot.Plot)(p), pointSlice)
	if err != nil {
		return err
	}

	// Save the plot to a PNG file.
	return (*plot.Plot)(p).Save(4*vg.Inch, 4*vg.Inch, fmt.Sprintf("%s.png", plotName))
}

// randomPoints returns some random x, y points.
func randomPoints(n int) plotter.XYs {
	pts := make(plotter.XYs, n)
	for i := range pts {
		if i == 0 {
			pts[i].X = rand.Float64()
		} else {
			pts[i].X = pts[i-1].X + rand.Float64()
		}
		pts[i].Y = pts[i].X + 10*rand.Float64()
	}
	return pts
}
