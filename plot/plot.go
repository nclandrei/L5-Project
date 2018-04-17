package plot

import (
	"fmt"
	"github.com/nclandrei/ticketguru/jira"
	"github.com/wcharczuk/go-chart"
	"github.com/wcharczuk/go-chart/drawing"
	"os"
)

const (
	graphsPath = "resources/graphs"
)

// Plot defines a standard analysis plotting function.
type Plot func(...jira.Ticket) error

// Attachments draws a stacked barchart for attachments analysis.
func Attachments(tickets ...jira.Ticket) error {
	return stackedBarChart("filename", 1, 2, 3)
}

// StepsToReproduce produces a barchart for presence of steps to reproduce in tickets.
func StepsToReproduce(tickets ...jira.Ticket) error {
	return nil
}

// Stacktraces produces a barchart for presence of stacktraces in tickets.
func Stacktraces(tickets ...jira.Ticket) error {
	return nil
}

// CommentsComplexity produces a scatter plot with trendline for comments complexity analysis.
func CommentsComplexity(tickets ...jira.Ticket) error {
	var comms []float64
	var times []float64
	for _, ticket := range tickets {
		if ticket.TimeToClose > 0 && ticket.CommentWordsCount > 0 {
			comms = append(comms, float64(ticket.CommentWordsCount))
			times = append(times, ticket.TimeToClose)
		}
	}
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	filePath := fmt.Sprintf("%s/%s/%s", wd, graphsPath, "comment_complexity.png")
	return scatter("Time-To-Close", "Comments Complexity", "Comments Complexity Analysis", filePath, comms, times)
}

// FieldsComplexity produces a scatter plot with trendline for fields (i.e. summary and description) complexity analysis.
func FieldsComplexity(tickets ...jira.Ticket) error {
	var fields []float64
	var times []float64
	for _, ticket := range tickets {
		if ticket.TimeToClose > 0 && ticket.SummaryDescWordsCount > 0 {
			fields = append(fields, float64(ticket.SummaryDescWordsCount))
			times = append(times, ticket.TimeToClose)
		}
	}
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	filePath := fmt.Sprintf("%s/%s/%s", wd, graphsPath, "fields_complexity.png")
	return scatter("Time-To-Close", "Fields Complexity", "Fields Complexity Analysis", filePath, fields, times)
}

// GrammarCorrectness produces a scatter plot with trendline for grammar correctness scores analysis.
func GrammarCorrectness(tickets ...jira.Ticket) error {
	return nil
}

// SentimentAnalysis produces a scatter plot with trendline for sentiment scores analysis.
func SentimentAnalysis(tickets ...jira.Ticket) error {
	return nil
}

func scatter(xAxis, yAxis, title, filepath string, xs []float64, ys []float64) error {
	viridisByY := func(xr, yr chart.Range, index int, x, y float64) drawing.Color {
		return chart.Viridis(y, yr.GetMin(), yr.GetMax())
	}

	s := chart.Chart{
		XAxis: chart.XAxis{
			Name:  xAxis,
			Style: chart.Style{Show: true},
		},
		YAxis: chart.YAxis{
			Name:  yAxis,
			Style: chart.Style{Show: true},
		},
		Title: title,
		Series: []chart.Series{
			chart.ContinuousSeries{
				Style: chart.Style{
					Show:             true,
					StrokeWidth:      chart.Disabled,
					DotWidth:         5,
					DotColorProvider: viridisByY,
				},
				XValues: xs,
				YValues: ys,
			},
		},
	}

	s.Elements = []chart.Renderable{
		chart.Legend(&s),
	}

	file, err := os.Create(filepath)
	if err != nil {
		return err
	}

	return s.Render(chart.PNG, file)
}

// stackedBarChart produces a
func stackedBarChart(filepath string, vals ...int) error {
	sbc := chart.StackedBarChart{
		Title:      "Presence and type of attachments analysis",
		TitleStyle: chart.StyleShow(),
		Background: chart.Style{
			Padding: chart.Box{
				Top: 50,
			},
			Show: true,
		},
		Height: 512,
		XAxis: chart.Style{
			Show: true,
		},
		YAxis: chart.Style{
			Show: true,
		},
		Bars: []chart.StackedBar{
			{
				Name: "With Attachments",
				Values: []chart.Value{
					{Value: 5, Label: "Blue"},
					{Value: 5, Label: "Green"},
					{Value: 4, Label: "Gray"},
					{Value: 3, Label: "Orange"},
					{Value: 3, Label: "Test"},
				},
			},
			{
				Name: "Without Attachments",
				Values: []chart.Value{
					{Value: 10, Label: "Blue"},
					{Value: 5, Label: "Green"},
					{Value: 1, Label: "Gray"},
				},
			},
		},
	}

	file, err := os.Create(filepath)
	if err != nil {
		return err
	}

	return sbc.Render(chart.PNG, file)
}
