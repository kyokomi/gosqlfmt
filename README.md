# gosqlfmt

Goソースファイル内のSQL文字列リテラルを自動フォーマットするCLIツール。
`gofmt` のように使える。

## インストール

```bash
go install github.com/kyokomi/gosqlfmt/cmd/gosqlfmt@latest
```

## 使い方

### 基本: フォーマット結果を標準出力に表示

```bash
gosqlfmt main.go
```

### ファイルを直接書き換え

```bash
gosqlfmt -w main.go
```

### 差分を表示

```bash
gosqlfmt -d main.go
```

### フォーマットが必要なファイル一覧を表示

```bash
gosqlfmt -l ./...
```

### ディレクトリ内の全 .go ファイルを再帰的にフォーマット

```bash
gosqlfmt -w ./
```

## フォーマット例

### 短いSQL: キーワード大文字化 + スペース正規化のみ

80文字以下のSQLは1行のまま、キーワード大文字化と無駄なスペースの除去のみ行う。

```go
// Before
q := `select * from hoge where id = ?`

// After
q := `SELECT * FROM hoge WHERE id = ?`
```

```go
// Before: 無駄なスペースも正規化
q := `SELECT  *  FROM   hoge   WHERE  id = ?`

// After
q := `SELECT * FROM hoge WHERE id = ?`
```

### 長いSQL: 複数行に展開

80文字を超えるSQLは主要句で改行し、body部分を2スペースインデントする。

#### SELECT

```go
// Before
var hogeSQL = `SELECT * FROM hoge WHERE id = ? AND status = 'active' ORDER BY created_at DESC LIMIT 1`

// After
var hogeSQL = `
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
```

#### SELECT + JOIN

```go
// Before
query := `select u.id, u.name, u.email from users u inner join orders o on u.id = o.user_id where u.status = ? and o.created_at > ? order by o.created_at desc limit 10 offset 20`

// After
query := `
SELECT
  u.id,
  u.name,
  u.email
FROM
  users u
INNER JOIN
  orders o
ON
  u.id = o.user_id
WHERE
  u.status = ?
  AND o.created_at > ?
ORDER BY
  o.created_at DESC
LIMIT
  10
OFFSET
  20
`
```

#### INSERT

```go
// Before
insertSQL := `insert into users (id, name, email, status) values (?, ?, ?, 'active')`

// After
insertSQL := `
INSERT INTO
  users (id, name, email, status)
VALUES
  (?, ?, ?, 'active')
`
```

#### UPDATE

```go
// Before
updateSQL := `UPDATE users SET name = ?, email = ?, updated_at = NOW() WHERE id = ? AND deleted_at IS NULL`

// After
updateSQL := `
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
```

## フォーマットルール

### 対象

- バッククォート（`` ` ``）で囲まれた文字列リテラルのみ
- `SELECT`, `INSERT`, `UPDATE`, `DELETE` で始まるSQLを自動検出

### キーワード

SQLキーワードは大文字に統一される:

```
SELECT, FROM, WHERE, JOIN, ON, AND, OR,
GROUP BY, ORDER BY, HAVING, LIMIT, OFFSET,
INSERT INTO, VALUES, UPDATE, SET, DELETE FROM,
INNER JOIN, LEFT JOIN, RIGHT JOIN, AS, IN, NOT,
NULL, IS, BETWEEN, LIKE, EXISTS, DISTINCT, ASC, DESC,
CASE, WHEN, THEN, ELSE, END, UNION, ALL
```

### 1行SQLの扱い

- キーワード大文字化と無駄なスペースの除去は常に行う
- 80文字を超えるSQLのみ複数行に展開する（将来的に閾値は設定で変更可能にする予定）

### 改行・インデント（複数行展開時）

- 主要句（`SELECT`, `FROM`, `WHERE`, `JOIN` 等）の前で改行
- 句のbody部分は2スペースインデント
- カラムリストは各カラムごとに改行

### プレースホルダ

以下のプレースホルダはそのまま保持される:

- `?` (MySQL)
- `:name` (名前付きプレースホルダ)
- `$1`, `$2` ... (PostgreSQL)

## オプション

| フラグ | 説明 |
|--------|------|
| `-w` | フォーマット結果でファイルを上書きする |
| `-d` | フォーマット前後のdiffを表示する |
| `-l` | フォーマットが必要なファイルパスを一覧表示する |
| (なし) | フォーマット結果を標準出力に出力する |

## CI/CDでの利用

`-l` フラグはフォーマットが必要なファイルがあれば終了コード1を返すので、CIでのチェックにそのまま使える:

```bash
# フォーマットが必要なファイルがあれば終了コード1で終了
gosqlfmt -l ./...
```

## 制限事項

- ダブルクォート文字列内のSQLは対象外（改行を含められないため）
- SQL構文の正しさは検証しない（フォーマットのみ）
- 初期バージョンはMySQL構文に最適化（PostgreSQL, SQLiteは将来対応予定）

## ライセンス

MIT
