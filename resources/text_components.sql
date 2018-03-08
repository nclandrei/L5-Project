SELECT issue.key, issue.created AS issue_created, issue.wc_summary, issue.wc_description, changelog_history.time_diff, changelog_history_item.field, changelog_history_item.from_string, changelog_history_item.to_string, changelog_history.created AS ch_created
FROM issue
INNER JOIN priority ON priority.issue_key = issue.key
INNER JOIN changelog_history ON changelog_history.issue_key = issue.key
INNER JOIN changelog_history_item ON changelog_history.id = changelog_history_item.history_id
WHERE (priority.id = '1' or priority.id = '2') and changelog_history_item.field = 'status' and changelog_history_item.from_string = 'Open' and changelog_history_item.to_string = 'Closed'
GROUP BY issue.key, issue.wc_summary, issue.wc_description, changelog_history.time_diff, changelog_history_item.field, changelog_history_item.from_string, changelog_history_item.to_string, changelog_history.created
ORDER BY changelog_history.time_diff;