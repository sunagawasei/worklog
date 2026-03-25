package project

import (
	"errors"
	"os"
	"testing"
	"time"

	"worklog/internal/domain"
	"worklog/internal/storage"
)

// モックのCurrentStorage
type mockCurrentStorage struct {
	readResult   string
	readError    error
	writeProject string
	writeTag     string
	writeError   error
	cleared      bool
	clearError   error
}

func (m *mockCurrentStorage) Read() (string, error) {
	return m.readResult, m.readError
}

func (m *mockCurrentStorage) Write(project, tag string) error {
	m.writeProject = project
	m.writeTag = tag
	return m.writeError
}

func (m *mockCurrentStorage) Clear() error {
	m.cleared = true
	return m.clearError
}

// モックのLogStorage
type mockLogStorage struct {
	appendCalls         []appendCall
	appendError         error
	readResult          []storage.LogEntry
	readError           error
	readAllResult       map[string][]storage.LogEntry
	readAllError        error
	readAllOnDateResult map[string][]storage.LogEntry
	readAllOnDateError  error
	readRangeResult     map[string][]storage.LogEntry
	readRangeError      error
}

type appendCall struct {
	project   string
	action    string
	tag       string
	timestamp time.Time
}

func (m *mockLogStorage) Append(project, action, tag string, timestamp time.Time) error {
	m.appendCalls = append(m.appendCalls, appendCall{
		project:   project,
		action:    action,
		tag:       tag,
		timestamp: timestamp,
	})
	return m.appendError
}

func (m *mockLogStorage) ReadToday(project string) ([]storage.LogEntry, error) {
	return m.readResult, m.readError
}

func (m *mockLogStorage) ReadAllToday() (map[string][]storage.LogEntry, error) {
	return m.readAllResult, m.readAllError
}

func (m *mockLogStorage) ReadAllOnDate(date time.Time) (map[string][]storage.LogEntry, error) {
	return m.readAllOnDateResult, m.readAllOnDateError
}

func (m *mockLogStorage) ReadRange(startDate, endDate time.Time) (map[string][]storage.LogEntry, error) {
	return m.readRangeResult, m.readRangeError
}

// モックのTagStorage
type mockTagStorage struct {
	tags      []storage.Tag
	loadError error
}

func (m *mockTagStorage) Load() ([]storage.Tag, error) {
	return m.tags, m.loadError
}

func (m *mockTagStorage) Save(tags []storage.Tag) error {
	m.tags = tags
	return nil
}

func (m *mockTagStorage) Add(name string) (storage.Tag, error) {
	// 同名チェック
	for _, tag := range m.tags {
		if tag.Name == name {
			return storage.Tag{}, os.ErrExist
		}
	}

	// 最大IDを見つける
	maxID := 0
	for _, tag := range m.tags {
		if tag.ID > maxID {
			maxID = tag.ID
		}
	}

	// 新しいタグを作成
	newTag := storage.Tag{
		ID:   maxID + 1,
		Name: name,
	}

	m.tags = append(m.tags, newTag)
	return newTag, nil
}

func (m *mockTagStorage) Delete(id int) error {
	// 指定したIDのタグを探して削除
	found := false
	newTags := make([]storage.Tag, 0, len(m.tags))
	for _, tag := range m.tags {
		if tag.ID == id {
			found = true
		} else {
			newTags = append(newTags, tag)
		}
	}

	if !found {
		return os.ErrNotExist
	}

	m.tags = newTags
	return nil
}

func TestProjectManager_New(t *testing.T) {
	// モックの準備
	currentStorage := &mockCurrentStorage{}
	logStorage := &mockLogStorage{}
	tagStorage := &mockTagStorage{
		tags: []storage.Tag{
			{ID: 1, Name: "開発"},
			{ID: 2, Name: "会議"},
		},
	}

	// ProjectManagerの作成
	manager := NewProjectManager(currentStorage, logStorage, tagStorage)

	// NewAtメソッドのテスト
	err := manager.NewAt("ProjectA", "Development", time.Now())
	if err != nil {
		t.Errorf("NewAt()でエラーが発生: %v", err)
	}

	// CurrentStorageへの書き込みを確認
	if currentStorage.writeProject != "ProjectA" {
		t.Errorf("プロジェクト名が異なる: got %s, want ProjectA", currentStorage.writeProject)
	}
	if currentStorage.writeTag != "Development" {
		t.Errorf("タグIDが異なる: got %s, want Development", currentStorage.writeTag)
	}

	// LogStorageへの記録を確認
	if len(logStorage.appendCalls) != 1 {
		t.Errorf("ログ記録回数が異なる: got %d, want 1", len(logStorage.appendCalls))
	} else {
		call := logStorage.appendCalls[0]
		if call.project != "ProjectA" {
			t.Errorf("ログのプロジェクト名が異なる: got %s, want ProjectA", call.project)
		}
		if call.action != "start" {
			t.Errorf("ログのアクションが異なる: got %s, want start", call.action)
		}
		if call.tag != "Development" {
			t.Errorf("ログのタグIDが異なる: got %s, want Development", call.tag)
		}
	}
}

func TestProjectManager_New_WithAutoStop(t *testing.T) {
	t.Run("稼働中のプロジェクトを自動停止して新規開始", func(t *testing.T) {
		// モックの準備（すでにプロジェクトが稼働中）
		currentStorage := &mockCurrentStorage{
			readResult: "ExistingProject\tMTG",
		}
		logStorage := &mockLogStorage{}
		tagStorage := &mockTagStorage{}

		// ProjectManagerの作成
		manager := NewProjectManager(currentStorage, logStorage, tagStorage)

		// NewAtメソッドのテスト（自動停止を期待）
		err := manager.NewAt("ProjectA", "Development", time.Now())
		if err != nil {
			t.Errorf("NewAt()でエラーが発生しました: %v", err)
			return
		}

		// LogStorageへの記録を確認（stop→startの2つ）
		if len(logStorage.appendCalls) != 2 {
			t.Errorf("ログ記録回数が異なる: got %d, want 2", len(logStorage.appendCalls))
		} else {
			// 最初のログはstop
			stopCall := logStorage.appendCalls[0]
			if stopCall.project != "ExistingProject" {
				t.Errorf("stopログのプロジェクト名が異なる: got %s, want ExistingProject", stopCall.project)
			}
			if stopCall.action != "stop" {
				t.Errorf("stopログのアクションが異なる: got %s, want stop", stopCall.action)
			}
			if stopCall.tag != "MTG" {
				t.Errorf("stopログのタグIDが異なる: got %s, want MTG", stopCall.tag)
			}

			// 2番目のログはstart
			startCall := logStorage.appendCalls[1]
			if startCall.project != "ProjectA" {
				t.Errorf("startログのプロジェクト名が異なる: got %s, want ProjectA", startCall.project)
			}
			if startCall.action != "start" {
				t.Errorf("startログのアクションが異なる: got %s, want start", startCall.action)
			}
			if startCall.tag != "Development" {
				t.Errorf("startログのタグIDが異なる: got %s, want Development", startCall.tag)
			}
		}

		// CurrentStorageへの書き込みを確認
		if currentStorage.writeProject != "ProjectA" {
			t.Errorf("プロジェクト名が異なる: got %s, want ProjectA", currentStorage.writeProject)
		}
		if currentStorage.writeTag != "Development" {
			t.Errorf("タグIDが異なる: got %s, want Development", currentStorage.writeTag)
		}
	})
}

func TestProjectManager_Switch(t *testing.T) {
	// モックの準備（既存のプロジェクトが稼働中）
	currentStorage := &mockCurrentStorage{
		readResult: "ProjectA\tDevelopment",
	}
	logStorage := &mockLogStorage{}
	tagStorage := &mockTagStorage{}

	// ProjectManagerの作成
	manager := NewProjectManager(currentStorage, logStorage, tagStorage)

	// SwitchAtメソッドのテスト
	err := manager.SwitchAt("ProjectB", "MTG", time.Now())
	if err != nil {
		t.Errorf("SwitchAt()でエラーが発生: %v", err)
	}

	// LogStorageへの記録を確認（stop→startの2つ）
	if len(logStorage.appendCalls) != 2 {
		t.Errorf("ログ記録回数が異なる: got %d, want 2", len(logStorage.appendCalls))
	} else {
		// 最初のログはstop
		stopCall := logStorage.appendCalls[0]
		if stopCall.project != "ProjectA" {
			t.Errorf("stopログのプロジェクト名が異なる: got %s, want ProjectA", stopCall.project)
		}
		if stopCall.action != "stop" {
			t.Errorf("stopログのアクションが異なる: got %s, want stop", stopCall.action)
		}
		if stopCall.tag != "Development" {
			t.Errorf("stopログのタグIDが異なる: got %s, want Development", stopCall.tag)
		}

		// 2番目のログはstart
		startCall := logStorage.appendCalls[1]
		if startCall.project != "ProjectB" {
			t.Errorf("startログのプロジェクト名が異なる: got %s, want ProjectB", startCall.project)
		}
		if startCall.action != "start" {
			t.Errorf("startログのアクションが異なる: got %s, want start", startCall.action)
		}
		if startCall.tag != "MTG" {
			t.Errorf("startログのタグIDが異なる: got %s, want MTG", startCall.tag)
		}
	}

	// CurrentStorageへの書き込みを確認
	if currentStorage.writeProject != "ProjectB" {
		t.Errorf("プロジェクト名が異なる: got %s, want ProjectB", currentStorage.writeProject)
	}
	if currentStorage.writeTag != "MTG" {
		t.Errorf("タグIDが異なる: got %s, want MTG", currentStorage.writeTag)
	}
}

func TestProjectManager_Stop(t *testing.T) {
	// モックの準備（プロジェクトが稼働中）
	currentStorage := &mockCurrentStorage{
		readResult: "ProjectA\tDevelopment",
	}
	logStorage := &mockLogStorage{}
	tagStorage := &mockTagStorage{}

	// ProjectManagerの作成
	manager := NewProjectManager(currentStorage, logStorage, tagStorage)

	// StopAtメソッドのテスト
	err := manager.StopAt(time.Now())
	if err != nil {
		t.Errorf("StopAt()でエラーが発生: %v", err)
	}

	// LogStorageへの記録を確認（stopの1つ）
	if len(logStorage.appendCalls) != 1 {
		t.Errorf("ログ記録回数が異なる: got %d, want 1", len(logStorage.appendCalls))
	} else {
		call := logStorage.appendCalls[0]
		if call.project != "ProjectA" {
			t.Errorf("ログのプロジェクト名が異なる: got %s, want ProjectA", call.project)
		}
		if call.action != "stop" {
			t.Errorf("ログのアクションが異なる: got %s, want stop", call.action)
		}
		if call.tag != "Development" {
			t.Errorf("ログのタグIDが異なる: got %s, want Development", call.tag)
		}
	}

	// CurrentStorageがクリアされたことを確認
	if !currentStorage.cleared {
		t.Error("currentファイルがクリアされるべき")
	}
}

func TestProjectManager_Status(t *testing.T) {
	t.Run("稼働中のプロジェクトがある場合（タグ名も取得）", func(t *testing.T) {
		// 期待される開始時刻
		expectedStartTime := time.Date(2025, 9, 26, 17, 30, 3, 0, time.Local)

		// モックの準備
		currentStorage := &mockCurrentStorage{
			readResult: "ProjectA\t5",
		}
		logStorage := &mockLogStorage{
			readResult: []storage.LogEntry{
				{
					Timestamp: expectedStartTime,
					Action:    "start",
					Tag:       "5",
				},
			},
		}
		tagStorage := &mockTagStorage{
			tags: []storage.Tag{
				{ID: 1, Name: "Backlog"},
				{ID: 5, Name: "Development"},
				{ID: 7, Name: "月次"},
			},
		}

		// ProjectManagerの作成
		manager := NewProjectManager(currentStorage, logStorage, tagStorage)

		// Statusメソッドのテスト
		status, err := manager.Status()
		if err != nil {
			t.Errorf("Status()でエラーが発生: %v", err)
		}

		// ステータスの確認
		if status == nil {
			t.Fatal("statusがnilではないべき")
		}
		if status.Project != "ProjectA" {
			t.Errorf("プロジェクト名が異なる: got %s, want ProjectA", status.Project)
		}
		if status.Tag != "5" {
			t.Errorf("タグIDが異なる: got %s, want 5", status.Tag)
		}
		// タグ名の確認（新しいフィールド）
		if status.TagName != "Development" {
			t.Errorf("タグ名が異なる: got %s, want Development", status.TagName)
		}
		// 開始時刻の確認（ログから取得されるべき）
		if !status.StartTime.Equal(expectedStartTime) {
			t.Errorf("開始時刻が異なる: got %v, want %v", status.StartTime, expectedStartTime)
		}
	})

	t.Run("稼働中のプロジェクトの累計時間を計算", func(t *testing.T) {
		// テスト用の時刻を定義
		baseTime := time.Date(2025, 9, 26, 10, 0, 0, 0, time.Local)

		// モックの準備
		currentStorage := &mockCurrentStorage{
			readResult: "ProjectA\t5",
		}
		logStorage := &mockLogStorage{
			readResult: []storage.LogEntry{
				{
					Timestamp: baseTime.Add(2 * time.Hour), // 12:00 最新のstart
					Action:    "start",
					Tag:       "5",
				},
			},
			readAllResult: map[string][]storage.LogEntry{
				"ProjectA": {
					{Timestamp: baseTime, Action: "start", Tag: "5"},                    // 10:00 start
					{Timestamp: baseTime.Add(1 * time.Hour), Action: "stop", Tag: "5"},  // 11:00 stop (1時間)
					{Timestamp: baseTime.Add(2 * time.Hour), Action: "start", Tag: "5"}, // 12:00 start
					{Timestamp: baseTime.Add(3 * time.Hour), Action: "stop", Tag: "5"},  // 13:00 stop (1時間)
					{Timestamp: baseTime.Add(4 * time.Hour), Action: "start", Tag: "5"}, // 14:00 start (現在稼働中)
				},
			},
		}
		tagStorage := &mockTagStorage{
			tags: []storage.Tag{
				{ID: 5, Name: "Development"},
			},
		}

		// ProjectManagerの作成
		manager := NewProjectManager(currentStorage, logStorage, tagStorage)

		// Statusメソッドのテスト
		status, err := manager.Status()
		if err != nil {
			t.Errorf("Status()でエラーが発生: %v", err)
		}

		// ステータスの確認
		if status == nil {
			t.Fatal("statusがnilではないべき")
		}

		// 累計時間の確認（新しい検証項目）
		// 注意: 現在時刻がテストの実行時刻に依存するため、範囲チェックを行う
		// 最低でも2時間（stop済みの分）は必ず含まれる
		// 期待される累計時間: 1時間 + 1時間 + (現在稼働中の時間) = 2時間以上
		minExpectedTime := 2 * time.Hour
		if status.TotalTime < minExpectedTime {
			t.Errorf("累計時間が期待値より少ない: got %v, want at least %v", status.TotalTime, minExpectedTime)
		}

		// 現在セッション時間の確認（新しい検証項目）
		// 最新のstartは baseTime.Add(2 * time.Hour) なので、現在時刻までの経過時間
		// テスト実行時刻に依存するため、0以上であることのみ確認
		if status.CurrentSessionTime < 0 {
			t.Errorf("現在セッション時間が負の値: got %v", status.CurrentSessionTime)
		}
	})

	t.Run("稼働中のプロジェクトがない場合", func(t *testing.T) {
		// モックの準備（currentが空）
		currentStorage := &mockCurrentStorage{
			readError: errors.New("no current project"),
		}
		logStorage := &mockLogStorage{}
		tagStorage := &mockTagStorage{}

		// ProjectManagerの作成
		manager := NewProjectManager(currentStorage, logStorage, tagStorage)

		// Statusメソッドのテスト（nilが返ることを期待）
		status, err := manager.Status()
		if err != nil {
			t.Errorf("Status()でエラーが発生: %v", err)
		}
		if status != nil {
			t.Error("稼働中のプロジェクトがない場合、statusはnilであるべき")
		}
	})
}

func TestProjectManager_Status_TagNameNotFound(t *testing.T) {
	t.Run("タグIDに対応するタグ名が見つからない場合", func(t *testing.T) {
		// モックの準備
		currentStorage := &mockCurrentStorage{
			readResult: "ProjectA\t999", // 存在しないタグID
		}
		logStorage := &mockLogStorage{}
		tagStorage := &mockTagStorage{
			tags: []storage.Tag{
				{ID: 1, Name: "Backlog"},
				{ID: 5, Name: "Development"},
			},
		}

		// ProjectManagerの作成
		manager := NewProjectManager(currentStorage, logStorage, tagStorage)

		// Statusメソッドのテスト
		status, err := manager.Status()
		if err != nil {
			t.Errorf("Status()でエラーが発生: %v", err)
		}

		// ステータスの確認
		if status == nil {
			t.Fatal("statusがnilではないべき")
		}
		if status.Project != "ProjectA" {
			t.Errorf("プロジェクト名が異なる: got %s, want ProjectA", status.Project)
		}
		if status.Tag != "999" {
			t.Errorf("タグIDが異なる: got %s, want 999", status.Tag)
		}
		// タグ名が空であることを確認
		if status.TagName != "" {
			t.Errorf("タグ名が見つからない場合は空文字列であるべき: got %s", status.TagName)
		}
	})
}

func TestProjectManager_Status_OpenStartNil(t *testing.T) {
	t.Run("全セッション終了済みだがcurrentに登録されている場合", func(t *testing.T) {
		baseTime := time.Date(2025, 9, 26, 11, 31, 0, 0, time.Local)
		currentStorage := &mockCurrentStorage{
			readResult: "worklog\t1",
		}
		logStorage := &mockLogStorage{
			readResult: []storage.LogEntry{
				{Timestamp: baseTime, Action: "start", Tag: "1"},
				{Timestamp: baseTime.Add(29 * time.Minute), Action: "stop", Tag: "1"},
			},
		}
		tagStorage := &mockTagStorage{
			tags: []storage.Tag{{ID: 1, Name: "開発"}},
		}

		manager := NewProjectManager(currentStorage, logStorage, tagStorage)
		status, err := manager.Status()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if status == nil {
			t.Fatal("status should not be nil")
		}
		// StartTimeは最後の完了セッションの開始時刻であるべき（time.Now()ではない）
		if !status.StartTime.Equal(baseTime) {
			t.Errorf("StartTime should be last session start: got %v, want %v", status.StartTime, baseTime)
		}
		// TotalTimeは完了セッション分
		expectedTotal := 29 * time.Minute
		if status.TotalTime != expectedTotal {
			t.Errorf("TotalTime: got %v, want %v", status.TotalTime, expectedTotal)
		}
		// CurrentSessionTimeはゼロ（オープンセッションなし）
		if status.CurrentSessionTime != 0 {
			t.Errorf("CurrentSessionTime should be 0: got %v", status.CurrentSessionTime)
		}
	})

	t.Run("ReadTodayが空エントリを返す場合", func(t *testing.T) {
		currentStorage := &mockCurrentStorage{
			readResult: "worklog\t1",
		}
		logStorage := &mockLogStorage{
			readResult: []storage.LogEntry{},
		}
		tagStorage := &mockTagStorage{
			tags: []storage.Tag{{ID: 1, Name: "開発"}},
		}

		manager := NewProjectManager(currentStorage, logStorage, tagStorage)
		status, err := manager.Status()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if status == nil {
			t.Fatal("status should not be nil")
		}
		// ログがない場合、StartTimeはゼロ値であるべき
		if !status.StartTime.IsZero() {
			t.Errorf("StartTime should be zero when no entries: got %v", status.StartTime)
		}
	})

	t.Run("ReadTodayがエラーの場合", func(t *testing.T) {
		currentStorage := &mockCurrentStorage{
			readResult: "worklog\t1",
		}
		logStorage := &mockLogStorage{
			readError: errors.New("file not found"),
		}
		tagStorage := &mockTagStorage{
			tags: []storage.Tag{{ID: 1, Name: "開発"}},
		}

		manager := NewProjectManager(currentStorage, logStorage, tagStorage)
		status, err := manager.Status()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if status == nil {
			t.Fatal("status should not be nil")
		}
		// エラー時もStartTimeはゼロ値であるべき
		if !status.StartTime.IsZero() {
			t.Errorf("StartTime should be zero when ReadToday errors: got %v", status.StartTime)
		}
	})
}

func TestProjectManager_List(t *testing.T) {
	// モックの準備
	currentStorage := &mockCurrentStorage{
		readResult: "ActiveProject\tDevelopment", // 現在稼働中
	}

	// 複数プロジェクトのログエントリを準備
	now := time.Now()
	logStorage := &mockLogStorage{
		readAllResult: map[string][]storage.LogEntry{
			"ProjectA": {
				{Action: "start", Tag: "Development", Timestamp: now.Add(-3 * time.Hour)},
				{Action: "stop", Tag: "Development", Timestamp: now.Add(-2 * time.Hour)},
			},
			"ProjectB": {
				{Action: "start", Tag: "MTG", Timestamp: now.Add(-90 * time.Minute)},
				{Action: "stop", Tag: "MTG", Timestamp: now.Add(-60 * time.Minute)},
			},
			"ActiveProject": {
				{Action: "start", Tag: "Development", Timestamp: now.Add(-30 * time.Minute)},
				// ActiveProjectは稼働中（stopなし）
			},
		},
	}
	tagStorage := &mockTagStorage{}

	// ProjectManagerの作成
	manager := NewProjectManager(currentStorage, logStorage, tagStorage)

	// Listメソッドのテスト
	summaries, err := manager.List()
	if err != nil {
		t.Errorf("List()でエラーが発生: %v", err)
	}

	// プロジェクト数の確認
	if len(summaries) != 3 {
		t.Errorf("プロジェクト数が異なる: got %d, want 3", len(summaries))
		return
	}

	// 各プロジェクトの稼働時間を確認
	// ProjectA: 1時間（start→stopの1セッション）
	projectA := findSummary(summaries, "ProjectA")
	if projectA == nil {
		t.Error("ProjectAが見つからない")
	} else if projectA.TotalTime != time.Hour {
		t.Errorf("ProjectAの稼働時間が異なる: got %v, want %v", projectA.TotalTime, time.Hour)
	}

	// ProjectB: 30分（start→stopの1セッション）
	projectB := findSummary(summaries, "ProjectB")
	if projectB == nil {
		t.Error("ProjectBが見つからない")
	} else if projectB.TotalTime != 30*time.Minute {
		t.Errorf("ProjectBの稼働時間が異なる: got %v, want %v", projectB.TotalTime, 30*time.Minute)
	}

	// ActiveProject: 30分（start→現在まで）
	activeProject := findSummary(summaries, "ActiveProject")
	if activeProject == nil {
		t.Error("ActiveProjectが見つからない")
	} else {
		// 現在稼働中なので、約30分（多少の誤差を許容）
		expectedTime := 30 * time.Minute
		tolerance := 1 * time.Minute
		if activeProject.TotalTime < expectedTime-tolerance || activeProject.TotalTime > expectedTime+tolerance {
			t.Errorf("ActiveProjectの稼働時間が異なる: got %v, want around %v", activeProject.TotalTime, expectedTime)
		}
	}
}

func findSummary(summaries []domain.ProjectSummary, project string) *domain.ProjectSummary {
	for _, s := range summaries {
		if s.Project == project {
			return &s
		}
	}
	return nil
}

// TestProjectManager_List_WithTimeRanges は時間範囲が正しく収集されることをテストする
func TestProjectManager_List_WithTimeRanges(t *testing.T) {
	// モックの準備
	currentStorage := &mockCurrentStorage{
		readResult: "", // 稼働中なし
	}

	// 複数の時間範囲を含むログエントリを準備
	baseTime := time.Date(2025, 9, 30, 10, 0, 0, 0, time.Local)
	logStorage := &mockLogStorage{
		readAllResult: map[string][]storage.LogEntry{
			"ProjectA": {
				// 1つ目の時間範囲: 10:00-10:50 (50分)
				{Action: "start", Tag: "1", Timestamp: baseTime},
				{Action: "stop", Tag: "1", Timestamp: baseTime.Add(50 * time.Minute)},
				// 2つ目の時間範囲: 13:00-14:40 (1時間40分)
				{Action: "start", Tag: "1", Timestamp: baseTime.Add(3 * time.Hour)},
				{Action: "stop", Tag: "1", Timestamp: baseTime.Add(4*time.Hour + 40*time.Minute)},
			},
			"ProjectB": {
				// 1つの時間範囲: 11:00-12:00 (1時間)
				{Action: "start", Tag: "2", Timestamp: baseTime.Add(1 * time.Hour)},
				{Action: "stop", Tag: "2", Timestamp: baseTime.Add(2 * time.Hour)},
			},
		},
	}
	tagStorage := &mockTagStorage{}

	// ProjectManagerの作成
	manager := NewProjectManager(currentStorage, logStorage, tagStorage)

	// Listメソッドのテスト
	summaries, err := manager.List()
	if err != nil {
		t.Fatalf("List()でエラーが発生: %v", err)
	}

	// ProjectAの時間範囲を確認
	projectA := findSummary(summaries, "ProjectA")
	if projectA == nil {
		t.Fatal("ProjectAが見つからない")
	}

	// TimeRangesが2つあることを確認
	if len(projectA.TimeRanges) != 2 {
		t.Errorf("ProjectAの時間範囲数が異なる: got %d, want 2", len(projectA.TimeRanges))
	} else {
		// 1つ目の時間範囲を確認
		tr1 := projectA.TimeRanges[0]
		if !tr1.Start.Equal(baseTime) {
			t.Errorf("1つ目の開始時刻が異なる: got %v, want %v", tr1.Start, baseTime)
		}
		expectedEnd1 := baseTime.Add(50 * time.Minute)
		if !tr1.End.Equal(expectedEnd1) {
			t.Errorf("1つ目の終了時刻が異なる: got %v, want %v", tr1.End, expectedEnd1)
		}
		if tr1.Duration != 50*time.Minute {
			t.Errorf("1つ目の稼働時間が異なる: got %v, want 50m", tr1.Duration)
		}

		// 2つ目の時間範囲を確認
		tr2 := projectA.TimeRanges[1]
		expectedStart2 := baseTime.Add(3 * time.Hour)
		if !tr2.Start.Equal(expectedStart2) {
			t.Errorf("2つ目の開始時刻が異なる: got %v, want %v", tr2.Start, expectedStart2)
		}
		expectedEnd2 := baseTime.Add(4*time.Hour + 40*time.Minute)
		if !tr2.End.Equal(expectedEnd2) {
			t.Errorf("2つ目の終了時刻が異なる: got %v, want %v", tr2.End, expectedEnd2)
		}
		expectedDuration2 := 1*time.Hour + 40*time.Minute
		if tr2.Duration != expectedDuration2 {
			t.Errorf("2つ目の稼働時間が異なる: got %v, want 1h40m", tr2.Duration)
		}
	}

	// ProjectBの時間範囲を確認
	projectB := findSummary(summaries, "ProjectB")
	if projectB == nil {
		t.Fatal("ProjectBが見つからない")
	}

	// TimeRangesが1つあることを確認
	if len(projectB.TimeRanges) != 1 {
		t.Errorf("ProjectBの時間範囲数が異なる: got %d, want 1", len(projectB.TimeRanges))
	} else {
		tr := projectB.TimeRanges[0]
		expectedStart := baseTime.Add(1 * time.Hour)
		if !tr.Start.Equal(expectedStart) {
			t.Errorf("開始時刻が異なる: got %v, want %v", tr.Start, expectedStart)
		}
		expectedEnd := baseTime.Add(2 * time.Hour)
		if !tr.End.Equal(expectedEnd) {
			t.Errorf("終了時刻が異なる: got %v, want %v", tr.End, expectedEnd)
		}
		if tr.Duration != 1*time.Hour {
			t.Errorf("稼働時間が異なる: got %v, want 1h", tr.Duration)
		}
	}
}

// TestProjectManager_List_SortedByTagName はタグ名でグループ化されてソートされることをテストする
func TestProjectManager_List_SortedByTagName(t *testing.T) {
	t.Run("タグ名でグループ化され、同一タグ内では最終アクティビティ降順でソート", func(t *testing.T) {
		// モックの準備
		currentStorage := &mockCurrentStorage{
			readResult: "", // 稼働中なし
		}

		// 複数タグのプロジェクトを準備（期待される表示順と異なる順序で定義）
		baseTime := time.Date(2025, 10, 14, 10, 0, 0, 0, time.Local)
		logStorage := &mockLogStorage{
			readAllResult: map[string][]storage.LogEntry{
				// Backlogタグのプロジェクト
				"TASK-002": {
					{Action: "start", Tag: "1", Timestamp: baseTime},
					{Action: "stop", Tag: "1", Timestamp: baseTime.Add(1 * time.Hour)},
				},
				"windowsサーバ": {
					{Action: "start", Tag: "1", Timestamp: baseTime.Add(2 * time.Hour)},
					{Action: "stop", Tag: "1", Timestamp: baseTime.Add(3 * time.Hour)},
				},
				"TASK-001": {
					{Action: "start", Tag: "1", Timestamp: baseTime.Add(4 * time.Hour)},
					{Action: "stop", Tag: "1", Timestamp: baseTime.Add(5 * time.Hour)},
				},
				// Learnタグのプロジェクト
				"Exploring CTRL+G": {
					{Action: "start", Tag: "2", Timestamp: baseTime.Add(1 * time.Hour)},
					{Action: "stop", Tag: "2", Timestamp: baseTime.Add(2 * time.Hour)},
				},
				// othersタグのプロジェクト
				"HW定例": {
					{Action: "start", Tag: "3", Timestamp: baseTime.Add(3 * time.Hour)},
					{Action: "stop", Tag: "3", Timestamp: baseTime.Add(4 * time.Hour)},
				},
				"稼働集計": {
					{Action: "start", Tag: "3", Timestamp: baseTime.Add(5 * time.Hour)},
					{Action: "stop", Tag: "3", Timestamp: baseTime.Add(6 * time.Hour)},
				},
				// Backendタグのプロジェクト
				"発行": {
					{Action: "start", Tag: "4", Timestamp: baseTime.Add(2 * time.Hour)},
					{Action: "stop", Tag: "4", Timestamp: baseTime.Add(3 * time.Hour)},
				},
			},
		}
		tagStorage := &mockTagStorage{
			tags: []storage.Tag{
				{ID: 1, Name: "Backlog"},
				{ID: 2, Name: "Learn"},
				{ID: 3, Name: "others"},
				{ID: 4, Name: "Backend"},
			},
		}

		// ProjectManagerの作成
		manager := NewProjectManager(currentStorage, logStorage, tagStorage)

		// Listメソッドのテスト
		summaries, err := manager.List()
		if err != nil {
			t.Fatalf("List()でエラーが発生: %v", err)
		}

		// プロジェクト数の確認
		if len(summaries) != 7 {
			t.Fatalf("プロジェクト数が異なる: got %d, want 7", len(summaries))
		}

		// タグ名でグループ化されていることを確認（Backend → Backlog → Learn → others）
		// 注意: アルファベット順ではBackend < Backlog < Learn < others
		expectedOrder := []struct {
			project string
			tagName string
		}{
			{"発行", "Backend"},       // Backendグループ
			{"TASK-001", "Backlog"}, // Backlogグループ（最終アクティビティ降順）
			{"windowsサーバ", "Backlog"},
			{"TASK-002", "Backlog"},
			{"Exploring CTRL+G", "Learn"}, // Learnグループ
			{"稼働集計", "others"},            // othersグループ（最終アクティビティ降順）
			{"HW定例", "others"},
		}

		for i, expected := range expectedOrder {
			if summaries[i].Project != expected.project {
				t.Errorf("位置%dのプロジェクトが異なる: got %s, want %s", i, summaries[i].Project, expected.project)
			}
			if summaries[i].TagName != expected.tagName {
				t.Errorf("位置%dのタグ名が異なる: got %s, want %s", i, summaries[i].TagName, expected.tagName)
			}
		}
	})

	t.Run("タグ名が空の場合は最後に表示", func(t *testing.T) {
		// モックの準備
		currentStorage := &mockCurrentStorage{
			readResult: "",
		}

		baseTime := time.Date(2025, 10, 14, 10, 0, 0, 0, time.Local)
		logStorage := &mockLogStorage{
			readAllResult: map[string][]storage.LogEntry{
				"ProjectA": {
					{Action: "start", Tag: "1", Timestamp: baseTime},
					{Action: "stop", Tag: "1", Timestamp: baseTime.Add(1 * time.Hour)},
				},
				"ProjectB": {
					{Action: "start", Tag: "999", Timestamp: baseTime.Add(1 * time.Hour)}, // 存在しないタグID
					{Action: "stop", Tag: "999", Timestamp: baseTime.Add(2 * time.Hour)},
				},
			},
		}
		tagStorage := &mockTagStorage{
			tags: []storage.Tag{
				{ID: 1, Name: "Backlog"},
			},
		}

		manager := NewProjectManager(currentStorage, logStorage, tagStorage)

		summaries, err := manager.List()
		if err != nil {
			t.Fatalf("List()でエラーが発生: %v", err)
		}

		if len(summaries) != 2 {
			t.Fatalf("プロジェクト数が異なる: got %d, want 2", len(summaries))
		}

		// タグ名があるプロジェクトが先、タグ名が空のプロジェクトが後
		if summaries[0].Project != "ProjectA" || summaries[0].TagName != "Backlog" {
			t.Errorf("1番目はProjectA(Backlog)であるべき: got %s(%s)", summaries[0].Project, summaries[0].TagName)
		}
		if summaries[1].Project != "ProjectB" || summaries[1].TagName != "" {
			t.Errorf("2番目はProjectB(空)であるべき: got %s(%s)", summaries[1].Project, summaries[1].TagName)
		}
	})
}

// TestProjectManager_NewAt は指定時刻でプロジェクトを開始するテスト
func TestProjectManager_NewAt(t *testing.T) {
	t.Run("指定時刻で新規プロジェクトを開始", func(t *testing.T) {
		// モックの準備
		currentStorage := &mockCurrentStorage{}
		logStorage := &mockLogStorage{}
		tagStorage := &mockTagStorage{}

		// 指定時刻
		specifiedTime := time.Date(2025, 9, 29, 14, 30, 0, 0, time.Local)

		// ProjectManagerの作成
		manager := NewProjectManager(currentStorage, logStorage, tagStorage)

		// NewAtメソッドのテスト
		err := manager.NewAt("ProjectA", "Development", specifiedTime)
		if err != nil {
			t.Errorf("NewAt()でエラーが発生: %v", err)
		}

		// currentStorageへの書き込みを確認
		if currentStorage.writeProject != "ProjectA" {
			t.Errorf("プロジェクト名が異なる: got %s, want ProjectA", currentStorage.writeProject)
		}
		if currentStorage.writeTag != "Development" {
			t.Errorf("タグが異なる: got %s, want Development", currentStorage.writeTag)
		}

		// logStorageへの追記を確認
		if len(logStorage.appendCalls) != 1 {
			t.Fatalf("ログ追記回数が異なる: got %d, want 1", len(logStorage.appendCalls))
		}

		call := logStorage.appendCalls[0]
		if call.project != "ProjectA" {
			t.Errorf("ログのプロジェクト名が異なる: got %s, want ProjectA", call.project)
		}
		if call.action != "start" {
			t.Errorf("ログのアクションが異なる: got %s, want start", call.action)
		}
		if call.tag != "Development" {
			t.Errorf("ログのタグが異なる: got %s, want Development", call.tag)
		}
		// 指定時刻が使用されていることを確認
		if !call.timestamp.Equal(specifiedTime) {
			t.Errorf("タイムスタンプが異なる: got %v, want %v", call.timestamp, specifiedTime)
		}
	})

	t.Run("既存プロジェクトがある場合は自動停止してから開始", func(t *testing.T) {
		// モックの準備
		currentStorage := &mockCurrentStorage{
			readResult: "OldProject\tOLD",
		}
		logStorage := &mockLogStorage{}
		tagStorage := &mockTagStorage{}

		// 指定時刻
		specifiedTime := time.Date(2025, 9, 29, 15, 0, 0, 0, time.Local)

		// ProjectManagerの作成
		manager := NewProjectManager(currentStorage, logStorage, tagStorage)

		// NewAtメソッドのテスト
		err := manager.NewAt("NewProject", "NEW", specifiedTime)
		if err != nil {
			t.Errorf("NewAt()でエラーが発生: %v", err)
		}

		// ログ追記が2回（既存停止と新規開始）
		if len(logStorage.appendCalls) != 2 {
			t.Fatalf("ログ追記回数が異なる: got %d, want 2", len(logStorage.appendCalls))
		}

		// 1回目: 既存プロジェクトの停止
		stopCall := logStorage.appendCalls[0]
		if stopCall.action != "stop" {
			t.Errorf("1回目のアクションが異なる: got %s, want stop", stopCall.action)
		}
		if !stopCall.timestamp.Equal(specifiedTime) {
			t.Errorf("停止時刻が異なる: got %v, want %v", stopCall.timestamp, specifiedTime)
		}

		// 2回目: 新規プロジェクトの開始
		startCall := logStorage.appendCalls[1]
		if startCall.action != "start" {
			t.Errorf("2回目のアクションが異なる: got %s, want start", startCall.action)
		}
		if !startCall.timestamp.Equal(specifiedTime) {
			t.Errorf("開始時刻が異なる: got %v, want %v", startCall.timestamp, specifiedTime)
		}
	})
}

// TestProjectManager_SwitchAt は指定時刻でプロジェクトを切り替えるテスト
func TestProjectManager_SwitchAt(t *testing.T) {
	t.Run("指定時刻でプロジェクトを切り替え", func(t *testing.T) {
		// モックの準備
		currentStorage := &mockCurrentStorage{
			readResult: "OldProject\tOLD",
		}
		logStorage := &mockLogStorage{}
		tagStorage := &mockTagStorage{}

		// 指定時刻
		specifiedTime := time.Date(2025, 9, 29, 16, 45, 0, 0, time.Local)

		// ProjectManagerの作成
		manager := NewProjectManager(currentStorage, logStorage, tagStorage)

		// SwitchAtメソッドのテスト
		err := manager.SwitchAt("NewProject", "NEW", specifiedTime)
		if err != nil {
			t.Errorf("SwitchAt()でエラーが発生: %v", err)
		}

		// currentStorageへの書き込みを確認
		if currentStorage.writeProject != "NewProject" {
			t.Errorf("プロジェクト名が異なる: got %s, want NewProject", currentStorage.writeProject)
		}

		// ログ追記が2回（既存停止と新規開始）
		if len(logStorage.appendCalls) != 2 {
			t.Fatalf("ログ追記回数が異なる: got %d, want 2", len(logStorage.appendCalls))
		}

		// 両方とも指定時刻が使用されていることを確認
		for i, call := range logStorage.appendCalls {
			if !call.timestamp.Equal(specifiedTime) {
				t.Errorf("ログ %d のタイムスタンプが異なる: got %v, want %v", i+1, call.timestamp, specifiedTime)
			}
		}
	})

	t.Run("稼働中プロジェクトがない状態から新規開始", func(t *testing.T) {
		// モックの準備（稼働中プロジェクトがない状態）
		currentStorage := &mockCurrentStorage{
			readError: errors.New("no current project"),
		}
		logStorage := &mockLogStorage{}
		tagStorage := &mockTagStorage{}

		// 指定時刻
		specifiedTime := time.Date(2025, 9, 29, 14, 0, 0, 0, time.Local)

		// ProjectManagerの作成
		manager := NewProjectManager(currentStorage, logStorage, tagStorage)

		// SwitchAtメソッドのテスト（稼働中なしから開始）
		err := manager.SwitchAt("ProjectA", "TAG1", specifiedTime)
		if err != nil {
			t.Errorf("SwitchAt()でエラーが発生: %v", err)
		}

		// currentStorageへの書き込みを確認
		if currentStorage.writeProject != "ProjectA" {
			t.Errorf("プロジェクト名が異なる: got %s, want ProjectA", currentStorage.writeProject)
		}
		if currentStorage.writeTag != "TAG1" {
			t.Errorf("タグが異なる: got %s, want TAG1", currentStorage.writeTag)
		}

		// ログ追記が1回のみ（新規開始のみ）
		if len(logStorage.appendCalls) != 1 {
			t.Fatalf("ログ追記回数が異なる: got %d, want 1", len(logStorage.appendCalls))
		}

		// 新規開始のログを確認
		call := logStorage.appendCalls[0]
		if call.action != "start" {
			t.Errorf("アクションが異なる: got %s, want start", call.action)
		}
		if call.project != "ProjectA" {
			t.Errorf("プロジェクト名が異なる: got %s, want ProjectA", call.project)
		}
		if !call.timestamp.Equal(specifiedTime) {
			t.Errorf("タイムスタンプが異なる: got %v, want %v", call.timestamp, specifiedTime)
		}
	})
}

// TestProjectManager_StopAt は指定時刻でプロジェクトを停止するテスト
func TestProjectManager_StopAt(t *testing.T) {
	t.Run("指定時刻でプロジェクトを停止", func(t *testing.T) {
		// モックの準備
		currentStorage := &mockCurrentStorage{
			readResult: "ProjectA\tDevelopment",
		}
		logStorage := &mockLogStorage{}
		tagStorage := &mockTagStorage{}

		// 指定時刻
		specifiedTime := time.Date(2025, 9, 29, 19, 30, 0, 0, time.Local)

		// ProjectManagerの作成
		manager := NewProjectManager(currentStorage, logStorage, tagStorage)

		// StopAtメソッドのテスト
		err := manager.StopAt(specifiedTime)
		if err != nil {
			t.Errorf("StopAt()でエラーが発生: %v", err)
		}

		// currentStorageがクリアされたことを確認
		if !currentStorage.cleared {
			t.Error("currentStorageがクリアされていない")
		}

		// logStorageへの追記を確認
		if len(logStorage.appendCalls) != 1 {
			t.Fatalf("ログ追記回数が異なる: got %d, want 1", len(logStorage.appendCalls))
		}

		call := logStorage.appendCalls[0]
		if call.action != "stop" {
			t.Errorf("ログのアクションが異なる: got %s, want stop", call.action)
		}
		// 指定時刻が使用されていることを確認
		if !call.timestamp.Equal(specifiedTime) {
			t.Errorf("タイムスタンプが異なる: got %v, want %v", call.timestamp, specifiedTime)
		}
	})
}

// TestProjectManager_ListOnDate は指定日のプロジェクト一覧を取得するテスト
func TestProjectManager_ListOnDate(t *testing.T) {
	t.Run("指定日のログエントリを読み込み", func(t *testing.T) {
		// モックの準備
		currentStorage := &mockCurrentStorage{
			readResult: "", // currentは考慮しない（過去の日付のため）
		}

		// 特定の日付
		targetDate := time.Date(2025, 10, 5, 0, 0, 0, 0, time.Local)

		// 指定日のログエントリを準備
		baseTime := time.Date(2025, 10, 5, 10, 0, 0, 0, time.Local)
		logStorage := &mockLogStorage{
			readAllOnDateResult: map[string][]storage.LogEntry{
				"ProjectA": {
					{Action: "start", Tag: "Development", Timestamp: baseTime},
					{Action: "stop", Tag: "Development", Timestamp: baseTime.Add(2 * time.Hour)},
				},
				"ProjectB": {
					{Action: "start", Tag: "MTG", Timestamp: baseTime.Add(3 * time.Hour)},
					{Action: "stop", Tag: "MTG", Timestamp: baseTime.Add(4 * time.Hour)},
				},
			},
		}
		tagStorage := &mockTagStorage{}

		// ProjectManagerの作成
		manager := NewProjectManager(currentStorage, logStorage, tagStorage)

		// ListOnDateメソッドのテスト
		summaries, err := manager.ListOnDate(targetDate)
		if err != nil {
			t.Fatalf("ListOnDate()でエラーが発生: %v", err)
		}

		// プロジェクト数の確認
		if len(summaries) != 2 {
			t.Fatalf("プロジェクト数が異なる: got %d, want 2", len(summaries))
		}

		// ProjectAの稼働時間を確認（2時間）
		projectA := findSummary(summaries, "ProjectA")
		if projectA == nil {
			t.Error("ProjectAが見つからない")
		} else if projectA.TotalTime != 2*time.Hour {
			t.Errorf("ProjectAの稼働時間が異なる: got %v, want %v", projectA.TotalTime, 2*time.Hour)
		}

		// ProjectBの稼働時間を確認（1時間）
		projectB := findSummary(summaries, "ProjectB")
		if projectB == nil {
			t.Error("ProjectBが見つからない")
		} else if projectB.TotalTime != 1*time.Hour {
			t.Errorf("ProjectBの稼働時間が異なる: got %v, want %v", projectB.TotalTime, 1*time.Hour)
		}
	})
}

// TestProjectManager_ListOnDate_SortedByTagName は指定日のリストがタグ名でソートされることをテストする
func TestProjectManager_ListOnDate_SortedByTagName(t *testing.T) {
	t.Run("タグ名でグループ化され、同一タグ内では最終アクティビティ降順でソート", func(t *testing.T) {
		// モックの準備
		currentStorage := &mockCurrentStorage{
			readResult: "", // currentは考慮しない
		}

		// 特定の日付
		targetDate := time.Date(2025, 10, 5, 0, 0, 0, 0, time.Local)
		baseTime := time.Date(2025, 10, 5, 10, 0, 0, 0, time.Local)

		logStorage := &mockLogStorage{
			readAllOnDateResult: map[string][]storage.LogEntry{
				// Backlogタグのプロジェクト
				"Project1": {
					{Action: "start", Tag: "1", Timestamp: baseTime},
					{Action: "stop", Tag: "1", Timestamp: baseTime.Add(1 * time.Hour)},
				},
				"Project2": {
					{Action: "start", Tag: "1", Timestamp: baseTime.Add(4 * time.Hour)},
					{Action: "stop", Tag: "1", Timestamp: baseTime.Add(5 * time.Hour)},
				},
				// Learnタグのプロジェクト
				"Project3": {
					{Action: "start", Tag: "2", Timestamp: baseTime.Add(2 * time.Hour)},
					{Action: "stop", Tag: "2", Timestamp: baseTime.Add(3 * time.Hour)},
				},
			},
		}
		tagStorage := &mockTagStorage{
			tags: []storage.Tag{
				{ID: 1, Name: "Backlog"},
				{ID: 2, Name: "Learn"},
			},
		}

		manager := NewProjectManager(currentStorage, logStorage, tagStorage)

		summaries, err := manager.ListOnDate(targetDate)
		if err != nil {
			t.Fatalf("ListOnDate()でエラーが発生: %v", err)
		}

		if len(summaries) != 3 {
			t.Fatalf("プロジェクト数が異なる: got %d, want 3", len(summaries))
		}

		// タグ名でグループ化されていることを確認
		expectedOrder := []struct {
			project string
			tagName string
		}{
			{"Project2", "Backlog"}, // Backlogグループ（最終アクティビティ降順）
			{"Project1", "Backlog"},
			{"Project3", "Learn"}, // Learnグループ
		}

		for i, expected := range expectedOrder {
			if summaries[i].Project != expected.project {
				t.Errorf("位置%dのプロジェクトが異なる: got %s, want %s", i, summaries[i].Project, expected.project)
			}
			if summaries[i].TagName != expected.tagName {
				t.Errorf("位置%dのタグ名が異なる: got %s, want %s", i, summaries[i].TagName, expected.tagName)
			}
		}
	})
}

func TestProjectManager_ListRecent(t *testing.T) {
	t.Run("過去2週間のプロジェクトを取得", func(t *testing.T) {
		// モックの準備
		currentStorage := &mockCurrentStorage{}

		// 3日間のログデータを準備
		day1 := time.Date(2025, 10, 1, 10, 0, 0, 0, time.Local)
		day2 := time.Date(2025, 10, 2, 10, 0, 0, 0, time.Local)
		day3 := time.Date(2025, 10, 3, 10, 0, 0, 0, time.Local)

		logStorage := &mockLogStorage{
			readRangeResult: map[string][]storage.LogEntry{
				"ProjectA": {
					{Timestamp: day1, Action: "start", Tag: "Development"},
					{Timestamp: day1.Add(1 * time.Hour), Action: "stop", Tag: "Development"},
					{Timestamp: day3, Action: "start", Tag: "Development"}, // Day3にも作業
					{Timestamp: day3.Add(1 * time.Hour), Action: "stop", Tag: "Development"},
				},
				"ProjectB": {
					{Timestamp: day2, Action: "start", Tag: "MTG"},
					{Timestamp: day2.Add(1 * time.Hour), Action: "stop", Tag: "MTG"},
				},
				"ProjectC": {
					{Timestamp: day1, Action: "start", Tag: "REV"},
					{Timestamp: day1.Add(30 * time.Minute), Action: "stop", Tag: "REV"},
				},
			},
		}

		tagStorage := &mockTagStorage{
			tags: []storage.Tag{
				{ID: 1, Name: "開発"},
				{ID: 2, Name: "会議"},
				{ID: 3, Name: "レビュー"},
			},
		}

		manager := NewProjectManager(currentStorage, logStorage, tagStorage)

		// ListRecentで過去14日間のプロジェクトを取得
		summaries, err := manager.ListRecent(14)
		if err != nil {
			t.Fatalf("ListRecent()でエラーが発生: %v", err)
		}

		// プロジェクト数を確認
		if len(summaries) != 3 {
			t.Fatalf("プロジェクト数が異なる: got %d, want 3", len(summaries))
		}

		// 最終作業日でソートされていることを確認（新しい順）
		if summaries[0].Project != "ProjectA" {
			t.Errorf("1番目のプロジェクトが異なる: got %s, want ProjectA", summaries[0].Project)
		}
		if summaries[1].Project != "ProjectB" {
			t.Errorf("2番目のプロジェクトが異なる: got %s, want ProjectB", summaries[1].Project)
		}
		if summaries[2].Project != "ProjectC" {
			t.Errorf("3番目のプロジェクトが異なる: got %s, want ProjectC", summaries[2].Project)
		}

		// 各プロジェクトの最終作業日を確認
		if !summaries[0].LastActivity.Equal(day3.Add(1 * time.Hour)) {
			t.Errorf("ProjectAの最終作業日が異なる: got %v, want %v", summaries[0].LastActivity, day3.Add(1*time.Hour))
		}
		if !summaries[1].LastActivity.Equal(day2.Add(1 * time.Hour)) {
			t.Errorf("ProjectBの最終作業日が異なる: got %v, want %v", summaries[1].LastActivity, day2.Add(1*time.Hour))
		}
		if !summaries[2].LastActivity.Equal(day1.Add(30 * time.Minute)) {
			t.Errorf("ProjectCの最終作業日が異なる: got %v, want %v", summaries[2].LastActivity, day1.Add(30*time.Minute))
		}
	})

	t.Run("重複プロジェクトを除外", func(t *testing.T) {
		// モックの準備
		currentStorage := &mockCurrentStorage{}

		day1 := time.Date(2025, 10, 1, 10, 0, 0, 0, time.Local)
		day2 := time.Date(2025, 10, 2, 10, 0, 0, 0, time.Local)

		// 同じプロジェクト名+タグIDの組み合わせが複数日にある場合
		logStorage := &mockLogStorage{
			readRangeResult: map[string][]storage.LogEntry{
				"ProjectA": {
					{Timestamp: day1, Action: "start", Tag: "Development"},
					{Timestamp: day1.Add(1 * time.Hour), Action: "stop", Tag: "Development"},
					{Timestamp: day2, Action: "start", Tag: "Development"},
					{Timestamp: day2.Add(1 * time.Hour), Action: "stop", Tag: "Development"},
				},
			},
		}

		tagStorage := &mockTagStorage{
			tags: []storage.Tag{
				{ID: 1, Name: "開発"},
			},
		}

		manager := NewProjectManager(currentStorage, logStorage, tagStorage)

		summaries, err := manager.ListRecent(14)
		if err != nil {
			t.Fatalf("ListRecent()でエラーが発生: %v", err)
		}

		// 重複排除されて1つのProjectAのみになることを確認
		if len(summaries) != 1 {
			t.Fatalf("プロジェクト数が異なる: got %d, want 1", len(summaries))
		}

		// 最新の作業日が採用されることを確認
		if !summaries[0].LastActivity.Equal(day2.Add(1 * time.Hour)) {
			t.Errorf("最終作業日が異なる: got %v, want %v", summaries[0].LastActivity, day2.Add(1*time.Hour))
		}

		// 稼働時間の合計を確認（2時間）
		if summaries[0].TotalTime != 2*time.Hour {
			t.Errorf("稼働時間が異なる: got %v, want %v", summaries[0].TotalTime, 2*time.Hour)
		}
	})

	t.Run("データが存在しない場合", func(t *testing.T) {
		// モックの準備
		currentStorage := &mockCurrentStorage{}
		logStorage := &mockLogStorage{
			readRangeResult: make(map[string][]storage.LogEntry),
		}
		tagStorage := &mockTagStorage{
			tags: []storage.Tag{},
		}

		manager := NewProjectManager(currentStorage, logStorage, tagStorage)

		summaries, err := manager.ListRecent(14)
		if err != nil {
			t.Fatalf("ListRecent()でエラーが発生: %v", err)
		}

		// 空のスライスが返されることを確認
		if len(summaries) != 0 {
			t.Errorf("空のスライスが期待されるが、%d個のプロジェクトが返された", len(summaries))
		}
	})
}

func TestProjectManager_GetTags(t *testing.T) {
	t.Run("タグ一覧を取得できる", func(t *testing.T) {
		currentStorage := &mockCurrentStorage{}
		logStorage := &mockLogStorage{}
		tagStorage := &mockTagStorage{
			tags: []storage.Tag{
				{ID: 1, Name: "開発"},
				{ID: 2, Name: "会議"},
				{ID: 3, Name: "レビュー"},
			},
		}

		manager := NewProjectManager(currentStorage, logStorage, tagStorage)

		tags, err := manager.GetTags()
		if err != nil {
			t.Fatalf("GetTags()でエラーが発生: %v", err)
		}

		if len(tags) != 3 {
			t.Errorf("期待: 3タグ, 実際: %d", len(tags))
		}

		if tags[0].ID != 1 || tags[0].Name != "開発" {
			t.Errorf("最初のタグが不正: %+v", tags[0])
		}
	})

	t.Run("空のタグ一覧", func(t *testing.T) {
		currentStorage := &mockCurrentStorage{}
		logStorage := &mockLogStorage{}
		tagStorage := &mockTagStorage{
			tags: []storage.Tag{},
		}

		manager := NewProjectManager(currentStorage, logStorage, tagStorage)

		tags, err := manager.GetTags()
		if err != nil {
			t.Fatalf("GetTags()でエラーが発生: %v", err)
		}

		if len(tags) != 0 {
			t.Errorf("期待: 0タグ, 実際: %d", len(tags))
		}
	})

	t.Run("タグ読み込みエラー", func(t *testing.T) {
		currentStorage := &mockCurrentStorage{}
		logStorage := &mockLogStorage{}
		tagStorage := &mockTagStorage{
			loadError: errors.New("読み込みエラー"),
		}

		manager := NewProjectManager(currentStorage, logStorage, tagStorage)

		_, err := manager.GetTags()
		if err == nil {
			t.Error("エラーが期待されるが、nilが返された")
		}
	})
}

func TestProjectManager_AddTag(t *testing.T) {
	t.Run("新しいタグを追加", func(t *testing.T) {
		currentStorage := &mockCurrentStorage{}
		logStorage := &mockLogStorage{}
		tagStorage := &mockTagStorage{
			tags: []storage.Tag{
				{ID: 1, Name: "開発"},
				{ID: 2, Name: "会議"},
			},
		}

		manager := NewProjectManager(currentStorage, logStorage, tagStorage)

		// 新しいタグを追加
		newTag, err := manager.AddTag("レビュー")
		if err != nil {
			t.Fatalf("AddTag()でエラーが発生: %v", err)
		}

		// IDが自動採番されているか確認
		if newTag.ID != 3 {
			t.Errorf("newTag.ID = %d, want 3", newTag.ID)
		}

		// 名前が正しいか確認
		if newTag.Name != "レビュー" {
			t.Errorf("newTag.Name = %q, want %q", newTag.Name, "レビュー")
		}

		// タグが追加されたことを確認
		tags, err := manager.GetTags()
		if err != nil {
			t.Fatalf("GetTags()でエラーが発生: %v", err)
		}

		if len(tags) != 3 {
			t.Errorf("タグ数が異なる: got %d, want 3", len(tags))
		}
	})

	t.Run("同名タグが既に存在する場合はエラー", func(t *testing.T) {
		currentStorage := &mockCurrentStorage{}
		logStorage := &mockLogStorage{}
		tagStorage := &mockTagStorage{
			tags: []storage.Tag{
				{ID: 1, Name: "開発"},
				{ID: 2, Name: "会議"},
			},
		}

		manager := NewProjectManager(currentStorage, logStorage, tagStorage)

		// 既に存在するタグ名で追加を試みる
		_, err := manager.AddTag("開発")
		if err == nil {
			t.Error("同名タグの追加でエラーが発生しなかった")
		}
	})
}

func TestProjectManager_DeleteTag(t *testing.T) {
	t.Run("指定したIDのタグを削除", func(t *testing.T) {
		currentStorage := &mockCurrentStorage{}
		logStorage := &mockLogStorage{}
		tagStorage := &mockTagStorage{
			tags: []storage.Tag{
				{ID: 1, Name: "開発"},
				{ID: 2, Name: "会議"},
				{ID: 3, Name: "レビュー"},
			},
		}

		manager := NewProjectManager(currentStorage, logStorage, tagStorage)

		// ID=2のタグを削除
		err := manager.DeleteTag(2)
		if err != nil {
			t.Fatalf("DeleteTag()でエラーが発生: %v", err)
		}

		// タグが削除されたことを確認
		tags, err := manager.GetTags()
		if err != nil {
			t.Fatalf("GetTags()でエラーが発生: %v", err)
		}

		if len(tags) != 2 {
			t.Errorf("タグ数が異なる: got %d, want 2", len(tags))
		}

		// 削除されたタグが含まれていないことを確認
		for _, tag := range tags {
			if tag.ID == 2 {
				t.Error("削除されたはずのタグが含まれている")
			}
		}
	})

	t.Run("存在しないIDを指定した場合はエラー", func(t *testing.T) {
		currentStorage := &mockCurrentStorage{}
		logStorage := &mockLogStorage{}
		tagStorage := &mockTagStorage{
			tags: []storage.Tag{
				{ID: 1, Name: "開発"},
				{ID: 2, Name: "会議"},
			},
		}

		manager := NewProjectManager(currentStorage, logStorage, tagStorage)

		// 存在しないID=99のタグを削除しようとする
		err := manager.DeleteTag(99)
		if err == nil {
			t.Error("存在しないIDの削除でエラーが発生しなかった")
		}
	})
}
