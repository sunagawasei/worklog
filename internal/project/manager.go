// Package project はワークログプロジェクトの管理ビジネスロジックを担当する
// 稼働時間の記録、プロジェクト切り替え、状態レポート機能を提供
package project

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"worklog/internal/domain"
	"worklog/internal/storage"
)

const (
	actionStart = "start"
	actionStop  = "stop"
)

// ProjectManager はプロジェクト管理のビジネスロジックを扱うインターフェース
type ProjectManager interface {
	New(project, tag string) error                              // 新規プロジェクト開始
	Switch(project, tag string) error                           // プロジェクト切り替え
	Stop() error                                                // 停止
	Status() (*domain.ProjectStatus, error)                     // 現在の状況
	List() ([]domain.ProjectSummary, error)                     // 本日の一覧
	ListOnDate(date time.Time) ([]domain.ProjectSummary, error) // 指定日の一覧
	ListRecent(days int) ([]domain.ProjectSummary, error)       // 過去N日間の一覧
	GetTags() ([]domain.Tag, error)                             // タグ一覧を取得

	// 時刻指定版メソッド
	NewAt(project, tag string, timestamp time.Time) error    // 指定時刻で新規プロジェクト開始
	SwitchAt(project, tag string, timestamp time.Time) error // 指定時刻でプロジェクト切り替え
	StopAt(timestamp time.Time) error                        // 指定時刻で停止
}

// projectManager はProjectManagerインターフェースの実装
type projectManager struct {
	currentStorage storage.CurrentStorage
	logStorage     storage.LogStorage
	tagStorage     storage.TagStorage
}

// NewProjectManager は新しいProjectManagerインスタンスを作成する
func NewProjectManager(
	current storage.CurrentStorage,
	log storage.LogStorage,
	tag storage.TagStorage,
) ProjectManager {
	return &projectManager{
		currentStorage: current,
		logStorage:     log,
		tagStorage:     tag,
	}
}

// New は新規プロジェクトを開始する
func (m *projectManager) New(project, tag string) error {
	return m.NewAt(project, tag, time.Now())
}

// Switch はプロジェクトを切り替える
func (m *projectManager) Switch(project, tag string) error {
	return m.SwitchAt(project, tag, time.Now())
}

// Stop は現在のプロジェクトを停止する
func (m *projectManager) Stop() error {
	return m.StopAt(time.Now())
}

// getTagName はタグIDからタグ名を取得するヘルパーメソッド
func (m *projectManager) getTagName(tagID string) string {
	tags, err := m.tagStorage.Load()
	if err != nil {
		return ""
	}

	// タグIDを数値に変換
	id := 0
	fmt.Sscanf(tagID, "%d", &id)

	// 対応するタグ名を検索
	for _, t := range tags {
		if t.ID == id {
			return t.Name
		}
	}

	return ""
}

// Status は現在の稼働状況を返す（簡潔な名前）
func (m *projectManager) Status() (*domain.ProjectStatus, error) {
	// 既存のプロジェクト情報を取得
	current, err := m.currentStorage.Read()
	if err != nil {
		// 稼働中のプロジェクトがない場合はnilを返す（エラーではない）
		return nil, nil
	}

	// current文字列を解析（format: "ProjectName\tTagID"）
	var project, tag string
	if idx := strings.IndexRune(current, '\t'); idx != -1 {
		project = current[:idx]
		tag = current[idx+1:]
	} else {
		return nil, errors.New("current情報の形式が不正です")
	}

	// タグ名を取得
	tagName := m.getTagName(tag)

	// ログから開始時刻と累計時間を取得
	startTime := time.Now()
	totalTime := time.Duration(0)
	currentSessionTime := time.Duration(0)

	logs, err := m.logStorage.ReadToday(project)
	if err == nil {
		// 最新のstartアクションを探す
		for i := len(logs) - 1; i >= 0; i-- {
			if logs[i].Action == actionStart {
				startTime = logs[i].Timestamp
				break
			}
		}

		// 累計時間を計算（List()と同じロジック）
		var sessionStart *time.Time
		for _, entry := range logs {
			switch entry.Action {
			case actionStart:
				t := entry.Timestamp
				sessionStart = &t
			case actionStop:
				if sessionStart != nil {
					duration := entry.Timestamp.Sub(*sessionStart)
					totalTime += duration
					sessionStart = nil
				}
			}
		}

		// 現在稼働中の場合、現在時刻までの時間を加算
		if sessionStart != nil {
			now := time.Now()
			duration := now.Sub(*sessionStart)
			totalTime += duration
			// 現在セッションの経過時間を計算
			currentSessionTime = now.Sub(startTime)
		}
	}

	// ProjectStatusを作成して返す
	status := &domain.ProjectStatus{
		Project:            project,
		Tag:                tag,
		TagName:            tagName,
		StartTime:          startTime,
		CurrentSessionTime: currentSessionTime,
		TotalTime:          totalTime,
	}

	return status, nil
}

// List は本日のプロジェクト一覧を返す（簡潔な名前）
func (m *projectManager) List() ([]domain.ProjectSummary, error) {
	// 本日のすべてのログエントリを取得
	allLogs, err := m.logStorage.ReadAllToday()
	if err != nil {
		return nil, err
	}

	// 現在稼働中のプロジェクトを取得
	currentProject := ""
	currentTag := ""
	current, err := m.currentStorage.Read()
	if err == nil && current != "" {
		// current文字列を解析
		if idx := strings.IndexRune(current, '\t'); idx != -1 {
			currentProject = current[:idx]
			currentTag = current[idx+1:]

			// 現在稼働中のプロジェクトのログがない場合、空のエントリを追加
			if _, exists := allLogs[currentProject]; !exists {
				allLogs[currentProject] = []storage.LogEntry{}
			}
		}
	}

	// 各プロジェクトの稼働時間を計算
	var summaries []domain.ProjectSummary
	for project, entries := range allLogs {
		summary := domain.ProjectSummary{
			Project: project,
		}

		// タグと最終アクティビティ時刻を取得
		if len(entries) > 0 {
			summary.Tag = entries[0].Tag
			summary.LastActivity = entries[len(entries)-1].Timestamp
		} else if project == currentProject {
			// 現在稼働中で履歴がない場合
			summary.Tag = currentTag
			summary.LastActivity = time.Now()
		}

		// タグ名を取得
		summary.TagName = m.getTagName(summary.Tag)

		// 稼働時間と時間範囲を計算
		totalDuration := time.Duration(0)
		var startTime *time.Time
		var timeRanges []domain.TimeRange

		for _, entry := range entries {
			switch entry.Action {
			case actionStart:
				// 開始時刻を記録
				t := entry.Timestamp
				startTime = &t
			case actionStop:
				// 開始時刻がある場合、期間を計算し時間範囲を追加
				if startTime != nil {
					duration := entry.Timestamp.Sub(*startTime)
					totalDuration += duration

					// 時間範囲を追加
					timeRanges = append(timeRanges, domain.TimeRange{
						Start:    *startTime,
						End:      entry.Timestamp,
						Duration: duration,
					})

					startTime = nil
				}
			}
		}

		// 現在稼働中の場合、現在時刻までの時間を追加
		if project == currentProject && startTime != nil {
			now := time.Now()
			duration := now.Sub(*startTime)
			totalDuration += duration

			// 稼働中の時間範囲を追加
			timeRanges = append(timeRanges, domain.TimeRange{
				Start:    *startTime,
				End:      now,
				Duration: duration,
			})
		}

		summary.TotalTime = totalDuration
		summary.TimeRanges = timeRanges
		summaries = append(summaries, summary)
	}

	// タグ名でソート（同一タグ内では最終アクティビティ降順）
	sort.Slice(summaries, func(i, j int) bool {
		// 第1キー: タグ名（昇順、空は最後）
		if summaries[i].TagName != summaries[j].TagName {
			if summaries[i].TagName == "" {
				return false
			}
			if summaries[j].TagName == "" {
				return true
			}
			return summaries[i].TagName < summaries[j].TagName
		}
		// 第2キー: 最終アクティビティ（降順）
		return summaries[i].LastActivity.After(summaries[j].LastActivity)
	})

	return summaries, nil
}

// ListOnDate は指定日のプロジェクト一覧と稼働時間を返す（currentは考慮しない）
func (m *projectManager) ListOnDate(date time.Time) ([]domain.ProjectSummary, error) {
	// 指定日のすべてのログエントリを取得
	allLogs, err := m.logStorage.ReadAllOnDate(date)
	if err != nil {
		return nil, err
	}

	// 各プロジェクトの稼働時間を計算
	var summaries []domain.ProjectSummary
	for project, entries := range allLogs {
		summary := domain.ProjectSummary{
			Project: project,
		}

		// タグと最終アクティビティ時刻を取得
		if len(entries) > 0 {
			summary.Tag = entries[0].Tag
			summary.LastActivity = entries[len(entries)-1].Timestamp
		}

		// タグ名を取得
		summary.TagName = m.getTagName(summary.Tag)

		// 稼働時間と時間範囲を計算
		totalDuration := time.Duration(0)
		var startTime *time.Time
		var timeRanges []domain.TimeRange

		for _, entry := range entries {
			switch entry.Action {
			case actionStart:
				// 開始時刻を記録
				t := entry.Timestamp
				startTime = &t
			case actionStop:
				// 開始時刻がある場合、期間を計算し時間範囲を追加
				if startTime != nil {
					duration := entry.Timestamp.Sub(*startTime)
					totalDuration += duration

					// 時間範囲を追加
					timeRanges = append(timeRanges, domain.TimeRange{
						Start:    *startTime,
						End:      entry.Timestamp,
						Duration: duration,
					})

					startTime = nil
				}
			}
		}

		// 指定日の場合、未完了のセッションは無視（currentを考慮しない）

		summary.TotalTime = totalDuration
		summary.TimeRanges = timeRanges
		summaries = append(summaries, summary)
	}

	// タグ名でソート（同一タグ内では最終アクティビティ降順）
	sort.Slice(summaries, func(i, j int) bool {
		// 第1キー: タグ名（昇順、空は最後）
		if summaries[i].TagName != summaries[j].TagName {
			if summaries[i].TagName == "" {
				return false
			}
			if summaries[j].TagName == "" {
				return true
			}
			return summaries[i].TagName < summaries[j].TagName
		}
		// 第2キー: 最終アクティビティ（降順）
		return summaries[i].LastActivity.After(summaries[j].LastActivity)
	})

	return summaries, nil
}

// NewAt は指定時刻で新規プロジェクトを開始する
func (m *projectManager) NewAt(project, tag string, timestamp time.Time) error {
	// 既存のプロジェクトが稼働中か確認
	current, err := m.currentStorage.Read()
	if err == nil && current != "" {
		// 稼働中のプロジェクトを自動停止
		// current文字列を解析（format: "ProjectName\tTagID"）
		var oldProject, oldTag string
		if idx := strings.IndexRune(current, '\t'); idx != -1 {
			oldProject = current[:idx]
			oldTag = current[idx+1:]
		} else {
			return errors.New("current情報の形式が不正です")
		}

		// 既存プロジェクトを停止（指定時刻で）
		err = m.logStorage.Append(oldProject, actionStop, oldTag, timestamp)
		if err != nil {
			return fmt.Errorf("既存プロジェクトの停止に失敗: %w", err)
		}
	}

	// currentファイルに新規プロジェクトを記録
	err = m.currentStorage.Write(project, tag)
	if err != nil {
		return err
	}

	// ログに開始を記録（指定時刻で）
	err = m.logStorage.Append(project, actionStart, tag, timestamp)
	if err != nil {
		// ログの記録に失敗した場合、currentファイルをクリア（ロールバック）
		m.currentStorage.Clear()
		return err
	}

	return nil
}

// SwitchAt は指定時刻でプロジェクトを切り替える
func (m *projectManager) SwitchAt(project, tag string, timestamp time.Time) error {
	// 既存のプロジェクト情報を取得
	current, err := m.currentStorage.Read()
	if err == nil && current != "" {
		// 稼働中のプロジェクトがある場合のみ停止処理
		// current文字列を解析（format: "ProjectName\tTagID"）
		var oldProject, oldTag string
		if idx := strings.IndexRune(current, '\t'); idx != -1 {
			oldProject = current[:idx]
			oldTag = current[idx+1:]
		} else {
			return errors.New("current情報の形式が不正です")
		}

		// 既存プロジェクトを停止（指定時刻で）
		err = m.logStorage.Append(oldProject, actionStop, oldTag, timestamp)
		if err != nil {
			return fmt.Errorf("既存プロジェクトの停止に失敗: %w", err)
		}
	}

	// 新規プロジェクトを開始（指定時刻で）
	err = m.logStorage.Append(project, actionStart, tag, timestamp)
	if err != nil {
		return fmt.Errorf("新規プロジェクトの開始に失敗: %w", err)
	}

	// currentファイルを更新
	err = m.currentStorage.Write(project, tag)
	if err != nil {
		return fmt.Errorf("currentファイルの更新に失敗: %w", err)
	}

	return nil
}

// StopAt は指定時刻でプロジェクトを停止する
func (m *projectManager) StopAt(timestamp time.Time) error {
	// 既存のプロジェクト情報を取得
	current, err := m.currentStorage.Read()
	if err != nil {
		return errors.New("稼働中のプロジェクトがありません")
	}

	// current文字列を解析（format: "ProjectName\tTagID"）
	var project, tag string
	if idx := strings.IndexRune(current, '\t'); idx != -1 {
		project = current[:idx]
		tag = current[idx+1:]
	} else {
		return errors.New("current情報の形式が不正です")
	}

	// プロジェクトを停止（指定時刻で）
	err = m.logStorage.Append(project, actionStop, tag, timestamp)
	if err != nil {
		return err
	}

	// currentファイルをクリア
	err = m.currentStorage.Clear()
	if err != nil {
		return err
	}

	return nil
}

// ListRecent は過去N日間に作業したプロジェクトの一覧を返す
func (m *projectManager) ListRecent(days int) ([]domain.ProjectSummary, error) {
	// 日付範囲を計算（今日から過去N日間）
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -days+1) // days日前から今日まで

	// 指定期間のすべてのログエントリを取得
	allLogs, err := m.logStorage.ReadRange(startDate, endDate)
	if err != nil {
		return nil, err
	}

	// プロジェクトごとの集計用マップ（プロジェクト名+タグIDをキーとする）
	projectMap := make(map[string]*domain.ProjectSummary)

	// 各プロジェクトの稼働時間を計算
	for project, entries := range allLogs {
		if len(entries) == 0 {
			continue
		}

		// プロジェクト名とタグIDの組み合わせでキーを作成
		tag := entries[0].Tag
		key := project + "\t" + tag

		// 既存のサマリーがあればそれを使用、なければ新規作成
		summary, exists := projectMap[key]
		if !exists {
			summary = &domain.ProjectSummary{
				Project: project,
				Tag:     tag,
				TagName: m.getTagName(tag),
			}
			projectMap[key] = summary
		}

		// 稼働時間と時間範囲を計算
		var startTime *time.Time
		for _, entry := range entries {
			switch entry.Action {
			case actionStart:
				// 開始時刻を記録
				t := entry.Timestamp
				startTime = &t
			case actionStop:
				// 開始時刻がある場合、期間を計算し時間範囲を追加
				if startTime != nil {
					duration := entry.Timestamp.Sub(*startTime)
					summary.TotalTime += duration

					// 時間範囲を追加
					summary.TimeRanges = append(summary.TimeRanges, domain.TimeRange{
						Start:    *startTime,
						End:      entry.Timestamp,
						Duration: duration,
					})

					startTime = nil
				}
			}

			// 最終アクティビティ時刻を更新
			if entry.Timestamp.After(summary.LastActivity) {
				summary.LastActivity = entry.Timestamp
			}
		}
	}

	// マップからスライスに変換
	var summaries []domain.ProjectSummary
	for _, summary := range projectMap {
		summaries = append(summaries, *summary)
	}

	// 最終作業日でソート（新しい順）
	sort.Slice(summaries, func(i, j int) bool {
		return summaries[i].LastActivity.After(summaries[j].LastActivity)
	})

	return summaries, nil
}

// GetTags はタグ一覧を取得する
func (m *projectManager) GetTags() ([]domain.Tag, error) {
	return m.tagStorage.Load()
}
