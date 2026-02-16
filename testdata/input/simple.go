package testdata

var shortSQL = `select * from hoge where id = ?`

var longSQL = `SELECT * FROM hoge WHERE id = ? AND status = 'active' ORDER BY created_at DESC LIMIT 1`

func example() {
	fugasql := `select id, name from fuga where id = :id and status = :status order by created_at desc limit 30`
	_ = fugasql

	insertSQL := `insert into users (id, name, email, status) values (?, ?, ?, 'active')`
	_ = insertSQL

	notSQL := `This is not a SQL query, just a regular string`
	_ = notSQL

	jsonStr := `{"key": "value", "select": "from"}`
	_ = jsonStr
}
