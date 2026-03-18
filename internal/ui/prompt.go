// Package ui はワークログアプリケーションの対話的UIコンポーネントを担当する
// promptuiライブラリを使用してターミナルインタラクションを強化
package ui

import (
	"errors"
	"fmt"
	"strconv"
	"text/template"

	"worklog/internal/domain"

	"github.com/manifoldco/promptui"
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
	InputTime(label string) (string, error) // 時刻入力用メソッドを追加
}

// formatTagID はタグIDを2桁幅で右揃えフォーマットする
func formatTagID(id int) string {
	return fmt.Sprintf("%2d", id)
}

// faint はpromptuiの内部テンプレートで使用される可能性のある関数
// 色なしでそのまま返す
func faint(s string) string {
	return s
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

	maxProjectWidth := termWidth - fixedWidth
	if maxProjectWidth < 12 {
		maxProjectWidth = 12 // 最低でも12文字は表示（省略記号を含む）
	}

	currentWidth := runewidth.StringWidth(project)
	if currentWidth <= maxProjectWidth {
		return project
	}

	return runewidth.Truncate(project, maxProjectWidth, "…")
}

// truncateProjectWithTag はテンプレート関数としてタグ名を考慮してプロジェクト名を切り詰める
func truncateProjectWithTag(project, tagName string) string {
	tagWidth := runewidth.StringWidth(tagName)
	return truncateProjectName(project, GetTerminalWidth(), tagWidth)
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

	funcMap := template.FuncMap{
		"padID": formatTagID,
		"faint": faint,
	}

	templates := &promptui.SelectTemplates{
		Label:    "{{ . }}",
		Active:   "  " + Arrow + " {{ padID .ID }} - {{ .Name }}",
		Inactive: "    {{ padID .ID }} - {{ .Name }}",
		Selected: "{{ padID .ID }} - {{ .Name }}",
		Details:  "", // Detailsを空文字列に設定してデフォルトを無効化
		FuncMap:  funcMap,
	}

	// ヘッダーとヒントを含むラベル
	label := fmt.Sprintf("Select tag\n%s", renderSeparator(38))

	prompt := promptui.Select{
		Label:     label,
		Items:     tags,
		Templates: templates,
		Size:      10,
		// promptuiは自動的に "Use the arrow keys to navigate: ↓ ↑ → ←" を表示
	}

	i, _, err := prompt.Run()
	if err != nil {
		return "", err
	}

	return strconv.Itoa(tags[i].ID), nil
}

// InputProject はプロジェクト名を入力する
func (p *promptUIImpl) InputProject() (string, error) {
	validate := func(input string) error {
		if input == "" {
			return errors.New("プロジェクト名を入力してください")
		}
		if len(input) > 300 {
			return errors.New("プロジェクト名は300文字以内で入力してください")
		}
		return nil
	}

	prompt := promptui.Prompt{
		Label:    "Project name",
		Validate: validate,
	}

	result, err := prompt.Run()
	if err != nil {
		return "", err
	}

	return result, nil
}

// ConfirmAction はアクションの確認を行う
func (p *promptUIImpl) ConfirmAction(action string) (bool, error) {
	prompt := promptui.Prompt{
		Label:     fmt.Sprintf("%s しますか？", action),
		IsConfirm: true,
	}

	result, err := prompt.Run()
	if err != nil {
		// プロンプトがキャンセルされた場合
		if err == promptui.ErrAbort {
			return false, nil
		}
		return false, err
	}

	return result == "y" || result == "Y", nil
}

// SelectProjectFromList はプロジェクトリストから選択する
func (p *promptUIImpl) SelectProjectFromList(projects []ProjectDisplay) (*ProjectDisplay, error) {
	if len(projects) == 0 {
		return nil, errors.New("利用可能なプロジェクトがありません")
	}

	// プロジェクトの表示用リストを作成
	funcMap := template.FuncMap{
		"faint":           faint,
		"truncateProject": truncateProjectWithTag,
	}

	templates := &promptui.SelectTemplates{
		Label:    "{{ . }}",
		Active:   "{{ if .IsSeparator }}    {{ .SeparatorText }}{{ else }}  " + Arrow + " {{ truncateProject .Project .TagName }}{{ if .TagName }}  {{ faint .TagName }}{{ end }}  {{ .Time }}  {{ .Status }}{{ if .DateLabel }}  {{ .DateLabel }}{{ end }}{{ end }}",
		Inactive: "{{ if .IsSeparator }}    {{ .SeparatorText }}{{ else }}    {{ truncateProject .Project .TagName }}{{ if .TagName }}  {{ faint .TagName }}{{ end }}  {{ .Time }}  {{ .Status }}{{ if .DateLabel }}  {{ .DateLabel }}{{ end }}{{ end }}",
		Selected: "{{ if .IsSeparator }}{{ .SeparatorText }}{{ else }}{{ truncateProject .Project .TagName }}{{ if .TagName }}  {{ faint .TagName }}{{ end }}  {{ .Time }}  {{ .Status }}{{ if .DateLabel }}  {{ .DateLabel }}{{ end }}{{ end }}",
		Details:  "", // Detailsを空文字列に設定してデフォルトを無効化
		FuncMap:  funcMap,
	}

	// ヘッダーとヒントを含むラベル
	label := fmt.Sprintf("Select project\n%s", renderSeparator(38))

	prompt := promptui.Select{
		Label:     label,
		Items:     projects,
		Templates: templates,
		Size:      10,
	}

	i, _, err := prompt.Run()
	if err != nil {
		return nil, err
	}

	// セパレーターが選択された場合は再度選択を促す
	if projects[i].IsSeparator {
		return p.SelectProjectFromList(projects)
	}

	return &projects[i], nil
}

// InputTime は時刻を入力させる（空欄で現在時刻）
func (p *promptUIImpl) InputTime(label string) (string, error) {
	// 時刻入力のバリデーション関数
	validate := func(input string) error {
		// 空欄は許可（現在時刻を使用）
		if input == "" {
			return nil
		}

		// HHMM形式（4桁の数字）かチェック
		if len(input) == 4 {
			// 全て数字かチェック
			for _, ch := range input {
				if ch < '0' || ch > '9' {
					return errors.New("時刻は HH:MM または HHMM 形式で入力してください")
				}
			}
			// 4桁の数字の場合は有効
			return nil
		}

		// HH:MM形式の簡易チェック
		// parseTimeArgで詳細な検証を行うので、ここでは簡易チェックのみ
		if len(input) < 3 || len(input) > 5 {
			return errors.New("時刻は HH:MM または HHMM 形式で入力してください")
		}

		// コロンが含まれているかチェック
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

	prompt := promptui.Prompt{
		Label:    fmt.Sprintf("%s (HH:MM or HHMM, empty for now)", label),
		Validate: validate,
		Default:  "", // デフォルトは空欄（現在時刻）
	}

	result, err := prompt.Run()
	if err != nil {
		return "", err
	}

	return result, nil
}

