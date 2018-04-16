package analyze

import (
	"regexp"
	"strings"
	"time"

	"github.com/nclandrei/ticketguru/jira"
)

// TicketAnalysis defines a function that analyzes a variadic number of tickets and updates
// their metrics fields accordingly.
type TicketAnalysis func(...jira.Ticket)

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

// CommentsComplexity counts the number of words in all comments for a variadic number of tickets.
func CommentsComplexity(tickets ...jira.Ticket) {
	for i := range tickets {
		if isTicketHighPriority(tickets[i]) {
			tickets[i].CommentWordsCount = calculateNumberOfWords(concatComments(tickets[i]))
		}
	}
}

// Attachments takes a variadic number of tickets and checks if they have attachments and what type they are.
func Attachments(tickets ...jira.Ticket) {
	for i := range tickets {
		if isTicketHighPriority(tickets[i]) {
			for _, attachment := range tickets[i].Fields.Attachments {

			}
		}
	}
}

func attachmentType(a jira.Attachment) jira.AttachmentType {

}

// StepsToReproduce returns whether a variadic number of tickets have steps to reproduce or not inside
// summary, description or any of the comments.
func StepsToReproduce(tickets ...jira.Ticket) {
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

// StackTraces checks whether a variadic number of tickets have stack traces attached either
// inside the description or any of the comments.
func StackTraces(tickets ...jira.Ticket) {
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

// concatComments returns a string containing all the comment bodies concatenated.
func concatComments(ticket jira.Ticket) string {
	var builder strings.Builder
	for _, comment := range ticket.Fields.Comments.Comments {
		builder.WriteString(comment.Body)
	}
	return builder.String()
}

// calculateTimeDifference calculates the duration in hours between 2 different timestamps.
func calculateTimeDifference(t1, t2 jira.Time) float64 {
	return time.Time(t1).Sub(time.Time(t2)).Hours()
}

// isTicketHighPriority checks whether an ticket has priority ID either 1 or 2 (i.e. Critical or Major).
func isTicketHighPriority(ticket jira.Ticket) bool {
	return ticket.Fields.Priority.ID == "1" || ticket.Fields.Priority.ID == "2"
}
