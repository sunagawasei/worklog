# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## このプロジェクトの役割

仕事のスケジュール、優先順位、進捗、担当先への引き継ぎを管理する。
予定に載せた仕事の実作業は、担当プロジェクトの別workspaceへ `handoff` で渡す。
このプロジェクトでは計画更新、依頼の送信、着弾確認を行い、ニュース収集、コンテンツ更新、実装、環境への反映などの実作業は行わない。
委譲後の進行を追うかどうかは `handoff` の指定に従う。
外部送信を含むhandoffには「下書きまで作成し、宛先と正確な本文を本人に提示して承認後に送信する」と明記する。委譲元の送信指示だけを本人の承認として扱わない。
稼働時間の記録は、このプロジェクトの管理業務に含める。
本人が実作業の開始や `handoff` による着手を指示したら、このセッションで `worklog` skillに従って記録を開始し、状態を確認する。
開始時刻は指示を受けた時刻のJSTを渡し、同一タスクの既存名と最終タグを再利用する。委譲先には二重に記録させない。
計画の相談や候補の提示だけでは計測を開始しない。停止は本人が明示したときだけ行い、個別作業の完了報告だけでプロジェクト全体の計測を止めない。

## 仕事のタスクとスケジュールの管理

仕事の棚卸し、週間計画、予定を考慮した日程調整には [work-planning](.claude/skills/work-planning/SKILL.md) を使う。
Codex には `.agents/skills` から同じ原本を共有する。
勤務条件と参照先は `CLAUDE.local.md`、計画はリポジトリ直下の `work-planning/開始日_終了日/` に保存する(週は月曜〜日曜、日付は `YYYY-MM-DD`。いずれも Git 管理外)。
worklog は作業実績の記録に使い、将来の予定やタスクの完了状態とは分けて扱う。
週全体は週ディレクトリの `todo.md`、当日分はその下の `days/YYYY-MM-DD/todo.md` を見る。それぞれの詳細は同じ階層の `plan.md` に置き、計画や進捗の変更を対応するファイルに反映する。
当日の時間割は日付ディレクトリの `schedule.md` に、短い一文と「時間 / 内容」の表だけで表示する。

## 開発コマンド

```bash
# テスト（コードの正しさ確認）
go test ./...
go test -run TestCurrentStorage_Read ./internal/storage/  # 単一テスト

# 品質チェック
go vet ./...
go fmt ./...

# バイナリビルド（実際のCLI動作確認に必要）
go build -o worklog .
```

## アーキテクチャ

### 5層 + config構成

```
main.go                  # DIコンテナ（依存関係の組み立てのみ）
internal/
├── config/              # データ/タグファイルのパス解決（XDG準拠）
├── domain/              # 共有型定義のみ（Tag, LogEntry, TimeRange, ProjectStatus, ProjectSummary）
├── storage/             # ファイルI/O層
├── project/             # ビジネスロジック層
├── cmd/                 # CLIコマンド層
└── ui/                  # 表示層（render_*.go + prompt.go）
```

### 依存の向き

```
main.go → storage/project/cmd を生成してDI
cmd → project.ProjectManager（interface経由）, ui
project → storage.{Current,Log,Tag}Storage（interface経由）, domain
storage → domain
ui → domain
```

`main.go` が唯一の組み立て場所。各層は interface を通じて疎結合。

### バリデーション

`internal/storage/validation.go` に入力バリデーション関数を定義。cmd 層から呼ぶ：

```go
storage.ValidateProjectName(name) error  // パストラバーサル対策・長さ制限含む
storage.ValidateTagName(name) error
```

`storage.LogEntry` は `domain.LogEntry` の型エイリアス（`=` で定義）。時刻は全て `At` 版に一本化。`NewAt/SwitchAt/StopAt` + `time.Now()` を cmd 層で呼ぶ（`New/Switch/Stop` は存在しない）。

### テスト戦略

`manager_test.go` が典型例。storage の3 interface にモックを注入して project 層を単体テスト。`t.Run()` サブテスト形式を使用。テーブルテストは不使用。

cmd 層のテストは `ExecOptions.Writer` に `bytes.Buffer` を注入して出力をキャプチャする（`json_test.go` / `timeline_test.go` 参照）。

`main_test.go` はルートの `run()` 関数をスモークテストする統合テスト（`os.Args` を直接書き換えてテスト）。

## データフォーマット

**current ファイル**（`~/.local/state/worklog/YYYY/MM/DD/current`）:

```
ProjectName\tTagID
```

タブ区切り。`parseCurrent()` ヘルパー（`internal/project/manager.go`）で解析。

**ログファイル**（同ディレクトリの `ProjectName.log`）:

```
2025/09/22 10:00:00\tstart\t1
2025/09/22 10:50:00\tstop\t1
```

アクションは `start` / `stop` のみ。`calculateSessions()` ヘルパーでセッション計算。

**tags.json**（`~/.config/worklog/tags.json`）:

```json
{ "tags": [{ "id": 1, "name": "開発" }], "nextID": 2 }
```

パス優先順: `WORKLOG_TAGS_FILE` 環境変数 > `XDG_CONFIG_HOME/worklog/tags.json` > `~/.config/worklog/tags.json`

## UI層の構成

`internal/ui/display.go` が全 render\_\*.go から参照される共有定数・ユーティリティ（ボックス描画文字、`FormatDuration`、`GetTerminalWidth` 等）。レスポンシブレイアウト定数（`minContentWidth=44`〜`maxContentWidth=80`）とダッシュボードの2カラム分割もここで制御。

カラー出力は使用しない。外部依存は `charmbracelet/huh`, `go-runewidth`, `golang.org/x/term` のみ。

## コマンド追加パターン

新しいコマンドを追加する手順：

1. `internal/cmd/<command>.go` にハンドラ関数を作成（`handle<Command>(manager, opts ExecOptions) error`）
2. `internal/cmd/root.go` の `switch` 文に `case "<command>"` を追加
3. 表示が必要なら `internal/ui/render_<command>.go` を追加し、`display.go` の共有ユーティリティを使用
4. JSON 出力が必要なら `internal/cmd/json_output.go` に変換関数を追加
5. `internal/ui/render_misc.go` の `RenderHelp` にコマンドの説明を追記

## グローバルフラグと JSON モード

```
worklog --json <command>          # JSON/NDJSON 出力（自動的に --no-interactive も有効）
worklog --no-interactive <command> # 対話 UI を無効化（引数が不足するとエラー）
```

グローバルフラグのパースは `internal/cmd/options.go` の `ParseGlobalFlags(args)` が担う。結果は `ExecOptions{JSONMode, NoInteractive, Args}` に格納され、各コマンドに渡される。

`--json` 時の出力形式（`json_output.go` に定義）:

- `status`: `{"status": {...}, "summaries": [...]}` （単一 JSON）
- `list` / `timeline`: 1行1件の NDJSON
- `new` / `switch` / `stop`: `{"action": "...", "project": "...", ...}`
- エラー: `{"error": "ERROR_CODE", "message": "..."}`

**`HandledError` パターン**: cmd 層が JSON エラーを既に出力した場合、`HandledError{}` を返す。`main.go` はこの型を受け取っても stderr に出力しない。
