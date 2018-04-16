package analyze

import (
	"regexp"
	"strings"
	"time"

	"github.com/nclandrei/ticketguru/jira"
)

// TimesToClose returns how much time it took to close a variadic number of tickets.
func TimesToClose(tickets ...jira.Ticket) {
	for i := range tickets {
		if isTicketHighPriority(tickets[i]) {
			continue
		}
		var complete bool
		for _, history := range tickets[i].Changelog.Histories {
			for _, item := range history.Items {
				if item.Field == "status" && item.FromString == "Open" && item.ToString == "Closed" {
					tickets[i].TimeToClose = calculateTimeDifference(history.Created, tickets[i].Fields.Created)
					complete = true
				}
			}
		}
		if !complete {
			tickets[i].TimeToClose = -1
		}
	}
}

// CountWordsSummaryDesc counts the number of words in summary and description for a variadic number of tickets.
func CountWordsSummaryDesc(tickets ...jira.Ticket) {
	for i := range tickets {
		if isTicketHighPriority(tickets[i]) {
			tickets[i].SummaryDescWordsCount = calculateNumberOfWords(tickets[i].Fields.Description) +
				calculateNumberOfWords(tickets[i].Fields.Summary)
		}
	}
}

// HasStepsToReproduce returns whether an ticket has steps to reproduce or not inside either
// description or any of the comments.
func HasStepsToReproduce(tickets ...jira.Ticket) {
	expr := `(\n(\s*)\*(.*)){2,}`
	for i := range tickets {
		if !isTicketHighPriority(tickets[i]) {
			continue
		}
		contains := containsRegex(tickets[i].Fields.Description, expr)
		if contains {
			tickets[i].HasStepsToReproduce = true
			continue
		}
		for _, comment := range tickets[i].Fields.Comments.Comments {
			contains = containsRegex(comment.Body, expr)
			if contains {
				tickets[i].HasStepsToReproduce = true
				break
			}
		}
		if !contains {
			tickets[i].HasStepsToReproduce = false
		}
	}
}

// HasStackTrace checks whether a variadic number of tickets have stack traces attached either
// inside the description or any of the comments.
func HasStackTrace(tickets ...jira.Ticket) {
	expr := `^.+Exception[^\n]+\n(\s*at.+\s*\n)+`
	for i := range tickets {
		if !isTicketHighPriority(tickets[i]) {
			continue
		}
		contains := containsRegex(tickets[i].Fields.Description, expr)
		if contains {
			tickets[i].HasStackTrace = true
			continue
		}
		for _, comment := range tickets[i].Fields.Comments.Comments {
			contains = containsRegex(comment.Body, expr)
			if contains {
				tickets[i].HasStackTrace = true
				break
			}
		}
		if !contains {
			tickets[i].HasStackTrace = false
		}
	}
}

func containsRegex(s, expr string) bool {
	regex := regexp.MustCompile(expr)
	return regex.FindStringIndex(s) != nil
}

// calculateNumberOfWords returns the number of words in a string.
func calculateNumberOfWords(s string) int {
	wordCount := 0
	lines := strings.Split(s, "\n")
	for _, line := range lines {
		wordCount += len(strings.Split(strings.TrimSpace(line), " "))
	}
	return wordCount
}

// concatAndRemoveNewLines takes a variadic number of strings and returns a concatenated form with
// all of them having newlines replaced by whitespaces.
func concatAndRemoveNewlines(strs ...string) (string, error) {
	var strBuilder strings.Builder
	for _, str := range strs {
		str = strings.Replace(str, "\n", " ", -1)
		_, err := strBuilder.WriteString(str)
		if err != nil {
			return "", err
		}
		_, err = strBuilder.WriteRune(' ')
		if err != nil {
			return "", err
		}
	}
	return strBuilder.String(), nil
}

// ConcatenateComments returns a string containing all the comment bodies concatenated.
func concatenateComments(ticket jira.Ticket) (string, error) {
	var builder strings.Builder
	for _, comment := range ticket.Fields.Comments.Comments {
		if _, err := builder.WriteString(comment.Body); err != nil {
			return "", err
		}
	}
	return builder.String(), nil
}

// calculateTimeDifference calculates the duration in hours between 2 different timestamps.
func calculateTimeDifference(t1, t2 jira.Time) float64 {
	return time.Time(t1).Sub(time.Time(t2)).Hours()
}

// isTicketHighPriority checks whether an ticket has priority ID either 1 or 2 (i.e. Critical or Major).
func isTicketHighPriority(ticket jira.Ticket) bool {
	return ticket.Fields.Priority.ID == "1" || ticket.Fields.Priority.ID == "2"
}
