// Package domain はアプリケーション全体で共有されるドメイン型を定義する
package domain

import "time"

// LogEntry はログの1エントリを表す
type LogEntry struct {
	Timestamp time.Time
	Action    string
	Tag       string
}

// Tag はプロジェクトのタグを表す
type Tag struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// TimeRange は開始と終了の時間範囲を表す
type TimeRange struct {
	Start    time.Time
	End      time.Time
	Duration time.Duration
}

// ProjectSummary は各プロジェクトの稼働時間サマリーを表す
type ProjectSummary struct {
	Project      string
	Tag          string // タグID
	TagName      string // タグ名
	TotalTime    time.Duration
	LastActivity time.Time
	TimeRanges   []TimeRange // 作業時間の範囲リスト
}

// ProjectStatus は現在稼働中のプロジェクトの状態を表す
type ProjectStatus struct {
	Project            string
	Tag                string // タグID
	TagName            string // タグ名
	StartTime          time.Time
	CurrentSessionTime time.Duration // 現在セッションの経過時間
	TotalTime          time.Duration // 本日の累計稼働時間
}
