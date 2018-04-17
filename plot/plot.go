package plot

import (
	"fmt"
	"github.com/nclandrei/ticketguru/analyze"
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
	var withCount int
	var withSum, withoutSum float64
	for _, ticket := range tickets {
		if !analyze.IsTicketHighPriority(ticket) ||
			ticket.TimeToClose <= 0 ||
			ticket.TimeToClose > 9000 {
			continue
		}
		if ticket.HasStepsToReproduce {
			withCount++
			withSum += ticket.TimeToClose
		} else {
			withoutSum += ticket.TimeToClose
		}
	}
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	return barchart("Steps To Reproduce Analysis",
		fmt.Sprintf("%s/%s/%s", wd, graphsPath, "steps_to_reproduce.png"),
		map[float64]string{
			withSum / float64(withCount):                 "With Steps to Reproduce",
			withoutSum / float64(len(tickets)-withCount): "Without Steps to Reproduce",
		},
	)
}

// Stacktraces produces a barchart for presence of stacktraces in tickets.
func Stacktraces(tickets ...jira.Ticket) error {
	var withCount int
	var withSum, withoutSum float64
	for _, ticket := range tickets {
		if !analyze.IsTicketHighPriority(ticket) ||
			ticket.TimeToClose <= 0 ||
			ticket.TimeToClose > 9000 {
			continue
		}
		if ticket.HasStackTrace {
			withCount++
			withSum += ticket.TimeToClose
		} else {
			withoutSum += ticket.TimeToClose
		}
	}
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	return barchart("Stack Traces Analysis",
		fmt.Sprintf("%s/%s/%s", wd, graphsPath, "stack_traces.png"),
		map[float64]string{
			withSum / float64(withCount):                 "With Stack Traces",
			withoutSum / float64(len(tickets)-withCount): "Without Stack Traces",
		},
	)
}

// CommentsComplexity produces a scatter plot with trendline for comments complexity analysis.
func CommentsComplexity(tickets ...jira.Ticket) error {
	var comms []float64
	var times []float64
	for _, ticket := range tickets {
		if analyze.IsTicketHighPriority(ticket) &&
			ticket.TimeToClose > 0 &&
			ticket.TimeToClose < 9000 &&
			ticket.CommentWordsCount > 0 {
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
		if analyze.IsTicketHighPriority(ticket) &&
			ticket.TimeToClose > 0 &&
			ticket.TimeToClose <= 9000 &&
			ticket.SummaryDescWordsCount > 0 {
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
	var scores []float64
	var times []float64
	for _, ticket := range tickets {
		if analyze.IsTicketHighPriority(ticket) &&
			ticket.TimeToClose > 0 &&
			ticket.TimeToClose <= 9000 &&
			ticket.GrammarCorrectness.HasScore {
			scores = append(scores, float64(ticket.GrammarCorrectness.Score))
			times = append(times, ticket.TimeToClose)
		}
	}
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	filePath := fmt.Sprintf("%s/%s/%s", wd, graphsPath, "grammar_correctness.png")
	return scatter("Time-To-Close", "Grammar Correctness Score", "Grammar Correctness Analysis", filePath, scores, times)
}

// SentimentAnalysis produces a scatter plot with trendline for sentiment scores analysis.
func SentimentAnalysis(tickets ...jira.Ticket) error {
	var scores []float64
	var times []float64
	for _, ticket := range tickets {
		if analyze.IsTicketHighPriority(ticket) &&
			ticket.TimeToClose > 0 &&
			ticket.TimeToClose <= 9000 &&
			ticket.Sentiment.HasScore {
			scores = append(scores, ticket.Sentiment.Score)
			times = append(times, ticket.TimeToClose)
		}
	}
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	filePath := fmt.Sprintf("%s/%s/%s", wd, graphsPath, "sentiment_analysis.png")
	return scatter("Time-To-Close", "Sentiment Score", "Sentiment Analysis", filePath, scores, times)
}

// barchart computes and saves a barchart given a variadic number of bars.
func barchart(title, filepath string, vals map[float64]string) error {
	var bars []chart.Value
	for k, v := range vals {
		bars = append(bars, chart.Value{
			Value: k,
			Label: v,
		})
	}
	sbc := chart.BarChart{
		Title:      title,
		TitleStyle: chart.StyleShow(),
		Background: chart.Style{
			Padding: chart.Box{
				Top: 40,
			},
		},
		Height:   512,
		BarWidth: 60,
		XAxis: chart.Style{
			Show: true,
		},
		YAxis: chart.YAxis{
			Name: "Time-To-Close",
			NameStyle: chart.Style{
				Show: true,
			},
			Style: chart.Style{
				Show: true,
			},
		},
		Bars: bars,
	}

	file, err := os.Create(filepath)
	if err != nil {
		return err
	}

	return sbc.Render(chart.PNG, file)
}

func scatter(xAxis, yAxis, title, filepath string, xs []float64, ys []float64) error {
	viridisByY := func(xr, yr chart.Range, index int, x, y float64) drawing.Color {
		return chart.Viridis(y, yr.GetMin(), yr.GetMax())
	}

	s := chart.Chart{
		XAxis: chart.XAxis{
			Name:      xAxis,
			NameStyle: chart.Style{Show: true},
			Style:     chart.Style{Show: true},
		},
		YAxis: chart.YAxis{
			Name:      yAxis,
			NameStyle: chart.Style{Show: true},
			Style:     chart.Style{Show: true},
		},
		Title: title,
		TitleStyle: chart.Style{
			Show: true,
		},
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
