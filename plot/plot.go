package plot

import (
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/vg"
	"math/rand"
)

// NewPlotter returns a new plot.Plotter
func NewPlotter() (*plot.Plot, error) {
	p, err := plot.New()
	return p, err
}

// Draw computes a plot given a slice of XY points
func (p *plot.Plot) Draw(plotName, xLabel, yLabel string, pointSlice ...plotter.XYs) error {
	p.Title.Text = plotName
	p.X.Label.Text = xLabel
	p.Y.Label.Text = yLabel

	err := plotutil.AddLinePoints(p,
		"First", randomPoints(15),
		"Second", randomPoints(15),
		"Third", randomPoints(15))
	if err != nil {
		return err
	}

	// Save the plot to a PNG file.
	return p.Save(4*vg.Inch, 4*vg.Inch, "points.png")
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
