package logger

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"
)

type Logger struct {
	file *os.File
}

func New(logFile string) (*Logger, error) {
	file, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return nil, err
	}

	return &Logger{file: file}, nil
}

func (l *Logger) Info(format string, v ...interface{}) {
	l.log("INFO", format, v...)
}

func (l *Logger) Error(format string, v ...interface{}) {
	l.log("ERROR", format, v...)
}

func (l *Logger) Debug(format string, v ...interface{}) {
	l.log("DEBUG", format, v...)
}

func (l *Logger) Chat(user, message string) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	logLine := fmt.Sprintf("[%s] CHAT: %s: %s\n", timestamp, user, message)
	l.file.WriteString(logLine)
	l.file.Sync()
}

func (l *Logger) log(level, format string, v ...interface{}) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	message := fmt.Sprintf(format, v...)
	logLine := fmt.Sprintf("[%s] %s: %s\n", timestamp, level, message)

	// Write to both file and stdout
	l.file.WriteString(logLine)
	fmt.Print(logLine)
	l.file.Sync()
}

func (l *Logger) Close() error {
	return l.file.Close()
}

// GetRecentMessages reads the last N chat messages from the log file
func (l *Logger) GetRecentMessages(count int) ([]string, error) {
	// Get the log file path from the file descriptor
	logFile := l.file.Name()

	// Open file for reading
	file, err := os.Open(logFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Read all lines
	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		// Only include CHAT messages, not INFO/ERROR logs
		// Also filter out system join/leave messages (contain ">>>" or "<<<")
		if strings.Contains(line, "CHAT:") &&
			!strings.Contains(line, ">>>") &&
			!strings.Contains(line, "<<<") {
			lines = append(lines, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	// Return last N messages
	if len(lines) > count {
		lines = lines[len(lines)-count:]
	}

	return lines, nil
}
