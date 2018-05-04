package analyze

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/nclandrei/ticketguru/jira"
)

// TicketAnalysis defines a function that analyzes a variadic number of tickets and updates
// their metrics fields accordingly.
type TicketAnalysis func(...jira.JiraIssue)

// TimesToClose returns how much time it took to close a variadic number of tickets.
func TimesToClose(tickets ...jira.JiraIssue) {
	var count int
	for i := range tickets {
		if !isTicketHighPriority(tickets[i]) || tickets[i].Fields.Status.Name == "Open" {
			continue
		}
		var closed bool
		for _, history := range tickets[i].Changelog.Histories {
			for _, item := range history.Items {
				if item.Field == "status" && (item.ToString == "Closed" || item.ToString == "Resolved" ||
					item.ToString == "Done" || item.ToString == "Completed" || item.ToString == "Fixed") {
					tickets[i].TimeToClose = calculateTimeDifference(history.Created, tickets[i].Fields.Created)
					count++
					closed = true
					break
				}
			}
			if closed {
				break
			}
		}
		if !closed {
			tickets[i].TimeToClose = 0
		}
	}
	fmt.Println(count)
}

// FieldsComplexity counts the number of words in summary and description for a variadic number of tickets.
func FieldsComplexity(tickets ...jira.JiraIssue) {
	for i := range tickets {
		if isTicketHighPriority(tickets[i]) {
			tickets[i].SummaryDescWordsCount = calculateNumberOfWords(tickets[i].Fields.Description) +
				calculateNumberOfWords(tickets[i].Fields.Summary)
		}
	}
}

// CommentsComplexity counts the number of words in all comments for a variadic number of tickets.
func CommentsComplexity(tickets ...jira.JiraIssue) {
	for i := range tickets {
		if isTicketHighPriority(tickets[i]) {
			tickets[i].CommentWordsCount = calculateNumberOfWords(concatComments(tickets[i]))
		}
	}
}

// Attachments takes a variadic number of tickets and checks if they have attachments and what type they are.
func Attachments(tickets ...jira.JiraIssue) {
	for i := range tickets {
		if isTicketHighPriority(tickets[i]) {
			for j := range tickets[i].Fields.Attachments {
				tickets[i].Fields.Attachments[j].Type = attachmentType(tickets[i].Fields.Attachments[j])
			}
		}
	}
}

// attachmentType returns the type of attachment.
func attachmentType(a jira.Attachment) jira.AttachmentType {
	extension := fileExtension(a.Filename)
	switch extension {
	case "png", "jpg", "jpeg", "gif", "bmp", "tiff", "webp":
		return jira.ImageAttachment
	case "md", "txt", "pdf", "doc", "docx", "pages":
		return jira.TextAttachment
	case "go", "java", "groovy", "rs", "clj", "py", "rb", "jar", "php", "js", "c", "cpp",
		"h", "sh", "bat", "bin", "apk", "pl", "ex", "exs":
		return jira.CodeAttachment
	case "avi", "mkv", "mp4", "flv", "wmv", "mov":
		return jira.VideoAttachment
	case "xml", "json", "yml", "toml", "bson", "env":
		return jira.ConfigAttachment
	case "tar", "zip", "rar", "tgz", "7z", "z":
		return jira.ArchiveAttachment
	case "csv", "xls", "xslx", "numbers":
		return jira.SpreadsheetAttachment
	default:
		return jira.OtherAttachment
	}
}

// fileExtension returns the lowercased extension of a file given that file's name.
func fileExtension(f string) string {
	return strings.ToLower(f[(strings.LastIndex(f, ".") + 1):])
}

// StepsToReproduce returns whether a variadic number of tickets have steps to reproduce or not inside
// summary, description or any of the comments.
func StepsToReproduce(tickets ...jira.JiraIssue) {
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
func StackTraces(tickets ...jira.JiraIssue) {
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
func concatComments(ticket jira.JiraIssue) string {
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

// isTicketHighPriority checks whether a ticket is high priority.
func isTicketHighPriority(ticket jira.JiraIssue) bool {
	return ticket.Fields.Priority.ID == "1" || ticket.Fields.Priority.ID == "2" ||
		ticket.Fields.Priority.ID == "3" || ticket.Fields.Priority.ID == "4"
}
