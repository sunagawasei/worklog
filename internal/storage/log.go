package storage

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"worklog/internal/domain"
)

const logFileExtension = ".log"

// LogStorage はログファイルの読み書きを行うインターフェース
type LogStorage interface {
	Append(project, action, tag string, timestamp time.Time) error
	ReadToday(project string) ([]LogEntry, error)
	ReadAllToday() (map[string][]LogEntry, error)
	ReadAllOnDate(date time.Time) (map[string][]LogEntry, error)
	ReadRange(startDate, endDate time.Time) (map[string][]LogEntry, error)
}

// LogEntry は domain.LogEntry のエイリアス
type LogEntry = domain.LogEntry

// logStorage はLogStorageインターフェースの実装
type logStorage struct {
	basePath string
}

// NewLogStorage は新しいLogStorageインスタンスを作成
func NewLogStorage(basePath string) LogStorage {
	return &logStorage{
		basePath: basePath,
	}
}

// Append はログエントリをファイルに追記
func (s *logStorage) Append(project, action, tag string, timestamp time.Time) error {
	// 日付別ディレクトリのパスを生成
	year := timestamp.Format("2006")
	month := timestamp.Format("01")
	day := timestamp.Format("02")
	dirPath := filepath.Join(s.basePath, year, month, day)

	// ディレクトリが存在しない場合は作成
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// ログファイルのパスを生成
	logPath := filepath.Join(dirPath, project+logFileExtension)

	// ファイルを追記モードで開く（存在しない場合は作成）
	file, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}
	defer file.Close()

	// ログエントリをタブ区切り形式でフォーマット
	entry := fmt.Sprintf("%s\t%s\t%s\n",
		timestamp.Format("2006/01/02 15:04:05"),
		action,
		tag,
	)

	// ファイルに書き込み
	if _, err := file.WriteString(entry); err != nil {
		return fmt.Errorf("failed to write log entry: %w", err)
	}

	return nil
}

// ReadToday は本日分のログエントリを読み込む
func (s *logStorage) ReadToday(project string) ([]LogEntry, error) {
	// 本日の日付ディレクトリのパスを生成
	now := time.Now()
	year := now.Format("2006")
	month := now.Format("01")
	day := now.Format("02")
	logPath := filepath.Join(s.basePath, year, month, day, project+logFileExtension)

	// 共通のログファイル読み込み関数を使用
	return s.readLogFile(logPath)
}

// ReadAllToday は本日分のすべてのプロジェクトのログエントリを読み込む
func (s *logStorage) ReadAllToday() (map[string][]LogEntry, error) {
	// 本日の日付でReadAllOnDateを呼び出す
	return s.ReadAllOnDate(time.Now())
}

// ReadAllOnDate は指定日のすべてのプロジェクトのログエントリを読み込む
func (s *logStorage) ReadAllOnDate(date time.Time) (map[string][]LogEntry, error) {
	// 指定日の日付ディレクトリのパスを生成
	year := date.Format("2006")
	month := date.Format("01")
	day := date.Format("02")
	dirPath := filepath.Join(s.basePath, year, month, day)

	// ディレクトリが存在しない場合は空のマップを返す
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		return make(map[string][]LogEntry), nil
	}

	// ディレクトリ内のログファイルを列挙
	files, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	// プロジェクトごとのログエントリを格納するマップ
	result := make(map[string][]LogEntry)

	for _, file := range files {
		// .logファイル以外はスキップ
		if file.IsDir() || !strings.HasSuffix(file.Name(), logFileExtension) {
			continue
		}

		// プロジェクト名を取得（.logを除去）
		project := strings.TrimSuffix(file.Name(), logFileExtension)

		// 指定日のログファイルパスを生成
		logPath := filepath.Join(dirPath, file.Name())

		// ログエントリを読み込み
		entries, err := s.readLogFile(logPath)
		if err != nil {
			// エラーが発生してもスキップして続行
			continue
		}

		if len(entries) > 0 {
			result[project] = entries
		}
	}

	return result, nil
}

// readLogFile はログファイルからエントリを読み込む（共通処理）
func (s *logStorage) readLogFile(logPath string) ([]LogEntry, error) {
	// ファイルを開く（存在しない場合は空のスライスを返す）
	file, err := os.Open(logPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []LogEntry{}, nil
		}
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}
	defer file.Close()

	// bufio.Scannerで行ごとに読み込む
	var entries []LogEntry
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()

		// タブで分割
		parts := strings.Split(line, "\t")
		if len(parts) != 3 {
			// 不正な形式の行はスキップ
			continue
		}

		// タイムスタンプをパース
		timestamp, err := time.ParseInLocation("2006/01/02 15:04:05", parts[0], time.Local)
		if err != nil {
			// パースエラーの場合はスキップ
			continue
		}

		// エントリを追加
		entry := LogEntry{
			Timestamp: timestamp,
			Action:    parts[1],
			Tag:       parts[2],
		}
		entries = append(entries, entry)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading log file: %w", err)
	}

	return entries, nil
}

// ReadRange は指定期間内の全プロジェクトのログエントリを読み込む
func (s *logStorage) ReadRange(startDate, endDate time.Time) (map[string][]LogEntry, error) {
	// プロジェクトごとのログエントリを格納するマップ
	result := make(map[string][]LogEntry)

	// startDateからendDateまでの各日付を処理
	currentDate := startDate
	for !currentDate.After(endDate) {
		// 各日のログエントリを取得
		dailyEntries, err := s.ReadAllOnDate(currentDate)
		if err != nil {
			return nil, err
		}

		// 各プロジェクトのエントリをresultにマージ
		for project, entries := range dailyEntries {
			result[project] = append(result[project], entries...)
		}

		// 次の日に進む
		currentDate = currentDate.AddDate(0, 0, 1)
	}

	return result, nil
}
