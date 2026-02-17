package testdata

// SELECT文
var selectSQL = `select u.id, u.name, u.email from users u where u.status = 'active' and u.role = 'admin' order by u.created_at desc limit 10`

// INSERT文
var insertSQL = `insert into users (id, name, email, status, created_at) values (?, ?, ?, 'active', NOW())`

// UPDATE文
var updateSQL = `update users set name = ?, email = ?, updated_at = NOW() where id = ? and deleted_at is null`

// DELETE文
var deleteSQL = `delete from users where id = ? and status = 'inactive' and created_at < '2020-01-01'`

// 短いSQL（80文字以下、1行のまま）
var shortSelectSQL = `select * from hoge where id = ?`

// 非SQL文字列（変更されない）
var notSQL = `This is just a regular string, not SQL`

// JSON文字列（変更されない）
var jsonStr = `{"users": [{"id": 1, "name": "test"}]}`

// 空文字列（変更されない）
var emptyStr = ``
