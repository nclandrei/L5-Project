SELECT issue.key, SUM(array_length(regexp_split_to_array(issue.SUMMARY, '\s'), 1)) AS WORD_COUNT
FROM issue
INNER JOIN priority ON priority.issue_key = issue.key
WHERE priority.id = '1'
GROUP BY key;