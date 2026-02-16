package sqlfmt_test

import (
	"testing"

	"github.com/kyokomi/gosqlfmt/sqlfmt"
)

func TestNormalize(t *testing.T) {
	tests := []struct {
		name string
		sql  string
		want string
	}{
		{
			name: "キーワード小文字を大文字に変換",
			sql:  "select * from hoge where id = ?",
			want: "SELECT * FROM hoge WHERE id = ?",
		},
		{
			name: "キーワード混在を大文字に統一",
			sql:  "Select id From users Where status = ?",
			want: "SELECT id FROM users WHERE status = ?",
		},
		{
			name: "連続スペースを正規化",
			sql:  "SELECT  *  FROM   hoge   WHERE  id = ?",
			want: "SELECT * FROM hoge WHERE id = ?",
		},
		{
			name: "先頭と末尾の空白を除去",
			sql:  "  SELECT * FROM hoge  ",
			want: "SELECT * FROM hoge",
		},
		{
			name: "JOINキーワードの大文字化",
			sql:  "select * from users inner join orders on users.id = orders.user_id",
			want: "SELECT * FROM users INNER JOIN orders ON users.id = orders.user_id",
		},
		{
			name: "AND/ORの大文字化",
			sql:  "select * from hoge where id = ? and status = ? or name = ?",
			want: "SELECT * FROM hoge WHERE id = ? AND status = ? OR name = ?",
		},
		{
			name: "ORDER BY / GROUP BYの大文字化",
			sql:  "select * from hoge group by id order by name asc",
			want: "SELECT * FROM hoge GROUP BY id ORDER BY name ASC",
		},
		{
			name: "INSERT文のキーワード大文字化",
			sql:  "insert into users (id, name) values (?, ?)",
			want: "INSERT INTO users (id, name) VALUES (?, ?)",
		},
		{
			name: "UPDATE文のキーワード大文字化",
			sql:  "update users set name = ? where id = ?",
			want: "UPDATE users SET name = ? WHERE id = ?",
		},
		{
			name: "DELETE文のキーワード大文字化",
			sql:  "delete from users where id = ?",
			want: "DELETE FROM users WHERE id = ?",
		},
		{
			name: "LIMITとOFFSETの大文字化",
			sql:  "select * from hoge limit 10 offset 5",
			want: "SELECT * FROM hoge LIMIT 10 OFFSET 5",
		},
		{
			name: "DISTINCTの大文字化",
			sql:  "select distinct name from users",
			want: "SELECT DISTINCT name FROM users",
		},
		{
			name: "BETWEEN / IS NULL / LIKE の大文字化",
			sql:  "select * from hoge where id between 1 and 10 and name is null or name like '%test%'",
			want: "SELECT * FROM hoge WHERE id BETWEEN 1 AND 10 AND name IS NULL OR name LIKE '%test%'",
		},
		{
			name: "LEFT JOIN / RIGHT JOINの大文字化",
			sql:  "select * from a left join b on a.id = b.id right join c on b.id = c.id",
			want: "SELECT * FROM a LEFT JOIN b ON a.id = b.id RIGHT JOIN c ON b.id = c.id",
		},
		{
			name: "プレースホルダ ? をそのまま保持",
			sql:  "select * from hoge where id = ?",
			want: "SELECT * FROM hoge WHERE id = ?",
		},
		{
			name: "名前付きプレースホルダ :name をそのまま保持",
			sql:  "select * from hoge where id = :id",
			want: "SELECT * FROM hoge WHERE id = :id",
		},
		{
			name: "PostgreSQLプレースホルダ $1 をそのまま保持",
			sql:  "select * from hoge where id = $1 and name = $2",
			want: "SELECT * FROM hoge WHERE id = $1 AND name = $2",
		},
		{
			name: "既にフォーマット済みのSQLは変化しない",
			sql:  "SELECT * FROM hoge WHERE id = ?",
			want: "SELECT * FROM hoge WHERE id = ?",
		},
		{
			name: "HAVINGの大文字化",
			sql:  "select count(*) from users group by status having count(*) > 1",
			want: "SELECT count(*) FROM users GROUP BY status HAVING count(*) > 1",
		},
		{
			name: "UNION ALLの大文字化",
			sql:  "select id from a union all select id from b",
			want: "SELECT id FROM a UNION ALL SELECT id FROM b",
		},
		{
			name: "CASE WHEN の大文字化",
			sql:  "select case when status = 1 then 'active' else 'inactive' end from users",
			want: "SELECT CASE WHEN status = 1 THEN 'active' ELSE 'inactive' END FROM users",
		},
		{
			name: "DESC/ASCの大文字化",
			sql:  "select * from hoge order by id desc, name asc",
			want: "SELECT * FROM hoge ORDER BY id DESC, name ASC",
		},
		{
			name: "IN句の大文字化",
			sql:  "select * from hoge where id in (1, 2, 3)",
			want: "SELECT * FROM hoge WHERE id IN (1, 2, 3)",
		},
		{
			name: "NOT / EXISTSの大文字化",
			sql:  "select * from hoge where not exists (select 1 from fuga where fuga.id = hoge.id)",
			want: "SELECT * FROM hoge WHERE NOT EXISTS (SELECT 1 FROM fuga WHERE fuga.id = hoge.id)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sqlfmt.Normalize(tt.sql)
			if got != tt.want {
				t.Errorf("Normalize(%q)\n  got:  %q\n  want: %q", tt.sql, got, tt.want)
			}
		})
	}
}

func TestFormat(t *testing.T) {
	tests := []struct {
		name string
		sql  string
		want string
	}{
		{
			name: "短いSELECTは1行のまま（キーワード大文字化のみ）",
			sql:  "select * from hoge where id = ?",
			want: "SELECT * FROM hoge WHERE id = ?",
		},
		{
			name: "長いSELECTは複数行に展開",
			sql:  "SELECT * FROM hoge WHERE id = ? AND status = 'active' ORDER BY created_at DESC LIMIT 1",
			want: "\nSELECT\n  *\nFROM\n  hoge\nWHERE\n  id = ?\n  AND status = 'active'\nORDER BY\n  created_at DESC\nLIMIT\n  1\n",
		},
		{
			name: "SELECT + カラムリスト（カンマ区切りで改行）",
			sql:  "select u.id, u.name, u.email from users u where u.status = ? order by u.created_at desc limit 10",
			want: "\nSELECT\n  u.id,\n  u.name,\n  u.email\nFROM\n  users u\nWHERE\n  u.status = ?\nORDER BY\n  u.created_at DESC\nLIMIT\n  10\n",
		},
		{
			name: "INNER JOIN",
			sql:  "select u.id, u.name from users u inner join orders o on u.id = o.user_id where u.status = ? limit 10",
			want: "\nSELECT\n  u.id,\n  u.name\nFROM\n  users u\nINNER JOIN\n  orders o\nON\n  u.id = o.user_id\nWHERE\n  u.status = ?\nLIMIT\n  10\n",
		},
		{
			name: "LEFT JOIN",
			sql:  "select a.id, b.name from table_a a left join table_b b on a.id = b.a_id where a.status = 'active' order by a.id",
			want: "\nSELECT\n  a.id,\n  b.name\nFROM\n  table_a a\nLEFT JOIN\n  table_b b\nON\n  a.id = b.a_id\nWHERE\n  a.status = 'active'\nORDER BY\n  a.id\n",
		},
		{
			name: "INSERT INTO",
			sql:  "insert into users (id, name, email, status) values (?, ?, ?, 'active')",
			want: "\nINSERT INTO\n  users (id, name, email, status)\nVALUES\n  (?, ?, ?, 'active')\n",
		},
		{
			name: "UPDATE + SET + WHERE",
			sql:  "UPDATE users SET name = ?, email = ?, updated_at = NOW() WHERE id = ? AND deleted_at IS NULL",
			want: "\nUPDATE\n  users\nSET\n  name = ?,\n  email = ?,\n  updated_at = NOW()\nWHERE\n  id = ?\n  AND deleted_at IS NULL\n",
		},
		{
			name: "DELETE FROM",
			sql:  "delete from users where id = ? and status = 'inactive' and created_at < '2020-01-01'",
			want: "\nDELETE FROM\n  users\nWHERE\n  id = ?\n  AND status = 'inactive'\n  AND created_at < '2020-01-01'\n",
		},
		{
			name: "GROUP BY + HAVING",
			sql:  "select status, count(*) as cnt from users group by status having count(*) > 1 order by cnt desc",
			want: "\nSELECT\n  status,\n  count(*) AS cnt\nFROM\n  users\nGROUP BY\n  status\nHAVING\n  count(*) > 1\nORDER BY\n  cnt DESC\n",
		},
		{
			name: "LIMIT + OFFSET",
			sql:  "select id, name, email from users where status = 'active' order by created_at desc limit 10 offset 20",
			want: "\nSELECT\n  id,\n  name,\n  email\nFROM\n  users\nWHERE\n  status = 'active'\nORDER BY\n  created_at DESC\nLIMIT\n  10\nOFFSET\n  20\n",
		},
		{
			name: "WHERE句の複数条件（AND/OR）",
			sql:  "select * from users where status = ? and role = 'admin' and created_at > ? or name like '%test%' limit 10",
			want: "\nSELECT\n  *\nFROM\n  users\nWHERE\n  status = ?\n  AND role = 'admin'\n  AND created_at > ?\n  OR name LIKE '%test%'\nLIMIT\n  10\n",
		},
		{
			name: "名前付きプレースホルダ",
			sql:  "select id, name from fuga where id = :id and status = :status order by created_at desc limit 30",
			want: "\nSELECT\n  id,\n  name\nFROM\n  fuga\nWHERE\n  id = :id\n  AND status = :status\nORDER BY\n  created_at DESC\nLIMIT\n  30\n",
		},
		{
			name: "PostgreSQLプレースホルダ",
			sql:  "select id, name from users where id = $1 and status = $2 order by created_at desc limit $3 offset $4",
			want: "\nSELECT\n  id,\n  name\nFROM\n  users\nWHERE\n  id = $1\n  AND status = $2\nORDER BY\n  created_at DESC\nLIMIT\n  $3\nOFFSET\n  $4\n",
		},
		{
			name: "無駄なスペースを含む長いSQL",
			sql:  "SELECT  id,  name,  email  FROM   users   WHERE  status = ?  AND  role = 'admin'  ORDER BY  id  DESC",
			want: "\nSELECT\n  id,\n  name,\n  email\nFROM\n  users\nWHERE\n  status = ?\n  AND role = 'admin'\nORDER BY\n  id DESC\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sqlfmt.Format(tt.sql)
			if got != tt.want {
				t.Errorf("Format(%q)\n  got:  %q\n  want: %q", tt.sql, got, tt.want)
			}
		})
	}
}

func TestFormatWithWidth(t *testing.T) {
	tests := []struct {
		name      string
		sql       string
		lineWidth int
		want      string
	}{
		{
			name:      "閾値40で短いSQLは展開されない",
			sql:       "select * from hoge where id = ?",
			lineWidth: 40,
			want:      "SELECT * FROM hoge WHERE id = ?",
		},
		{
			name:      "閾値30で同じSQLは展開される",
			sql:       "select * from hoge where id = ?",
			lineWidth: 30,
			want:      "\nSELECT\n  *\nFROM\n  hoge\nWHERE\n  id = ?\n",
		},
		{
			name:      "閾値0は常に展開する",
			sql:       "select * from hoge",
			lineWidth: 0,
			want:      "\nSELECT\n  *\nFROM\n  hoge\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sqlfmt.FormatWithWidth(tt.sql, tt.lineWidth)
			if got != tt.want {
				t.Errorf("FormatWithWidth(%q, %d)\n  got:  %q\n  want: %q", tt.sql, tt.lineWidth, got, tt.want)
			}
		})
	}
}
