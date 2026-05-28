package cmd

import (
	"time"

	"worklog/internal/domain"
)

// --- エラー出力 ---

// errorJSON は JSON モードのエラー出力
type errorJSON struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

// --- status コマンド ---

// statusJSON は status コマンドの JSON 出力
type statusJSON struct {
	Project            string  `json:"project"`
	TagID              string  `json:"tag_id"`
	TagName            string  `json:"tag_name"`
	StartTime          string  `json:"start_time"`              // RFC3339
	CurrentSessionSecs float64 `json:"current_session_seconds"` // ナノ秒ではなく秒数
	TotalSecs          float64 `json:"total_seconds"`
}

// summaryStatusJSON は status と共に返す本日の集計情報
type summaryStatusJSON struct {
	Status    *statusJSON    `json:"status"`    // nil = 稼働中プロジェクトなし
	Summaries []listItemJSON `json:"summaries"` // 本日の全プロジェクト
}

func toStatusJSON(status *domain.ProjectStatus) statusJSON {
	return statusJSON{
		Project:            status.Project,
		TagID:              status.Tag,
		TagName:            status.TagName,
		StartTime:          status.StartTime.Format(time.RFC3339),
		CurrentSessionSecs: status.CurrentSessionTime.Seconds(),
		TotalSecs:          status.TotalTime.Seconds(),
	}
}

// --- list コマンド ---

// listItemJSON は list コマンドの NDJSON 1行分
type listItemJSON struct {
	Project      string  `json:"project"`
	TagID        string  `json:"tag_id"`
	TagName      string  `json:"tag_name"`
	TotalSecs    float64 `json:"total_seconds"`
	LastActivity string  `json:"last_activity"` // RFC3339
}

func toListItemJSON(s domain.ProjectSummary) listItemJSON {
	return listItemJSON{
		Project:      s.Project,
		TagID:        s.Tag,
		TagName:      s.TagName,
		TotalSecs:    s.TotalTime.Seconds(),
		LastActivity: s.LastActivity.Format(time.RFC3339),
	}
}

// --- timeline コマンド ---

// timeRangeJSON はタイムライン用の作業時間範囲
type timeRangeJSON struct {
	Start   string  `json:"start"`         // RFC3339
	End     string  `json:"end,omitempty"` // RFC3339（セッション継続中は空）
	DurSecs float64 `json:"duration_seconds"`
}

// timelineItemJSON は timeline コマンドの NDJSON 1行分
type timelineItemJSON struct {
	Project    string          `json:"project"`
	TagID      string          `json:"tag_id"`
	TagName    string          `json:"tag_name"`
	TotalSecs  float64         `json:"total_seconds"`
	TimeRanges []timeRangeJSON `json:"time_ranges"`
}

func toTimelineItemJSON(s domain.ProjectSummary) timelineItemJSON {
	ranges := make([]timeRangeJSON, 0, len(s.TimeRanges))
	for _, r := range s.TimeRanges {
		tr := timeRangeJSON{
			Start:   r.Start.Format(time.RFC3339),
			DurSecs: r.Duration.Seconds(),
		}
		if !r.End.IsZero() {
			tr.End = r.End.Format(time.RFC3339)
		}
		ranges = append(ranges, tr)
	}
	return timelineItemJSON{
		Project:    s.Project,
		TagID:      s.Tag,
		TagName:    s.TagName,
		TotalSecs:  s.TotalTime.Seconds(),
		TimeRanges: ranges,
	}
}

// --- new/switch/stop コマンド ---

// actionJSON は new/switch/stop コマンドの実行結果
type actionJSON struct {
	Action      string `json:"action"`
	Project     string `json:"project,omitempty"`
	TagID       string `json:"tag_id,omitempty"`
	TagName     string `json:"tag_name,omitempty"`
	StartTime   string `json:"start_time,omitempty"` // RFC3339
	StopTime    string `json:"stop_time,omitempty"`  // RFC3339
	PrevProject string `json:"prev_project,omitempty"`
}

// --- tag コマンド ---

// tagResultJSON は tag add/delete の結果
type tagResultJSON struct {
	Action string `json:"action"` // "added" or "deleted"
	ID     int    `json:"id"`
	Name   string `json:"name"`
}
