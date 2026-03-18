package ui

import (
	"errors"
	"testing"
	"worklog/internal/domain"
)

// テスト用のエラー定義
var ErrNoItems = errors.New("利用可能な項目がありません")

// モックPromptUI実装（テスト用）
type mockPromptUI struct {
	selectedTag          string
	selectTagError       error
	inputProject         string
	inputError           error
	confirmResult        bool
	confirmError         error
	selectedProjectIndex int
	selectProjectError   error
	inputTime            string
	inputTimeError       error
}

func (m *mockPromptUI) SelectTag(tags []domain.Tag) (string, error) {
	return m.selectedTag, m.selectTagError
}

func (m *mockPromptUI) InputTime(label string) (string, error) {
	return m.inputTime, m.inputTimeError
}

func (m *mockPromptUI) InputProject() (string, error) {
	return m.inputProject, m.inputError
}

func (m *mockPromptUI) ConfirmAction(action string) (bool, error) {
	return m.confirmResult, m.confirmError
}

func (m *mockPromptUI) SelectProjectFromList(projects []ProjectDisplay) (*ProjectDisplay, error) {
	if m.selectProjectError != nil {
		return nil, m.selectProjectError
	}
	if len(projects) == 0 {
		return nil, ErrNoItems
	}
	if m.selectedProjectIndex >= len(projects) || m.selectedProjectIndex < 0 {
		return &projects[0], nil
	}

	// セパレーターが選択された場合はエラーを返す
	selected := &projects[m.selectedProjectIndex]
	if selected.IsSeparator {
		return nil, errors.New("セパレーターは選択できません")
	}

	return selected, nil
}

func TestPromptUI_SelectTag_EmptyTags(t *testing.T) {
	// 空のタグリストでエラーが返ることを確認
	ui := NewPromptUI()
	_, err := ui.SelectTag([]domain.Tag{})
	if err == nil {
		t.Error("空のタグリストでエラーが返るべき")
	}
}

func TestMockPromptUI_SelectTag(t *testing.T) {
	mockUI := &mockPromptUI{
		selectedTag: "1",
	}

	tags := []domain.Tag{
		{ID: 1, Name: "開発"},
		{ID: 2, Name: "会議"},
	}

	result, err := mockUI.SelectTag(tags)
	if err != nil {
		t.Errorf("SelectTagでエラーが発生: %v", err)
	}

	if result != "1" {
		t.Errorf("期待値と異なる: got %s, want 1", result)
	}
}

func TestMockPromptUI_InputProject(t *testing.T) {
	mockUI := &mockPromptUI{
		inputProject: "テストプロジェクト",
	}

	result, err := mockUI.InputProject()
	if err != nil {
		t.Errorf("InputProjectでエラーが発生: %v", err)
	}

	if result != "テストプロジェクト" {
		t.Errorf("期待値と異なる: got %s, want テストプロジェクト", result)
	}
}

func TestMockPromptUI_ConfirmAction(t *testing.T) {
	t.Run("確認がtrueの場合", func(t *testing.T) {
		mockUI := &mockPromptUI{
			confirmResult: true,
		}

		result, err := mockUI.ConfirmAction("プロジェクトを終了")
		if err != nil {
			t.Errorf("ConfirmActionでエラーが発生: %v", err)
		}

		if !result {
			t.Error("確認結果がtrueであるべき")
		}
	})

	t.Run("確認がfalseの場合", func(t *testing.T) {
		mockUI := &mockPromptUI{
			confirmResult: false,
		}

		result, err := mockUI.ConfirmAction("プロジェクトを終了")
		if err != nil {
			t.Errorf("ConfirmActionでエラーが発生: %v", err)
		}

		if result {
			t.Error("確認結果がfalseであるべき")
		}
	})
}

func TestMockPromptUI_SelectProjectFromList(t *testing.T) {
	t.Run("プロジェクトリストから選択できる", func(t *testing.T) {
		mockUI := &mockPromptUI{
			selectedProjectIndex: 1,
		}

		projects := []ProjectDisplay{
			{Project: "ProjectA", Tag: "Development", Time: "1時間30分"},
			{Project: "ProjectB", Tag: "MTG", Time: "2時間00分"},
			{Project: "ProjectC", Tag: "REV", Time: "0時間45分"},
		}

		selected, err := mockUI.SelectProjectFromList(projects)
		if err != nil {
			t.Errorf("エラーが発生しました: %v", err)
		}
		if selected == nil {
			t.Error("選択されたプロジェクトがnilです")
			return
		}
		if selected.Project != "ProjectB" {
			t.Errorf("期待されるプロジェクト名: ProjectB, 実際: %s", selected.Project)
		}
		if selected.Tag != "MTG" {
			t.Errorf("期待されるタグ: MTG, 実際: %s", selected.Tag)
		}
	})

	t.Run("空のリストの場合エラーを返す", func(t *testing.T) {
		mockUI := &mockPromptUI{}
		projects := []ProjectDisplay{}

		_, err := mockUI.SelectProjectFromList(projects)
		if err == nil {
			t.Error("空のリストでエラーが返されませんでした")
		}
		if err != ErrNoItems {
			t.Errorf("期待されるエラー: %v, 実際: %v", ErrNoItems, err)
		}
	})

	t.Run("インデックスが範囲外の場合最初の項目を返す", func(t *testing.T) {
		mockUI := &mockPromptUI{
			selectedProjectIndex: 10, // 範囲外
		}

		projects := []ProjectDisplay{
			{Project: "ProjectA", Tag: "Development", Time: "1時間30分"},
			{Project: "ProjectB", Tag: "MTG", Time: "2時間00分"},
		}

		selected, err := mockUI.SelectProjectFromList(projects)
		if err != nil {
			t.Errorf("エラーが発生しました: %v", err)
		}
		if selected.Project != "ProjectA" {
			t.Errorf("最初の項目が選択されるべき: got %s, want ProjectA", selected.Project)
		}
	})
}

func TestFormatTagID(t *testing.T) {
	t.Run("1桁の数字は前にスペースが追加される", func(t *testing.T) {
		result := formatTagID(5)
		if result != " 5" {
			t.Errorf("期待値: ' 5', 実際: '%s'", result)
		}
	})

	t.Run("2桁の数字はそのまま表示される", func(t *testing.T) {
		result := formatTagID(12)
		if result != "12" {
			t.Errorf("期待値: '12', 実際: '%s'", result)
		}
	})

	t.Run("3桁の数字もそのまま表示される", func(t *testing.T) {
		result := formatTagID(100)
		if result != "100" {
			t.Errorf("期待値: '100', 実際: '%s'", result)
		}
	})
}

func TestProjectDisplay_SeparatorFields(t *testing.T) {
	t.Run("通常のプロジェクトはセパレーターではない", func(t *testing.T) {
		pd := ProjectDisplay{
			Project:     "TestProject",
			Tag:         "Development",
			Time:        "1時間30分",
			Status:      "▫ paused",
			IsRunning:   false,
			DateLabel:   "[Today]",
			IsSeparator: false,
		}

		if pd.IsSeparator {
			t.Error("通常のプロジェクトでIsSeparatorがtrueになっています")
		}
	})

	t.Run("セパレーターアイテムの作成", func(t *testing.T) {
		separator := ProjectDisplay{
			IsSeparator:   true,
			SeparatorText: "────────────────────────────────",
		}

		if !separator.IsSeparator {
			t.Error("セパレーターアイテムでIsSeparatorがfalseです")
		}
		if separator.SeparatorText == "" {
			t.Error("セパレーターテキストが空です")
		}
	})
}

func TestMockPromptUI_SelectProjectFromList_WithSeparator(t *testing.T) {
	t.Run("セパレーターが選択された場合はエラーを返す", func(t *testing.T) {
		mockUI := &mockPromptUI{
			selectedProjectIndex: 2, // セパレーターを選択
		}

		projects := []ProjectDisplay{
			{Project: "ProjectA", Tag: "Development", Time: "1時間30分", DateLabel: "[Today]"},
			{Project: "ProjectB", Tag: "MTG", Time: "2時間00分", DateLabel: "[Today]"},
			{IsSeparator: true, SeparatorText: "────────────────────────────────"},
			{Project: "ProjectC", Tag: "REV", Time: "0時間45分", DateLabel: "[2 days ago]"},
		}

		selected, err := mockUI.SelectProjectFromList(projects)

		// セパレーターが選択された場合、エラーが返されるべき
		if err == nil {
			t.Error("セパレーター選択時にエラーが返されるべきです")
		}
		if selected != nil {
			t.Error("セパレーター選択時はnilを返すべきです")
		}
	})

	t.Run("セパレーター以外の項目は正常に選択できる", func(t *testing.T) {
		mockUI := &mockPromptUI{
			selectedProjectIndex: 3, // セパレーター後のプロジェクト
		}

		projects := []ProjectDisplay{
			{Project: "ProjectA", Tag: "Development", Time: "1時間30分", DateLabel: "[Today]"},
			{Project: "ProjectB", Tag: "MTG", Time: "2時間00分", DateLabel: "[Today]"},
			{IsSeparator: true, SeparatorText: "────────────────────────────────"},
			{Project: "ProjectC", Tag: "REV", Time: "0時間45分", DateLabel: "[2 days ago]"},
		}

		selected, err := mockUI.SelectProjectFromList(projects)
		if err != nil {
			t.Errorf("エラーが発生: %v", err)
		}
		if selected == nil {
			t.Error("選択されたプロジェクトがnilです")
			return
		}
		if selected.Project != "ProjectC" {
			t.Errorf("期待値: ProjectC, 実際: %s", selected.Project)
		}
	})
}

func TestTruncateProjectName(t *testing.T) {
	t.Run("プロジェクト名が最大幅以内の場合はそのまま返す", func(t *testing.T) {
		result := truncateProjectName("短い名前", 80, 0)
		if result != "短い名前" {
			t.Errorf("期待値: '短い名前', 実際: '%s'", result)
		}
	})

	t.Run("プロジェクト名が最大幅を超える場合は切り詰める", func(t *testing.T) {
		longName := "とても長いプロジェクト名でターミナル幅を超えてしまうもの"
		result := truncateProjectName(longName, 80, 0)

		// 80文字幅の場合、固定要素で46文字使うので、34文字に切り詰められる
		// 全角文字は2文字幅なので、17文字程度に切り詰められる
		if len(result) >= len(longName) {
			t.Errorf("プロジェクト名が切り詰められていません: '%s'", result)
		}

		// 省略記号が含まれることを確認
		if result[len(result)-3:] != "…" {
			t.Errorf("省略記号が含まれていません: '%s'", result)
		}
	})

	t.Run("ターミナル幅75文字の場合に適切に切り詰める", func(t *testing.T) {
		longName := "TASK-003 サービスのカラム追加に伴うbilling-tool修正"
		result := truncateProjectName(longName, 75, 0)

		// 切り詰められていることを確認
		if len(result) >= len(longName) {
			t.Errorf("プロジェクト名が切り詰められていません")
		}

		// 省略記号が含まれることを確認
		if result[len(result)-3:] != "…" {
			t.Errorf("省略記号が含まれていません: '%s'", result)
		}
	})

	t.Run("ターミナル幅が非常に小さい場合でも最低幅を確保", func(t *testing.T) {
		result := truncateProjectName("プロジェクト名", 30, 0)

		// 最低でも10文字は表示されるべき
		if displayWidth(result) < 10 {
			t.Errorf("最低幅が確保されていません: '%s' (width: %d)", result, displayWidth(result))
		}
	})

	t.Run("タグ名の幅を考慮してプロジェクト名を切り詰める", func(t *testing.T) {
		longName := "とても長いプロジェクト名でターミナル幅を超えてしまうもの"
		// タグ名「開発」(幅4) + スペース2 = 6文字分余分に削る
		resultWithTag := truncateProjectName(longName, 80, 4)
		resultWithout := truncateProjectName(longName, 80, 0)

		// タグあり版の方がより短く切り詰められているはず
		if len(resultWithTag) >= len(resultWithout) {
			t.Errorf("タグ幅が考慮されていません: withTag='%s', without='%s'", resultWithTag, resultWithout)
		}
	})
}

func TestGetTerminalWidth(t *testing.T) {
	t.Run("ターミナル幅が取得できる場合", func(t *testing.T) {
		width := GetTerminalWidth()

		// テスト環境によってはターミナルがない可能性があるので、
		// デフォルト値（80）が返されるか、正の値が返されることを確認
		if width <= 0 {
			t.Errorf("無効なターミナル幅: %d", width)
		}
	})
}
