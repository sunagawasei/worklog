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
	NewAt(project, tag string, timestamp time.Time) error    // 指定時刻で新規プロジェクト開始
	SwitchAt(project, tag string, timestamp time.Time) error // 指定時刻でプロジェクト切り替え
	StopAt(timestamp time.Time) error                        // 指定時刻で停止
	Status() (*domain.ProjectStatus, error)                  // 現在の状況
	List() ([]domain.ProjectSummary, error)                  // 本日の一覧
	ListOnDate(date time.Time) ([]domain.ProjectSummary, error) // 指定日の一覧
	ListRecent(days int) ([]domain.ProjectSummary, error)    // 過去N日間の一覧
	GetTags() ([]domain.Tag, error)                          // タグ一覧を取得
	AddTag(name string) (domain.Tag, error)                  // タグを追加
	DeleteTag(id int) error                                   // タグを削除
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

// parseCurrent はcurrent文字列をプロジェクト名とタグIDに分割する
// format: "ProjectName\tTagID"
func parseCurrent(current string) (project, tag string, err error) {
	idx := strings.IndexRune(current, '\t')
	if idx == -1 {
		return "", "", errors.New("current情報の形式が不正です")
	}
	return current[:idx], current[idx+1:], nil
}

// calculateSessions はログエントリから完了セッションの合計時間・時間範囲と、
// 未完了セッションの開始時刻（openStart）を返す。
// openStart が非nilの場合、対応するstopなしのstartが存在する（現在稼働中）。
func calculateSessions(entries []domain.LogEntry) (total time.Duration, ranges []domain.TimeRange, openStart *time.Time) {
	for _, entry := range entries {
		switch entry.Action {
		case actionStart:
			t := entry.Timestamp
			openStart = &t
		case actionStop:
			if openStart != nil {
				d := entry.Timestamp.Sub(*openStart)
				total += d
				ranges = append(ranges, domain.TimeRange{
					Start:    *openStart,
					End:      entry.Timestamp,
					Duration: d,
				})
				openStart = nil
			}
		}
	}
	return
}

// Status は現在の稼働状況を返す（簡潔な名前）
func (m *projectManager) Status() (*domain.ProjectStatus, error) {
	current, err := m.currentStorage.Read()
	if err != nil {
		// 稼働中のプロジェクトがない場合はnilを返す（エラーではない）
		return nil, nil
	}

	project, tag, err := parseCurrent(current)
	if err != nil {
		return nil, err
	}

	tagName := m.getTagName(tag)

	var startTime time.Time
	totalTime := time.Duration(0)
	currentSessionTime := time.Duration(0)

	logs, err := m.logStorage.ReadToday(project)
	if err == nil {
		total, ranges, openStart := calculateSessions(logs)
		totalTime = total
		if openStart != nil {
			now := time.Now()
			startTime = *openStart
			totalTime += now.Sub(*openStart)
			currentSessionTime = now.Sub(startTime)
		} else if len(ranges) > 0 {
			// 全セッション完了済み: 最後の完了セッションの開始時刻を使用
			startTime = ranges[len(ranges)-1].Start
		}
	}

	return &domain.ProjectStatus{
		Project:            project,
		Tag:                tag,
		TagName:            tagName,
		StartTime:          startTime,
		CurrentSessionTime: currentSessionTime,
		TotalTime:          totalTime,
	}, nil
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
		if p, t, parseErr := parseCurrent(current); parseErr == nil {
			currentProject = p
			currentTag = t

			// 現在稼働中のプロジェクトのログがない場合、空のエントリを追加
			if _, exists := allLogs[currentProject]; !exists {
				allLogs[currentProject] = []domain.LogEntry{}
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
		total, ranges, openStart := calculateSessions(entries)

		// 現在稼働中の場合、現在時刻までの時間を追加
		if project == currentProject && openStart != nil {
			now := time.Now()
			d := now.Sub(*openStart)
			total += d
			ranges = append(ranges, domain.TimeRange{
				Start:    *openStart,
				End:      now,
				Duration: d,
			})
		}

		summary.TotalTime = total
		summary.TimeRanges = ranges
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

		// 稼働時間と時間範囲を計算（指定日の未完了セッションは無視）
		total, ranges, _ := calculateSessions(entries)
		summary.TotalTime = total
		summary.TimeRanges = ranges
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
		oldProject, oldTag, err := parseCurrent(current)
		if err != nil {
			return err
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
		oldProject, oldTag, err := parseCurrent(current)
		if err != nil {
			return err
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

	project, tag, err := parseCurrent(current)
	if err != nil {
		return err
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
		total, ranges, _ := calculateSessions(entries)
		summary.TotalTime += total
		summary.TimeRanges = append(summary.TimeRanges, ranges...)

		// 最終アクティビティ時刻を更新
		for _, entry := range entries {
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

// AddTag は新しいタグを追加する
func (m *projectManager) AddTag(name string) (domain.Tag, error) {
	return m.tagStorage.Add(name)
}

// DeleteTag は指定したIDのタグを削除する
func (m *projectManager) DeleteTag(id int) error {
	return m.tagStorage.Delete(id)
}
