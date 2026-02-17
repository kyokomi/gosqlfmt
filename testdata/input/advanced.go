package testdata

// BETWEEN...AND
var betweenSQL = `select * from products where price between 100 and 200 and status = 'active' order by price asc`

// CASE WHEN
var caseSQL = `select id, case when status = 1 then 'active' when status = 2 then 'inactive' else 'unknown' end as status_label from users`

// ブロックコメント
var commentSQL = `select id, /* primary key */ name from users where status = 'active' order by id desc limit 10`

// FROM句のサブクエリ
var subquerySQL = `select * from (select id, name from users where status = ?) t where t.id > 0 order by t.id`

// WHERE IN サブクエリ
var whereInSQL = `select * from users where id in (select user_id from orders where status = 'active') and role = 'admin'`

// EXISTS サブクエリ
var existsSQL = `select * from users u where exists (select 1 from orders o where o.user_id = u.id and o.status = 'active')`

// 複合: CASE + サブクエリ + BETWEEN
var complexSQL = `select u.id, case when u.role = 'admin' then 'Admin' else 'User' end as role_label, u.name from users u where u.created_at between '2020-01-01' and '2024-12-31' and u.id in (select user_id from orders where total > 100) order by u.id`

func advancedExample() {
	// 関数呼び出し内のカンマ保護
	funcSQL := `select coalesce(u.name, u.email, 'unknown'), count(distinct u.id) from users u where u.status = 'active' group by u.role`
	_ = funcSQL

	// 通常の文字列（変更されないこと）
	notSQL := `This contains SELECT but is not SQL`
	_ = notSQL
}
