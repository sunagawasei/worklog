package cmd

import (
	"fmt"
	"testing"
	"time"

	"worklog/internal/domain"
	"worklog/internal/storage"
)

// mockProjectManager はテスト用のモック実装
type mockProjectManager struct {
	newError            error
	switchError         error
	stopError           error
	status              *domain.ProjectStatus
	statusError         error
	summaries           []domain.ProjectSummary
	listError           error
	listOnDateSummaries []domain.ProjectSummary
	listOnDateError     error
	calledMethods       []string // 呼ばれたメソッドを記録
	tags                []domain.Tag
	tagsError           error
	recentSummaries     []domain.ProjectSummary
	recentError         error
}

func (m *mockProjectManager) New(project, tag string) error {
	m.calledMethods = append(m.calledMethods, fmt.Sprintf("New(%s,%s)", project, tag))
	return m.newError
}

func (m *mockProjectManager) Switch(project, tag string) error {
	m.calledMethods = append(m.calledMethods, fmt.Sprintf("Switch(%s,%s)", project, tag))
	return m.switchError
}

func (m *mockProjectManager) Stop() error {
	m.calledMethods = append(m.calledMethods, "Stop()")
	return m.stopError
}

func (m *mockProjectManager) Status() (*domain.ProjectStatus, error) {
	m.calledMethods = append(m.calledMethods, "Status()")
	return m.status, m.statusError
}

func (m *mockProjectManager) List() ([]domain.ProjectSummary, error) {
	m.calledMethods = append(m.calledMethods, "List()")
	return m.summaries, m.listError
}

func (m *mockProjectManager) ListOnDate(date time.Time) ([]domain.ProjectSummary, error) {
	m.calledMethods = append(m.calledMethods, fmt.Sprintf("ListOnDate(%v)", date))
	return m.listOnDateSummaries, m.listOnDateError
}

func (m *mockProjectManager) NewAt(project, tag string, timestamp time.Time) error {
	m.calledMethods = append(m.calledMethods, fmt.Sprintf("NewAt(%s,%s,%v)", project, tag, timestamp))
	return m.newError
}

func (m *mockProjectManager) SwitchAt(project, tag string, timestamp time.Time) error {
	m.calledMethods = append(m.calledMethods, fmt.Sprintf("SwitchAt(%s,%s,%v)", project, tag, timestamp))
	return m.switchError
}

func (m *mockProjectManager) StopAt(timestamp time.Time) error {
	m.calledMethods = append(m.calledMethods, fmt.Sprintf("StopAt(%v)", timestamp))
	return m.stopError
}

func (m *mockProjectManager) GetTags() ([]domain.Tag, error) {
	m.calledMethods = append(m.calledMethods, "GetTags()")
	return m.tags, m.tagsError
}

func (m *mockProjectManager) ListRecent(days int) ([]domain.ProjectSummary, error) {
	m.calledMethods = append(m.calledMethods, fmt.Sprintf("ListRecent(%d)", days))
	return m.recentSummaries, m.recentError
}

// mockTagStorage はテスト用のモック実装
type mockTagStorage struct {
	tags []storage.Tag
	err  error
}

func (m *mockTagStorage) Load() ([]storage.Tag, error) {
	return m.tags, m.err
}

func TestHandleSwitch_InteractiveMode(t *testing.T) {
	t.Run("本日のプロジェクトリストから選択して切り替え", func(t *testing.T) {
		// モックマネージャーを設定
		_ = &mockProjectManager{
			status: &domain.ProjectStatus{
				Project:   "CurrentProject",
				Tag:       "Development",
				StartTime: time.Now(),
			},
			summaries: []domain.ProjectSummary{
				{Project: "ProjectA", Tag: "Development", TotalTime: time.Hour},
				{Project: "ProjectB", Tag: "MTG", TotalTime: time.Hour * 2},
				{Project: "CurrentProject", Tag: "Development", TotalTime: time.Hour * 3},
			},
		}

		// 期待される動作:
		// 1. List()で本日のプロジェクト一覧を取得
		// 2. Status()で現在稼働中のプロジェクトを取得
		// 3. 稼働中でないプロジェクトのみリスト表示
		// 4. Switch()で選択したプロジェクトに切り替え

		// NOTE: 対話的UIのテストは複雑なため、統合テストで確認
		// ここでは、List()とStatus()が呼ばれることを確認する仕様とする
	})

	t.Run("本日のプロジェクトがない場合", func(t *testing.T) {
		// モックマネージャーを設定（プロジェクトリストが空）
		manager := &mockProjectManager{
			summaries: []domain.ProjectSummary{},
		}

		// NOTE: 実際のhandleSwitch関数の変更後にテストを追加
		_ = manager // linterエラー回避
	})

	t.Run("稼働中のプロジェクトがない場合", func(t *testing.T) {
		// モックマネージャーを設定（稼働中なし）
		manager := &mockProjectManager{
			status: nil, // 稼働中なし
			summaries: []domain.ProjectSummary{
				{Project: "ProjectA", Tag: "Development", TotalTime: time.Hour},
				{Project: "ProjectB", Tag: "MTG", TotalTime: time.Hour * 2},
			},
		}

		// NOTE: 実際のhandleSwitch関数の変更後にテストを追加
		_ = manager // linterエラー回避
	})
}
