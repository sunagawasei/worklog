package cmd

import (
	"bytes"
	"encoding/json"
	"testing"
	"time"

	"worklog/internal/domain"
)

// TestToTimelineItemJSON は toTimelineItemJSON の JSON 出力契約を固定する
func TestToTimelineItemJSON(t *testing.T) {
	base := time.Date(2025, 6, 1, 10, 0, 0, 0, time.UTC)

	t.Run("完了セッション → start/end/duration_secondsが揃う", func(t *testing.T) {
		end := base.Add(90 * time.Minute)
		s := domain.ProjectSummary{
			Project:   "MyProject",
			Tag:       "1",
			TagName:   "開発",
			TotalTime: 90 * time.Minute,
			TimeRanges: []domain.TimeRange{
				{
					Start:    base,
					End:      end,
					Duration: 90 * time.Minute,
				},
			},
		}

		item := toTimelineItemJSON(s)

		if item.Project != "MyProject" {
			t.Errorf("project: got %q, want %q", item.Project, "MyProject")
		}
		if item.TagID != "1" {
			t.Errorf("tag_id: got %q, want %q", item.TagID, "1")
		}
		if item.TagName != "開発" {
			t.Errorf("tag_name: got %q, want %q", item.TagName, "開発")
		}
		if item.TotalSecs != 5400.0 {
			t.Errorf("total_seconds: got %f, want 5400.0", item.TotalSecs)
		}
		if len(item.TimeRanges) != 1 {
			t.Fatalf("time_ranges len: got %d, want 1", len(item.TimeRanges))
		}

		tr := item.TimeRanges[0]
		if tr.Start != base.Format(time.RFC3339) {
			t.Errorf("start: got %q, want %q", tr.Start, base.Format(time.RFC3339))
		}
		if tr.End != end.Format(time.RFC3339) {
			t.Errorf("end: got %q, want %q", tr.End, end.Format(time.RFC3339))
		}
		if tr.DurSecs != 5400.0 {
			t.Errorf("duration_seconds: got %f, want 5400.0", tr.DurSecs)
		}
	})

	t.Run("進行中セッション（End.IsZero()）→ 構造体の End は空・JSONにendキーなし", func(t *testing.T) {
		// json_output.go:92: if !r.End.IsZero() { tr.End = ... }
		// omitempty により end フィールドが省略されること（API契約）
		s := domain.ProjectSummary{
			Project: "ActiveProject",
			Tag:     "2",
			TimeRanges: []domain.TimeRange{
				{
					Start:    base,
					End:      time.Time{}, // ゼロ値 = 進行中
					Duration: 30 * time.Minute,
				},
			},
		}

		item := toTimelineItemJSON(s)

		if len(item.TimeRanges) != 1 {
			t.Fatalf("time_ranges len: got %d, want 1", len(item.TimeRanges))
		}
		tr := item.TimeRanges[0]
		if tr.End != "" {
			t.Errorf("End field should be empty for in-progress session, got %q", tr.End)
		}

		// JSONシリアライズして "end" キーが存在しないことを確認
		data, err := json.Marshal(tr)
		if err != nil {
			t.Fatalf("marshal error: %v", err)
		}
		var m map[string]any
		if err := json.Unmarshal(data, &m); err != nil {
			t.Fatalf("unmarshal error: %v", err)
		}
		if _, ok := m["end"]; ok {
			t.Errorf("'end' key must be absent in JSON for in-progress session, got: %s", data)
		}
	})

	t.Run("TimeRanges nil → time_rangesは空スライス（nullではない）", func(t *testing.T) {
		// json_output.go:86: make([]timeRangeJSON, 0, ...) で常に空スライスを生成
		s := domain.ProjectSummary{
			Project:    "EmptyProject",
			Tag:        "1",
			TotalTime:  0,
			TimeRanges: nil, // nil スライスを渡す
		}

		item := toTimelineItemJSON(s)

		if item.TimeRanges == nil {
			t.Error("time_ranges should be empty slice, not nil")
		}
		if len(item.TimeRanges) != 0 {
			t.Errorf("time_ranges len: got %d, want 0", len(item.TimeRanges))
		}

		// JSONシリアライズして "time_ranges":[] になることを確認（null ではない）
		data, err := json.Marshal(item)
		if err != nil {
			t.Fatalf("marshal error: %v", err)
		}
		if !bytes.Contains(data, []byte(`"time_ranges":[]`)) {
			t.Errorf("expected \"time_ranges\":[] in JSON, got: %s", data)
		}
	})

	t.Run("複数セッション → 全 range が正しく変換される", func(t *testing.T) {
		end1 := base.Add(30 * time.Minute)
		start2 := base.Add(60 * time.Minute)
		end2 := base.Add(90 * time.Minute)
		s := domain.ProjectSummary{
			Project:   "MultiSession",
			Tag:       "3",
			TotalTime: 60 * time.Minute,
			TimeRanges: []domain.TimeRange{
				{Start: base, End: end1, Duration: 30 * time.Minute},
				{Start: start2, End: end2, Duration: 30 * time.Minute},
			},
		}

		item := toTimelineItemJSON(s)

		if len(item.TimeRanges) != 2 {
			t.Fatalf("time_ranges len: got %d, want 2", len(item.TimeRanges))
		}
		if item.TimeRanges[0].Start != base.Format(time.RFC3339) {
			t.Errorf("ranges[0].start: got %q, want %q", item.TimeRanges[0].Start, base.Format(time.RFC3339))
		}
		if item.TimeRanges[1].Start != start2.Format(time.RFC3339) {
			t.Errorf("ranges[1].start: got %q, want %q", item.TimeRanges[1].Start, start2.Format(time.RFC3339))
		}
	})
}
