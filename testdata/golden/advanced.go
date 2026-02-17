package testdata

// BETWEEN...AND
var betweenSQL = `
SELECT
  *
FROM
  products
WHERE
  price BETWEEN 100 AND 200
  AND status = 'active'
ORDER BY
  price ASC
`

// CASE WHEN
var caseSQL = `
SELECT
  id,
  CASE
    WHEN status = 1 THEN 'active'
    WHEN status = 2 THEN 'inactive'
    ELSE 'unknown'
  END AS status_label
FROM
  users
`

// ブロックコメント
var commentSQL = `
SELECT
  id,
  /* primary key */ name
FROM
  users
WHERE
  status = 'active'
ORDER BY
  id DESC
LIMIT
  10
`

// FROM句のサブクエリ
var subquerySQL = `
SELECT
  *
FROM
  (
    SELECT
      id,
      name
    FROM
      users
    WHERE
      status = ?
  ) t
WHERE
  t.id > 0
ORDER BY
  t.id
`

// WHERE IN サブクエリ
var whereInSQL = `
SELECT
  *
FROM
  users
WHERE
  id IN (
    SELECT
      user_id
    FROM
      orders
    WHERE
      status = 'active'
  )
  AND role = 'admin'
`

// EXISTS サブクエリ
var existsSQL = `
SELECT
  *
FROM
  users u
WHERE
  EXISTS (
    SELECT
      1
    FROM
      orders o
    WHERE
      o.user_id = u.id
      AND o.status = 'active'
  )
`

// 複合: CASE + サブクエリ + BETWEEN
var complexSQL = `
SELECT
  u.id,
  CASE
    WHEN u.role = 'admin' THEN 'Admin'
    ELSE 'User'
  END AS role_label,
  u.name
FROM
  users u
WHERE
  u.created_at BETWEEN '2020-01-01' AND '2024-12-31'
  AND u.id IN (
    SELECT
      user_id
    FROM
      orders
    WHERE
      total > 100
  )
ORDER BY
  u.id
`

func advancedExample() {
	// 関数呼び出し内のカンマ保護
	funcSQL := `
SELECT
  coalesce(u.name, u.email, 'unknown'),
  count(DISTINCT u.id)
FROM
  users u
WHERE
  u.status = 'active'
GROUP BY
  u.role
`
	_ = funcSQL

	// 通常の文字列（変更されないこと）
	notSQL := `This contains SELECT but is not SQL`
	_ = notSQL
}
