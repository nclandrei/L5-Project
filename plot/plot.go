package plot

import (
	"github.com/nclandrei/ticketguru/jira"
	"github.com/wcharczuk/go-chart"
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
	return nil
}

// FieldsComplexity produces a scatter plot with trendline for fields (i.e. summary and description) complexity analysis.
func FieldsComplexity(tickets ...jira.Ticket) error {
	return nil
}

// GrammarCorrectness produces a scatter plot with trendline for grammar correctness scores analysis.
func GrammarCorrectness(tickets ...jira.Ticket) error {
	return nil
}

// SentimentAnalysis produces a scatter plot with trendline for sentiment scores analysis.
func SentimentAnalysis(tickets ...jira.Ticket) error {
	return nil
}

// stackedBarChart produces a
func stackedBarChart(filename string, vals ...int) error {
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

	return sbc.Render(chart.PNG, nil)
}
