SELECT issue.key, attachment.filename, issue.created AS issue_created, changelog_history.time_diff, changelog_history_item.field, changelog_history_item.from_string, changelog_history_item.to_string, changelog_history.created AS ch_created
FROM issue
INNER JOIN priority ON priority.issue_key = issue.key
INNER JOIN changelog_history ON changelog_history.issue_key = issue.key
INNER JOIN changelog_history_item ON changelog_history.id = changelog_history_item.history_id
INNER JOIN attachment ON attachment.issue_key = issue.key
WHERE (priority.id = '1' or priority.id = '2') and changelog_history_item.field = 'status' and changelog_history_item.from_string = 'Open' and changelog_history_item.to_string = 'Closed' and attachment.attachment_type = 1
GROUP BY issue.key, attachment.filename, changelog_history.time_diff, changelog_history_item.field, changelog_history_item.from_string, changelog_history_item.to_string, changelog_history.created
ORDER BY changelog_history.time_diff;