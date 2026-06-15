# Worklog CLI

プロジェクトごとの稼働時間を記録・管理するCLIツール

## 特徴

- **シンプルな操作** - 直感的なコマンドで作業時間を記録
- **タグベース管理** - プロジェクトをタグで分類
- **対話モード** - 引数なしで対話的に操作可能（charmbracelet/huh 使用）
- **時刻指定機能** - 任意の時刻で打刻可能（過去の記録も可能）
- **時間集計** - 日次の作業時間を自動計算・時間範囲表示
- **プロジェクト切り替え** - 複数プロジェクトをスムーズに切り替え
- **ダッシュボード表示** - 現在の稼働状態と本日の集計を2カラムで同時表示
- **タイムライン可視化** - 9:00-20:00の作業状況を視覚的に表示

## インストール

### ビルドから実行

```bash
# ビルド
go build -o worklog .

# 実行
./worklog help
```

## 使い方

### 基本コマンド

| コマンド   | 説明                             | 例                         |
| ---------- | -------------------------------- | -------------------------- |
| `new`      | 新規プロジェクト開始             | `worklog new "開発作業" 1` |
| `switch`   | プロジェクト切り替え             | `worklog switch "会議" 2`  |
| `status`   | 現在の状況確認（ダッシュボード） | `worklog status`           |
| `stop`     | プロジェクト停止                 | `worklog stop`             |
| `list`     | 作業一覧（日付指定可能）         | `worklog list [-1d]`       |
| `timeline` | タイムライン表示（日付指定可能） | `worklog timeline [-1d]`   |
| `tag`      | タグ管理（一覧/追加/削除）       | `worklog tag list`         |
| `help`     | ヘルプ表示                       | `worklog help`             |

**タグID**: `tags.json`に定義された数値ID（1, 2, 3...）を使用します

### 対話モード

引数を省略すると対話モードが起動します：

```bash
# 対話的に新規プロジェクトを開始
worklog new

# 対話的にプロジェクトを切り替え
worklog switch
```

### 1日の作業フロー

```bash
# 朝、開発作業を開始
worklog new "機能実装" 1

# 会議に切り替え
worklog switch "定例会議" 2

# 会議終了、開発に戻る
worklog switch "機能実装" 1

# 昼休み
worklog stop

# 午後、作業再開（時刻指定も可能）
worklog new "機能実装" 1 13:00

# 本日の作業時間を確認
worklog list

# タイムラインで作業状況を可視化
worklog timeline

# 作業終了
worklog stop
```

### ダッシュボードとタイムライン

```bash
# 現在の稼働状況と本日のサマリーを2カラムで表示
worklog status

# 9:00-20:00のタイムラインで作業状況を可視化
# プロジェクトごとに異なる記号（■, ▦, ▨, ▥, ▧）で表示
# idle時間は · で表示
worklog timeline

# 過去の記録を確認
worklog timeline -1d           # 昨日のタイムライン
worklog timeline 2025-01-15    # 特定日のタイムライン
worklog timeline 20250115      # 連続形式も対応
```

## データ管理

作業ログは日付ごとにディレクトリで管理されます（XDG Base Directory準拠）：

```
~/.local/state/worklog/
├── 2025/
│   └── 09/
│       ├── 25/
│       │   ├── ProjectA.log
│       │   ├── ProjectB.log
│       │   └── current
│       └── 26/
│           └── ...
└── 2026/
    └── ...
```

**環境変数でカスタマイズ可能：**
- `WORKLOG_DATA_DIR` - データディレクトリを直接指定
- `XDG_STATE_HOME` - XDG準拠のベースディレクトリを指定

## 設定

### タグのカスタマイズ

`~/.config/worklog/tags.json` を編集してタグを追加・変更できます：

```json
{
    "tags": [
        {
            "id": 1,
            "name": "カスタムタグ"
        },
        {
            "id": 2,
            "name": "別のタグ"
        }
    ]
}
```

**タグファイルのパス（優先順位）：**
1. `WORKLOG_TAGS_FILE` 環境変数で指定したパス
2. `$XDG_CONFIG_HOME/worklog/tags.json`（デフォルト: `~/.config/worklog/tags.json`）

### アーキテクチャ

```
internal/
├── config/     # データ/タグファイルのパス解決（XDG準拠）
├── domain/     # 共有型定義
├── storage/    # ファイルI/O層
├── project/    # ビジネスロジック層
├── cmd/        # CLIコマンド層
└── ui/         # 表示層
    ├── display.go  # 共有定数・ユーティリティ（ボックス描画、プログレスバー）
    └── prompt.go   # 対話的UI（charmbracelet/huh）
```

