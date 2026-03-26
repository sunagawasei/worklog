package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"worklog/internal/domain"
)

// decodeJSON は buf から JSON をデコードするヘルパー
func decodeJSON(t *testing.T, buf *bytes.Buffer, v any) {
	t.Helper()
	if err := json.NewDecoder(buf).Decode(v); err != nil {
		t.Fatalf("JSON decode error: %v, body: %s", err, buf.String())
	}
}

// makeOpts は --json モードの ExecOptions を作るヘルパー
func makeOpts(args ...string) (ExecOptions, *bytes.Buffer) {
	buf := &bytes.Buffer{}
	allArgs := append([]string{"--json"}, args...)
	opts := parseGlobalFlags(allArgs)
	opts.Writer = buf
	return opts, buf
}

// TestJsonError は jsonError 関数のテスト
func TestJsonError(t *testing.T) {
	t.Run("JSONモード時にHandledErrorを返しWriterにJSON出力", func(t *testing.T) {
		buf := &bytes.Buffer{}
		opts := ExecOptions{JSONMode: true, NoInteractive: true, Args: []string{"test"}, Writer: buf}

		err := jsonError(opts, "SOME_CODE", "some message")

		if err == nil {
			t.Fatal("expected error, got nil")
		}
		var he HandledError
		if !errors.As(err, &he) {
			t.Errorf("expected HandledError, got %T: %v", err, err)
		}
		var out errorJSON
		decodeJSON(t, buf, &out)
		if out.Error != "SOME_CODE" {
			t.Errorf("error code: got %q, want %q", out.Error, "SOME_CODE")
		}
		if out.Message != "some message" {
			t.Errorf("message: got %q, want %q", out.Message, "some message")
		}
	})

	t.Run("非JSONモード時に通常エラーを返す", func(t *testing.T) {
		buf := &bytes.Buffer{}
		opts := ExecOptions{JSONMode: false, Args: []string{"test"}, Writer: buf}

		err := jsonError(opts, "CODE", "msg")

		if err == nil {
			t.Fatal("expected error, got nil")
		}
		var he HandledError
		if errors.As(err, &he) {
			t.Error("expected regular error, got HandledError")
		}
		if buf.Len() != 0 {
			t.Errorf("expected no output in non-JSON mode, got: %s", buf.String())
		}
	})
}

// TestHandleStatus_JSON は status --json のテスト
func TestHandleStatus_JSON(t *testing.T) {
	now := time.Now().Truncate(time.Second)

	t.Run("稼働中の場合はstatusオブジェクトを返す", func(t *testing.T) {
		manager := &mockProjectManager{
			status: &domain.ProjectStatus{
				Project:            "MyProject",
				Tag:                "1",
				TagName:            "開発",
				StartTime:          now,
				CurrentSessionTime: time.Hour,
				TotalTime:          2 * time.Hour,
			},
			summaries: []domain.ProjectSummary{},
		}
		opts, buf := makeOpts("status")

		if err := handleStatus(manager, opts); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		var out summaryStatusJSON
		decodeJSON(t, buf, &out)
		if out.Status == nil {
			t.Fatal("expected status != nil")
		}
		if out.Status.Project != "MyProject" {
			t.Errorf("project: got %q, want %q", out.Status.Project, "MyProject")
		}
		if out.Status.TotalSecs != 7200.0 {
			t.Errorf("total_seconds: got %f, want 7200.0", out.Status.TotalSecs)
		}
	})

	t.Run("停止中の場合はstatus nullを返す", func(t *testing.T) {
		manager := &mockProjectManager{
			status:    nil,
			summaries: []domain.ProjectSummary{},
		}
		opts, buf := makeOpts("status")

		if err := handleStatus(manager, opts); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		var out summaryStatusJSON
		decodeJSON(t, buf, &out)
		if out.Status != nil {
			t.Errorf("expected status == nil, got %+v", out.Status)
		}
	})
}

// TestHandleList_JSON は list --json のテスト
func TestHandleList_JSON(t *testing.T) {
	t.Run("プロジェクトがある場合はNDJSONを返す", func(t *testing.T) {
		manager := &mockProjectManager{
			summaries: []domain.ProjectSummary{
				{Project: "A", Tag: "1", TagName: "開発", TotalTime: time.Hour},
				{Project: "B", Tag: "2", TagName: "MTG", TotalTime: 30 * time.Minute},
			},
		}
		opts, buf := makeOpts("list")

		if err := handleList(manager, opts); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
		if len(lines) != 2 {
			t.Fatalf("expected 2 NDJSON lines, got %d: %s", len(lines), buf.String())
		}
		var first listItemJSON
		if err := json.Unmarshal([]byte(lines[0]), &first); err != nil {
			t.Fatalf("line1 decode: %v", err)
		}
		if first.Project != "A" {
			t.Errorf("project: got %q, want %q", first.Project, "A")
		}
		if first.TotalSecs != 3600.0 {
			t.Errorf("total_seconds: got %f, want 3600.0", first.TotalSecs)
		}
	})

	t.Run("プロジェクトが0件の場合は空出力", func(t *testing.T) {
		manager := &mockProjectManager{summaries: []domain.ProjectSummary{}}
		opts, buf := makeOpts("list")

		if err := handleList(manager, opts); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if buf.Len() != 0 {
			t.Errorf("expected empty output, got: %s", buf.String())
		}
	})
}

// TestHandleNew_JSON は new --json のテスト
func TestHandleNew_JSON(t *testing.T) {
	t.Run("正常に新規プロジェクトを開始してJSONを返す", func(t *testing.T) {
		manager := &mockProjectManager{status: nil}
		opts, buf := makeOpts("new", "MyProject", "1")

		if err := handleNew(manager, opts); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		var out actionJSON
		decodeJSON(t, buf, &out)
		if out.Action != "new" {
			t.Errorf("action: got %q, want %q", out.Action, "new")
		}
		if out.Project != "MyProject" {
			t.Errorf("project: got %q, want %q", out.Project, "MyProject")
		}
		if out.TagID != "1" {
			t.Errorf("tag_id: got %q, want %q", out.TagID, "1")
		}
	})

	t.Run("引数不足でMISSING_ARGUMENTSエラー", func(t *testing.T) {
		manager := &mockProjectManager{}
		opts, buf := makeOpts("new")

		err := handleNew(manager, opts)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		var out errorJSON
		decodeJSON(t, buf, &out)
		if out.Error != "MISSING_ARGUMENTS" {
			t.Errorf("error code: got %q, want %q", out.Error, "MISSING_ARGUMENTS")
		}
	})

	t.Run("不正なプロジェクト名でINVALID_PROJECT_NAMEエラー", func(t *testing.T) {
		manager := &mockProjectManager{}
		opts, buf := makeOpts("new", "bad\x00name", "1")

		err := handleNew(manager, opts)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		var out errorJSON
		decodeJSON(t, buf, &out)
		if out.Error != "INVALID_PROJECT_NAME" {
			t.Errorf("error code: got %q, want %q", out.Error, "INVALID_PROJECT_NAME")
		}
	})

	t.Run("非数値タグIDでINVALID_TAG_IDエラー", func(t *testing.T) {
		manager := &mockProjectManager{}
		opts, buf := makeOpts("new", "MyProject", "abc")

		err := handleNew(manager, opts)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		var out errorJSON
		decodeJSON(t, buf, &out)
		if out.Error != "INVALID_TAG_ID" {
			t.Errorf("error code: got %q, want %q", out.Error, "INVALID_TAG_ID")
		}
	})
}

// TestHandleStop_JSON は stop --json のテスト
func TestHandleStop_JSON(t *testing.T) {
	now := time.Now().Truncate(time.Second)

	t.Run("正常停止でJSONを返す", func(t *testing.T) {
		manager := &mockProjectManager{
			status: &domain.ProjectStatus{
				Project:   "MyProject",
				Tag:       "1",
				TagName:   "開発",
				StartTime: now,
			},
		}
		opts, buf := makeOpts("stop")

		if err := handleStop(manager, opts); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		var out actionJSON
		decodeJSON(t, buf, &out)
		if out.Action != "stop" {
			t.Errorf("action: got %q, want %q", out.Action, "stop")
		}
		if out.Project != "MyProject" {
			t.Errorf("project: got %q, want %q", out.Project, "MyProject")
		}
	})

	t.Run("稼働中プロジェクトなしでNO_ACTIVE_PROJECTエラー", func(t *testing.T) {
		manager := &mockProjectManager{status: nil}
		opts, buf := makeOpts("stop")

		err := handleStop(manager, opts)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		var out errorJSON
		decodeJSON(t, buf, &out)
		if out.Error != "NO_ACTIVE_PROJECT" {
			t.Errorf("error code: got %q, want %q", out.Error, "NO_ACTIVE_PROJECT")
		}
	})
}

// TestHandleTag_JSON は tag --json のテスト
func TestHandleTag_JSON(t *testing.T) {
	t.Run("tag list でJSON配列を返す", func(t *testing.T) {
		manager := &mockProjectManager{
			tags: []domain.Tag{
				{ID: 1, Name: "開発"},
				{ID: 2, Name: "MTG"},
			},
		}
		opts, buf := makeOpts("tag", "list")

		if err := handleTag(manager, opts); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		var tags []domain.Tag
		decodeJSON(t, buf, &tags)
		if len(tags) != 2 {
			t.Fatalf("expected 2 tags, got %d", len(tags))
		}
		if tags[0].Name != "開発" {
			t.Errorf("tag[0].name: got %q, want %q", tags[0].Name, "開発")
		}
	})

	t.Run("tag list でタグ0件のとき空配列を返す", func(t *testing.T) {
		manager := &mockProjectManager{tags: nil}
		opts, buf := makeOpts("tag", "list")

		if err := handleTag(manager, opts); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		var tags []domain.Tag
		decodeJSON(t, buf, &tags)
		if tags == nil {
			t.Error("expected empty array, got null")
		}
		if len(tags) != 0 {
			t.Errorf("expected 0 tags, got %d", len(tags))
		}
	})

	t.Run("tag add でtagResultJSONを返す", func(t *testing.T) {
		manager := &mockProjectManager{
			addTagResult: domain.Tag{ID: 3, Name: "新タグ"},
		}
		opts, buf := makeOpts("tag", "add", "新タグ")

		if err := handleTag(manager, opts); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		var out tagResultJSON
		decodeJSON(t, buf, &out)
		if out.Action != "added" {
			t.Errorf("action: got %q, want %q", out.Action, "added")
		}
		if out.ID != 3 {
			t.Errorf("id: got %d, want 3", out.ID)
		}
	})

	t.Run("不正なタグ名でINVALID_TAG_NAMEエラー", func(t *testing.T) {
		manager := &mockProjectManager{}
		opts, buf := makeOpts("tag", "add", "bad\x00tag")

		err := handleTag(manager, opts)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		var out errorJSON
		decodeJSON(t, buf, &out)
		if out.Error != "INVALID_TAG_NAME" {
			t.Errorf("error code: got %q, want %q", out.Error, "INVALID_TAG_NAME")
		}
	})

	t.Run("tag delete --no-interactiveでINTERACTIVE_REQUIREDエラー", func(t *testing.T) {
		manager := &mockProjectManager{
			tags: []domain.Tag{{ID: 1, Name: "開発"}},
		}
		opts, buf := makeOpts("tag", "delete", "1")

		err := handleTag(manager, opts)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		var out errorJSON
		decodeJSON(t, buf, &out)
		if out.Error != "INTERACTIVE_REQUIRED" {
			t.Errorf("error code: got %q, want %q", out.Error, "INTERACTIVE_REQUIRED")
		}
	})
}
