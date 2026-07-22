---
name: worklog-operation
version: 1.0.0
description: worklog CLI をエージェントから操作するためのガイド
---

# worklog CLI 操作スキル

worklog は作業時間を記録するCLIツールです。このスキルを使ってエージェントから安全に操作できます。

## 重要な不変条件

- **`new` / `switch` は既存プロジェクトを自動停止する**。停止前に現在の状態を確認すること
- **タグIDは数値**。`worklog tag list --json` で取得してから使用すること
- **`tag delete` はインタラクティブ確認が必要**。`--no-interactive` では実行不可
- **全ミューテーション操作は `worklog status --json` で事前確認**してから実行すること
- **画像（スクリーンショット等）から読み取った時刻で `new` / `switch` / `stop`（`[HH:MM]` 引数）を実行する場合、実行前に必ず `AskUserQuestion` で読み取った開始/終了時刻をユーザーに確認すること**。時刻の読み取りミスによる事後訂正が実際に発生している
- **ユーザーが対話モードで `worklog stop` / `new` / `switch` を実行し `could not open a new TTY` 等のTTYエラーで失敗した場合**、エージェントが代わりに `--json` を付けて同じ操作を実行する（`stop`は必要なら `[HH:MM]` 引数も添える）。TTYが存在しない環境（Claude Codeのbash実行環境等）ではhuhライブラリの対話プロンプトが開けないため

## グローバルフラグ

| フラグ | 説明 |
|--------|------|
| `--json` | 構造化JSON出力（`--no-interactive` を暗黙的に有効化） |
| `--no-interactive` | TUIプロンプトを無効化。引数不足時はエラーを返す |

**エージェントは常に `--json` を使うこと。**

## コマンドパス

`worklog` コマンドはビルド済みでどこからでも使用可能。PATHに入っていない場合は `/Users/s23159/poc/worklog/worklog` をフルパスで使うこと。

## コマンド一覧

### 現在の状態確認

```sh
worklog status --json
```

出力例（稼働中）:
```json
{
  "status": {
    "project": "MyProject",
    "tag_id": "1",
    "tag_name": "開発",
    "start_time": "2026-03-26T10:00:00+09:00",
    "current_session_seconds": 3600.0,
    "total_seconds": 7200.0
  },
  "summaries": [...]
}
```

出力例（停止中）:
```json
{"status": null, "summaries": [...]}
```

### 本日の作業一覧

```sh
worklog list --json
```

NDJSON形式（1行1プロジェクト）:
```json
{"project":"ProjectA","tag_id":"1","tag_name":"開発","total_seconds":3600.0,"last_activity":"2026-03-26T11:00:00+09:00"}
{"project":"ProjectB","tag_id":"2","tag_name":"MTG","total_seconds":1800.0,"last_activity":"2026-03-26T12:00:00+09:00"}
```

指定日:
```sh
worklog list --json 2026-03-25   # YYYY-MM-DD
worklog list --json -1d          # 相対日付
```

### タイムライン表示

```sh
worklog timeline --json
worklog timeline --json 2026-03-25
```

NDJSON形式:
```json
{"project":"ProjectA","tag_id":"1","tag_name":"開発","total_seconds":3600.0,"time_ranges":[{"start":"2026-03-26T10:00:00+09:00","end":"2026-03-26T11:00:00+09:00","duration_seconds":3600.0}]}
```

### 新規プロジェクト開始

```sh
worklog new --json <プロジェクト名> <タグID> [HH:MM]
```

例:
```sh
worklog new --json "MyProject" 1
worklog new --json "MyProject" 1 09:30
```

出力:
```json
{"action":"new","project":"MyProject","tag_id":"1","tag_name":"開発","start_time":"2026-03-26T09:30:00+09:00"}
```

`prev_project` フィールドは自動停止した旧プロジェクト名（稼働中だった場合）。

### プロジェクト切り替え

```sh
worklog switch --json <プロジェクト名> <タグID> [HH:MM]
```

出力（`new` と同形式、`action: "switch"`）。

### 停止

```sh
worklog stop --json [HH:MM]
```

`--json` 時はTUIプロンプトが出ない。`[HH:MM]` 引数を渡せばその時刻で停止し、省略時は `time.Now()`（実行時点の時刻）で即停止する。

出力:
```json
{"action":"stop","project":"MyProject","tag_id":"1","tag_name":"開発","stop_time":"2026-03-26T18:00:00+09:00"}
```

### タグ管理

```sh
worklog tag list --json
```

```json
[{"id":1,"name":"開発"},{"id":2,"name":"MTG"}]
```

```sh
worklog tag add --json <タグ名>
```

```json
{"action":"added","id":3,"name":"新タグ"}
```

```sh
worklog tag delete <タグID>   # --json / --no-interactive では実行不可（確認必要）
```

## エラー出力形式

エラーは stdout に JSON で出力され、終了コードは 1:

```json
{"error":"NO_ACTIVE_PROJECT","message":"稼働中のプロジェクトがありません"}
```

| エラーコード | 説明 |
|---|---|
| `NO_ACTIVE_PROJECT` | `stop` 時に稼働中プロジェクトがない |
| `MISSING_ARGUMENTS` | 必須引数が不足（引数なし実行を含む） |
| `INVALID_PROJECT_NAME` | 不正なプロジェクト名（制御文字・`..`・256文字超等） |
| `INVALID_TAG_ID` | タグIDが数値でない |
| `INVALID_TAG_NAME` | 不正なタグ名（制御文字・256文字超等） |
| `INVALID_TIME_FORMAT` | 時刻形式が不正 |
| `INVALID_DATE_FORMAT` | 日付形式が不正 |
| `TAG_ALREADY_EXISTS` | タグ名が重複 |
| `TAG_NOT_FOUND` | 指定タグIDが存在しない |
| `INTERACTIVE_REQUIRED` | `--no-interactive` では実行不可な操作 |
| `UNKNOWN_COMMAND` | 不明なコマンド名 |
| `UNKNOWN_SUBCOMMAND` | `tag` の不明なサブコマンド名 |
| `INTERNAL_ERROR` | 内部エラー |

## エージェント向け推奨フロー

```
1. worklog status --json         # 現在状態を確認
2. worklog tag list --json       # 利用可能なタグIDを取得
3. worklog new --json <name> <tagID>  # or switch / stop
4. worklog status --json         # 操作結果を確認
```

## 入力制約

**プロジェクト名**: 1〜255文字（文字数単位、マルチバイト対応）
- 禁止文字: 制御文字（U+0000-U+001F, U+007F）、`..`、`/`、`\`、`*`、`?`、`"`、`<`、`>`、`|`

**タグ名**: 1〜255文字（文字数単位、マルチバイト対応）
- 禁止文字: 制御文字（U+0000-U+001F, U+007F）、`..`
- ファイルシステム文字（`/`, `\` 等）は許可

**時刻**: `HH:MM`（当日）

**日付**: `YYYY-MM-DD`、`YYYYMMDD`、`-Nd`（N日前）

**出力形式の使い分け**:
- 単一オブジェクト: `status`, `new`, `switch`, `stop`, `tag add`
- JSON配列: `tag list`（タグ数が少ないため）
- NDJSON（1行1オブジェクト）: `list`, `timeline`（件数が多くなりうるため、`jq` でストリーム処理可能）

## 過去プロジェクトの再スタート

ユーザーが「◯◯をやってたタスク」「◯◯的な名前のプロジェクト」等と曖昧に指示した場合、以下のフローで再開する。

### 1. キーワード検索

```sh
.claude/scripts/find-past-project.sh <keyword>
```

出力（タブ区切り、新しい順）:
```
2026-03-16	perman棚卸し	14
2026-02-16	perman棚卸し	14
```

カラム: `日付 \t プロジェクト名 \t 最終タグID`

`wl` のコマンドではなく、`~/.local/state/worklog/YYYY/MM/DD/<project>.log` を直接 find するスクリプト。`list` は単日しか見られないため、複数日を跨ぐ検索にはこちらを使う。

### 2. ヒットが複数ある場合

候補をそのままユーザーに提示して選ばせる。ヒット0件なら `worklog tag list --json` でタグ名との混同がないか確認した上で、別キーワードを提案する。

### 3. 再スタート

前回の最終タグIDを引き継いで開始:

```sh
worklog new --json "<project>" <last_tag_id>
```

タグIDは `find-past-project.sh` の3カラム目。ユーザーが別タグを指定した場合はそちらを優先。タグ名を人間可読に表示するなら `worklog tag list --json` で引いておく。

## 外部ソースからのプロジェクト開始

GitHub issueやNotionページなど外部ソースを起点に新規プロジェクトを開始する場合、**URLスラッグをそのままプロジェクト名にしない**。URLスラッグは実際のタイトルと異なる（省略・表記揺れ等）ことがある。

### 手順

1. **実際のタイトルを取得**してから命名する
   - GitHub issue: `gh issue view <番号> --repo <owner>/<repo> --json title,body,state`
   - Notionページ: `mcp__plugin_Notion_notion__notion-fetch` でページを取得し `title` を確認（URLのIDだけでは不十分。ページ内のスラッグ表記はタイトルの正確な表現とは限らない）
2. **命名形式とタグをユーザーに確認**してから `worklog new` を実行する
   - GitHub issue起点の既定形式: `#<issue番号> <issueタイトル>`（例: `#159 stg移設手順書の作成`）
   - 同一リポジトリのissueを続けて記録する場合、直前に使ったタグ・命名形式をデフォルト候補として提案してよい（ユーザーが明示的に別形式を指定したらそちらに従う）

### 命名を誤って開始してしまった場合

worklogにはプロジェクト名の変更コマンドが存在しない。時間が経過してから正しいタイトルが判明した場合は、ログファイルを直接リネームする：

```sh
mv "~/.local/state/worklog/YYYY/MM/DD/<誤った名前>.log" "~/.local/state/worklog/YYYY/MM/DD/<正しい名前>.log"
```

`current` ファイルは稼働中プロジェクトのみを指すため、対象日が稼働中でなければ修正不要。稼働中の場合は `current` ファイル内のプロジェクト名も同様に書き換える。
