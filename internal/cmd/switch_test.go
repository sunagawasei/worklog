package cmd

import (
	"strings"
	"testing"
	"time"

	"worklog/internal/domain"
)

// parseSwitchTime は actionJSON.StartTime（RFC3339）をパースして time.Time を返す
func parseSwitchTime(t *testing.T, startTime string) time.Time {
	t.Helper()
	ts, err := time.Parse(time.RFC3339, startTime)
	if err != nil {
		t.Fatalf("start_time parse error: %v (value: %q)", err, startTime)
	}
	return ts
}

// TestHandleSwitch_JSON は handleSwitch の --json / --no-interactive 経路を固定する
// 対話 UI を使う経路は統合テスト対象のため除外
func TestHandleSwitch_JSON(t *testing.T) {
	t.Run("引数3つの成功 → action=switch JSON出力", func(t *testing.T) {
		manager := &mockProjectManager{status: nil}
		opts, buf := makeOpts("switch", "ProjectA", "1")

		if err := handleSwitch(manager, opts); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		var out actionJSON
		decodeJSON(t, buf, &out)
		if out.Action != "switch" {
			t.Errorf("action: got %q, want switch", out.Action)
		}
		if out.Project != "ProjectA" {
			t.Errorf("project: got %q, want ProjectA", out.Project)
		}
		if out.TagID != "1" {
			t.Errorf("tag_id: got %q, want 1", out.TagID)
		}
		if out.StartTime == "" {
			t.Error("start_time should not be empty")
		}
	})

	t.Run("旧プロジェクトあり → prev_project付与", func(t *testing.T) {
		now := time.Now().Truncate(time.Second)
		manager := &mockProjectManager{
			status: &domain.ProjectStatus{
				Project:   "OldProject",
				Tag:       "2",
				StartTime: now,
			},
		}
		opts, buf := makeOpts("switch", "NewProject", "3")

		if err := handleSwitch(manager, opts); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		var out actionJSON
		decodeJSON(t, buf, &out)
		if out.PrevProject != "OldProject" {
			t.Errorf("prev_project: got %q, want OldProject", out.PrevProject)
		}
		if out.Project != "NewProject" {
			t.Errorf("project: got %q, want NewProject", out.Project)
		}
	})

	t.Run("旧プロジェクトなし → prev_projectフィールドなし", func(t *testing.T) {
		manager := &mockProjectManager{status: nil}
		opts, buf := makeOpts("switch", "ProjectA", "1")

		if err := handleSwitch(manager, opts); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// prev_project が空の場合、omitempty で JSON フィールドが省略されること
		raw := buf.String()
		if strings.Contains(raw, `"prev_project"`) {
			t.Errorf("prev_project should be absent when no previous project, got: %s", raw)
		}
	})

	t.Run("引数不足+NoInteractive → MISSING_ARGUMENTS", func(t *testing.T) {
		// makeOpts("switch") は --json が付くため NoInteractive = true
		manager := &mockProjectManager{}
		opts, buf := makeOpts("switch")

		err := handleSwitch(manager, opts)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		var out errorJSON
		decodeJSON(t, buf, &out)
		if out.Error != "MISSING_ARGUMENTS" {
			t.Errorf("error code: got %q, want MISSING_ARGUMENTS", out.Error)
		}
	})

	t.Run("無効なプロジェクト名 → INVALID_PROJECT_NAME", func(t *testing.T) {
		manager := &mockProjectManager{}
		opts, buf := makeOpts("switch", "bad\x00name", "1")

		err := handleSwitch(manager, opts)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		var out errorJSON
		decodeJSON(t, buf, &out)
		if out.Error != "INVALID_PROJECT_NAME" {
			t.Errorf("error code: got %q, want INVALID_PROJECT_NAME", out.Error)
		}
	})

	t.Run("非数値タグID → INVALID_TAG_ID", func(t *testing.T) {
		manager := &mockProjectManager{}
		opts, buf := makeOpts("switch", "ProjectA", "abc")

		err := handleSwitch(manager, opts)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		var out errorJSON
		decodeJSON(t, buf, &out)
		if out.Error != "INVALID_TAG_ID" {
			t.Errorf("error code: got %q, want INVALID_TAG_ID", out.Error)
		}
	})

	t.Run("第4引数に有効な時刻 → SwitchAtが呼ばれ start_timeが10:30に反映", func(t *testing.T) {
		manager := &mockProjectManager{status: nil}
		opts, buf := makeOpts("switch", "ProjectA", "1", "10:30")

		if err := handleSwitch(manager, opts); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		var out actionJSON
		decodeJSON(t, buf, &out)
		if out.Action != "switch" {
			t.Errorf("action: got %q, want switch", out.Action)
		}
		if out.StartTime == "" {
			t.Fatal("start_time should not be empty")
		}

		// start_time が 10:30 を指しているか確認
		ts := parseSwitchTime(t, out.StartTime)
		if ts.Hour() != 10 || ts.Minute() != 30 {
			t.Errorf("start_time hour:minute: got %02d:%02d, want 10:30", ts.Hour(), ts.Minute())
		}

		// SwitchAt が呼ばれていること
		found := false
		for _, m := range manager.calledMethods {
			if strings.HasPrefix(m, "SwitchAt(") {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("SwitchAt should have been called, methods: %v", manager.calledMethods)
		}
	})

	t.Run("第4引数に無効な時刻 → INVALID_TIME_FORMAT", func(t *testing.T) {
		manager := &mockProjectManager{status: nil}
		opts, buf := makeOpts("switch", "ProjectA", "1", "99:99")

		err := handleSwitch(manager, opts)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		var out errorJSON
		decodeJSON(t, buf, &out)
		if out.Error != "INVALID_TIME_FORMAT" {
			t.Errorf("error code: got %q, want INVALID_TIME_FORMAT", out.Error)
		}
	})
}
