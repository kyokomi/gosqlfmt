package testdata

// SELECT文
var selectSQL = `
SELECT
  u.id,
  u.name,
  u.email
FROM
  users u
WHERE
  u.status = 'active'
  AND u.role = 'admin'
ORDER BY
  u.created_at DESC
LIMIT
  10
`

// INSERT文
var insertSQL = `
INSERT INTO
  users (id, name, email, status, created_at)
VALUES
  (?, ?, ?, 'active', NOW())
`

// UPDATE文
var updateSQL = `
UPDATE
  users
SET
  name = ?,
  email = ?,
  updated_at = NOW()
WHERE
  id = ?
  AND deleted_at IS NULL
`

// DELETE文
var deleteSQL = `
DELETE FROM
  users
WHERE
  id = ?
  AND status = 'inactive'
  AND created_at < '2020-01-01'
`

// 短いSQL（80文字以下、1行のまま）
var shortSelectSQL = `SELECT * FROM hoge WHERE id = ?`

// 非SQL文字列（変更されない）
var notSQL = `This is just a regular string, not SQL`

// JSON文字列（変更されない）
var jsonStr = `{"users": [{"id": 1, "name": "test"}]}`

// 空文字列（変更されない）
var emptyStr = ``
