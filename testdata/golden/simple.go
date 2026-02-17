package testdata

var shortSQL = `SELECT * FROM hoge WHERE id = ?`

var longSQL = `
SELECT
  *
FROM
  hoge
WHERE
  id = ?
  AND status = 'active'
ORDER BY
  created_at DESC
LIMIT
  1
`

func example() {
	fugasql := `
SELECT
  id,
  name
FROM
  fuga
WHERE
  id = :id
  AND status = :status
ORDER BY
  created_at DESC
LIMIT
  30
`
	_ = fugasql

	insertSQL := `
INSERT INTO
  users (id, name, email, created_at, status)
VALUES
  (?, ?, ?, NOW(), 'active')
`
	_ = insertSQL

	notSQL := `This is not a SQL query, just a regular string`
	_ = notSQL

	jsonStr := `{"key": "value", "select": "from"}`
	_ = jsonStr
}
