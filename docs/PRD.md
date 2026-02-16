# gosqlfmt - Product Requirements Document

## 概要

`gosqlfmt` は、Goソースファイル内に記述されたSQL文字列リテラルを自動フォーマットするCLIツール。
`gofmt` のようなインターフェースで、SQLの可読性を統一的に向上させる。

## 背景・モチベーション

Goのコードベースでは、SQLクエリを文字列リテラル（主にバッククォート）として直接記述することが多い。
しかし、以下の問題が発生しやすい：

- 開発者ごとにSQLの書き方がバラバラ（改行位置、インデント、キーワードの大文字/小文字）
- 長いSQLが1行で書かれて可読性が低い
- コードレビュー時にSQLの構造が把握しにくい

`gosqlfmt` はこれらの問題を自動フォーマットで解決する。

## ターゲットユーザー

- GoでSQLを直接記述するバックエンドエンジニア
- コードの一貫性を重視するチーム
- CI/CDパイプラインにフォーマットチェックを組み込みたいチーム

## 機能要件

### 1. SQL文字列リテラルの検出

- Goの `go/ast` パッケージを使用してソースファイルを解析する
- バッククォート（`` ` ``）で囲まれた文字列リテラルを抽出する
- SQL文かどうかを判定する（`SELECT`, `INSERT`, `UPDATE`, `DELETE`, `CREATE` 等のキーワードで始まるか）
- ダブルクォート文字列は改行を含められないため対象外とする

### 2. SQLフォーマットルール

#### キーワードの大文字化

以下のキーワードを大文字に統一する：

```
SELECT, FROM, WHERE, JOIN, LEFT JOIN, RIGHT JOIN, INNER JOIN,
OUTER JOIN, CROSS JOIN, ON, AND, OR, GROUP BY, ORDER BY, HAVING,
LIMIT, OFFSET, INSERT, INTO, VALUES, UPDATE, SET, DELETE,
AS, IN, NOT, NULL, IS, BETWEEN, LIKE, EXISTS,
CASE, WHEN, THEN, ELSE, END,
UNION, ALL, DISTINCT, ASC, DESC
```

#### 改行ルール

主要句の前で改行する：

- `SELECT`
- `FROM`
- `WHERE`
- `JOIN`（LEFT/RIGHT/INNER/OUTER/CROSS含む）
- `GROUP BY`
- `ORDER BY`
- `HAVING`
- `LIMIT`
- `OFFSET`
- `INSERT INTO`
- `VALUES`
- `UPDATE`
- `SET`（UPDATE文内）
- `DELETE FROM`
- `UNION` / `UNION ALL`

#### インデントルール

- 主要句のbody部分を2スペースでインデントする
- カラムリストは各カラムごとに改行してインデントする
- `AND` / `OR` は所属する句の配下としてインデントする

#### 1行SQLの扱い

SQLの長さに関わらず、以下のフォーマットは常に適用する：

- **キーワードの大文字化**: `select` → `SELECT`, `from` → `FROM` 等
- **無駄なスペースの正規化**: 連続するスペースを1つにまとめる

複数行への展開は、フォーマット後のSQLが一定の文字数（デフォルト: 80文字）を超える場合に行う。
閾値未満の短いSQLは1行のまま保持する。

```
例: 閾値80文字

// 短いSQL（1行のまま、キーワード大文字化のみ）
`select * from hoge where id = ?`
→ `SELECT * FROM hoge WHERE id = ?`

// 長いSQL（複数行に展開）
`select id, name, email from users where status = ? and role = ? order by created_at desc limit 10`
→ 複数行にフォーマット
```

### 3. CLIインターフェース

```
gosqlfmt [flags] [path ...]
```

#### フラグ

| フラグ | 説明 |
|--------|------|
| `-w` | フォーマット結果でファイルを上書きする |
| `-d` | フォーマット前後のdiffを表示する |
| `-l` | フォーマットが必要なファイルのパスを一覧表示する |
| (なし) | フォーマット結果を標準出力に出力する |

#### パス指定

- ファイルパスを指定: そのファイルをフォーマット
- ディレクトリパスを指定: 配下の `.go` ファイルを再帰的にフォーマット
- パス未指定: 標準入力から読み取り（パイプ対応）

### 4. 出力

- デフォルト: フォーマット結果を標準出力に出力
- `-w`: 元ファイルを直接書き換え
- `-d`: unified diff形式で差分を表示
- `-l`: フォーマットが必要なファイルパスを改行区切りで出力

## 非機能要件

- フォーマット後のGoファイルは `gofmt` の出力と整合性が取れること
- 元のGoコードの構造（変数宣言、関数呼び出し等）を壊さないこと
- プレースホルダ（`?`, `:name`, `$1` 等）をそのまま保持すること

## アーキテクチャ

### パッケージ構成

```
gosqlfmt/
├── cmd/gosqlfmt/       # CLIエントリーポイント
│   └── main.go
├── formatter/           # Goファイルのフォーマット処理
│   ├── formatter.go     # メインのフォーマットロジック
│   └── formatter_test.go
├── sqlfmt/              # SQLフォーマットエンジン
│   ├── sqlfmt.go        # SQLフォーマットのインターフェース
│   ├── sqlfmt_test.go
│   ├── lexer.go         # SQLの字句解析
│   ├── lexer_test.go
│   ├── token.go         # トークン定義
│   └── dialect/         # 将来的なダイアレクト対応
│       └── mysql.go
├── docs/
│   └── PRD.md
├── testdata/            # テスト用のGoファイル
│   ├── input/
│   └── golden/
├── go.mod
└── README.md
```

### 処理フロー

```
Goファイル読み込み
  ↓
go/ast でパース
  ↓
バッククォート文字列リテラルを走査
  ↓
SQL判定（キーワードベース）
  ↓
SQLフォーマット（字句解析 → トークン列 → 再構築）
  ↓
フォーマット後のSQLで文字列リテラルを置換
  ↓
出力（標準出力 / ファイル上書き / diff）
```

## 将来的な拡張

### Phase 2: マルチダイアレクト対応
- PostgreSQL固有の構文（`::` キャスト、`RETURNING` 句など）
- SQLite固有の構文
- `--dialect` フラグの追加

### Phase 3: 高度なフォーマット
- サブクエリの対応とネストされたインデント
- `CASE WHEN` のフォーマット
- コメント（`--`, `/* */`）の保持
- カラムのアラインメント（カンマ位置の揃え）

### Phase 4: エディタ統合
- VSCode拡張
- GoLandプラグイン
- `goimports` のようなセーブ時自動実行

### Phase 5: 設定ファイル対応
- `.gosqlfmt.yaml` による設定
- インデント幅のカスタマイズ
- キーワードの大文字/小文字の選択
- プロジェクト単位の設定

## 対象外（初期スコープ）

- SQL構文の正しさの検証（バリデーション）
- SQLの実行やDB接続
- ORM生成コードのフォーマット
- `//go:embed` で読み込まれた `.sql` ファイルのフォーマット
- ダブルクォート文字列内のSQL
