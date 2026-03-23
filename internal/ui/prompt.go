// Package ui はワークログアプリケーションの対話的UIコンポーネントを担当する
// charmbracelet/huh ライブラリを使用してターミナルインタラクションを提供
package ui

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"worklog/internal/domain"

	"github.com/charmbracelet/huh"
	"github.com/mattn/go-runewidth"
)

// ProjectDisplay は選択リストで表示するプロジェクト情報
type ProjectDisplay struct {
	Project       string
	Tag           string
	TagName       string // タグ名
	Time          string // 表示用の時間文字列
	Status        string // 状態アイコン（■ stopped / ▫ paused）
	IsRunning     bool   // 現在稼働中かどうか
	DateLabel     string // 日付ラベル（[Today], [2 days ago]など）
	IsSeparator   bool   // セパレーター行かどうか
	SeparatorText string // セパレーター行の表示テキスト
}

// PromptUI は対話的UIのインターフェース
type PromptUI interface {
	SelectTag(tags []domain.Tag) (string, error)
	InputProject() (string, error)
	ConfirmAction(action string) (bool, error)
	SelectProjectFromList(projects []ProjectDisplay) (*ProjectDisplay, error)
	InputTime(label string) (string, error)
}

// formatTagID はタグIDを2桁幅で右揃えフォーマットする
func formatTagID(id int) string {
	return fmt.Sprintf("%2d", id)
}

// truncateProjectName はプロジェクト名をターミナル幅に合わせて切り詰める
// termWidth: ターミナルの幅、tagWidth: タグ名の表示幅（0の場合はタグなし）
func truncateProjectName(project string, termWidth int, tagWidth int) string {
	// 固定要素の幅を計算
	// "  ▸ " (4) + "  " (2) + timeStr (8) + "  " (2) + status (10) + "  " (2) + dateLabel (13) + margin (5)
	fixedWidth := 46
	if tagWidth > 0 {
		fixedWidth += tagWidth + 2 // タグ名 + "  " セパレーター
	}

	maxProjectWidth := max(termWidth-fixedWidth, 12) // 最低でも12文字は表示（省略記号を含む）

	currentWidth := runewidth.StringWidth(project)
	if currentWidth <= maxProjectWidth {
		return project
	}

	return runewidth.Truncate(project, maxProjectWidth, "…")
}

// truncateProjectWithTag はタグ名を考慮してプロジェクト名を切り詰める
func truncateProjectWithTag(project, tagName string) string {
	tagWidth := runewidth.StringWidth(tagName)
	return truncateProjectName(project, GetTerminalWidth(), tagWidth)
}

// buildProjectLabel はProjectDisplayからhuh選択リスト用の表示文字列を生成する
func buildProjectLabel(pd ProjectDisplay) string {
	var sb strings.Builder
	name := truncateProjectWithTag(pd.Project, pd.TagName)
	sb.WriteString(name)
	if pd.TagName != "" {
		sb.WriteString("  ")
		sb.WriteString(pd.TagName)
	}
	sb.WriteString("  ")
	sb.WriteString(pd.Time)
	sb.WriteString("  ")
	sb.WriteString(pd.Status)
	if pd.DateLabel != "" {
		sb.WriteString("  ")
		sb.WriteString(pd.DateLabel)
	}
	return sb.String()
}

// validateTimeInput は時刻入力のバリデーション関数
func validateTimeInput(input string) error {
	// 空欄は許可（現在時刻を使用）
	if input == "" {
		return nil
	}

	// HHMM形式（4桁の数字）かチェック
	if len(input) == 4 {
		for _, ch := range input {
			if ch < '0' || ch > '9' {
				return errors.New("時刻は HH:MM または HHMM 形式で入力してください")
			}
		}
		return nil
	}

	// HH:MM形式の簡易チェック
	if len(input) < 3 || len(input) > 5 {
		return errors.New("時刻は HH:MM または HHMM 形式で入力してください")
	}

	colonCount := 0
	for _, ch := range input {
		if ch == ':' {
			colonCount++
		}
	}
	if colonCount != 1 {
		return errors.New("時刻は HH:MM または HHMM 形式で入力してください")
	}

	return nil
}

// promptUIImpl はPromptUIインターフェースの実装
type promptUIImpl struct{}

// NewPromptUI は新しいPromptUIインスタンスを作成する
func NewPromptUI() PromptUI {
	return &promptUIImpl{}
}

// SelectTag はタグをリストから選択する
func (p *promptUIImpl) SelectTag(tags []domain.Tag) (string, error) {
	if len(tags) == 0 {
		return "", errors.New("利用可能なタグがありません")
	}

	options := make([]huh.Option[string], len(tags))
	for i, tag := range tags {
		label := fmt.Sprintf("%s - %s", formatTagID(tag.ID), tag.Name)
		options[i] = huh.NewOption(label, strconv.Itoa(tag.ID))
	}

	var selected string
	err := huh.NewSelect[string]().
		Title("タグを選択").
		Options(options...).
		Value(&selected).
		WithTheme(huh.ThemeBase()).
		Run()
	if err != nil {
		return "", err
	}

	return selected, nil
}

// InputProject はプロジェクト名を入力する
func (p *promptUIImpl) InputProject() (string, error) {
	var result string
	err := huh.NewInput().
		Title("プロジェクト名").
		Value(&result).
		Validate(func(s string) error {
			if s == "" {
				return errors.New("プロジェクト名を入力してください")
			}
			if len(s) > 300 {
				return errors.New("プロジェクト名は300文字以内で入力してください")
			}
			return nil
		}).
		WithTheme(huh.ThemeBase()).
		Run()
	if err != nil {
		return "", err
	}

	return result, nil
}

// ConfirmAction はアクションの確認を行う
func (p *promptUIImpl) ConfirmAction(action string) (bool, error) {
	var confirmed bool
	err := huh.NewConfirm().
		Title(fmt.Sprintf("%s しますか？", action)).
		Value(&confirmed).
		WithTheme(huh.ThemeBase()).
		Run()
	if err != nil {
		// Escape/Ctrl+C をキャンセル（false）として扱う
		if errors.Is(err, huh.ErrUserAborted) {
			return false, nil
		}
		return false, err
	}

	return confirmed, nil
}

// SelectProjectFromList はプロジェクトリストから選択する
// セパレーター項目はオプションリストから除外し、DateLabelで日付グループを識別する
func (p *promptUIImpl) SelectProjectFromList(projects []ProjectDisplay) (*ProjectDisplay, error) {
	if len(projects) == 0 {
		return nil, errors.New("利用可能なプロジェクトがありません")
	}

	// セパレーターを除いたオプションリストを構築（値は元スライスのインデックス）
	var options []huh.Option[int]
	firstIdx := -1
	for i, pd := range projects {
		if pd.IsSeparator {
			continue
		}
		if firstIdx == -1 {
			firstIdx = i
		}
		options = append(options, huh.NewOption(buildProjectLabel(pd), i))
	}

	if len(options) == 0 {
		return nil, errors.New("利用可能なプロジェクトがありません")
	}

	selectedIdx := firstIdx
	err := huh.NewSelect[int]().
		Title("プロジェクトを選択").
		Options(options...).
		Value(&selectedIdx).
		WithTheme(huh.ThemeBase()).
		Run()
	if err != nil {
		return nil, err
	}

	return &projects[selectedIdx], nil
}

// InputTime は時刻を入力させる（空欄で現在時刻）
func (p *promptUIImpl) InputTime(label string) (string, error) {
	now := time.Now().Format("15:04")
	var result string
	err := huh.NewInput().
		Title(fmt.Sprintf("%s (HH:MM, 空欄で現在時刻 = %s)", label, now)).
		Value(&result).
		Validate(validateTimeInput).
		WithTheme(huh.ThemeBase()).
		Run()
	if err != nil {
		return "", err
	}

	return result, nil
}
