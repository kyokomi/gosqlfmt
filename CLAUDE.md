# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## プロジェクト概要

GoソースファイルのバッククォートSQL文字列リテラルを自動フォーマットするCLIツール。`gofmt`のように使える。
キーワード大文字化・スペース正規化を行い、80文字を超えるSQLは主要句ごとに改行+2スペースインデントで複数行に展開する。

## コマンド

```bash
# テスト全体実行
go test ./...

# 特定パッケージのテスト
go test ./sqlfmt/
go test ./formatter/

# 特定テストケース実行
go test ./sqlfmt/ -run TestFormat
go test ./formatter/ -run TestFormatSource

# ビルド（CLIはcmd/gosqlfmt/に配置予定だが、main.goは未作成）
go build ./...
```

## アーキテクチャ

### パッケージ構成

- **`sqlfmt/`** — SQLフォーマットのコアロジック（Go ASTに依存しない純粋なSQL処理）
  - `token.go`: トークン型定義、キーワード一覧（`keywords`）、主要句キーワード（`clauseKeywords`）、複合キーワードペア（`compoundPairs`）
  - `lexer.go`: SQL字句解析器。プレースホルダ（`?`, `:name`, `$1`）、ドット付き識別子（`table.column`）、演算子、文字列リテラルを処理
  - `sqlfmt.go`: フォーマット本体。`Normalize()`→1行正規化、`Format()`→閾値判定+複数行展開。内部で句分割→句種別ごとの整形（SELECT/SETはカンマ区切り、WHERE/ON/HAVINGはAND/OR区切り）

- **`formatter/`** — Go ASTを使ってGoソースファイル内のバッククォートSQL文字列を検出・置換
  - `FormatSource()`: `go/parser`でAST解析→バッククォートの`ast.BasicLit`を走査→SQL判定→`sqlfmt.Format()`で変換→後方から文字列置換

- **`cmd/gosqlfmt/`** — CLIエントリポイント（未実装）
- **`testdata/`** — `input/`（フォーマット前）と`golden/`（期待結果）のペアでformatterをテスト

### 処理の流れ

```
Goソース → formatter.FormatSource() → go/parser でAST解析
  → バッククォート文字列を検出 → isSQL() でSQL判定
  → sqlfmt.Format() でフォーマット
    → Normalize(): Lexer.Tokenize() → キーワード大文字化 + スペース正規化
    → 80文字超なら formatMultiLine(): 複合キーワード結合 → 句分割 → 句種別ごとに整形
  → 後方から文字列置換して返却
```

### フォーマットルール

- 80文字以下: キーワード大文字化+スペース正規化のみ（1行維持）
- 80文字超: 主要句（SELECT/FROM/WHERE/JOIN等）で改行、body部分を2スペースインデント
- SELECT/SET句: カンマごとに改行
- WHERE/ON/HAVING句: AND/ORごとに改行
- 括弧内のカンマやAND/ORでは分割しない（`parenDepth`で管理）
