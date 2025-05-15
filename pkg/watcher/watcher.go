package watcher

import (
	"bufio"
	"log/slog"
	"os"
	"strings"

	"github.com/fsnotify/fsnotify"
)

type ADIWatcher struct {
	filePath string
	watcher  *fsnotify.Watcher
	callback func(string)
}

func NewADIWatcher(filePath string, callback func(string)) (*ADIWatcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	slog.Info("Creating ADI file watcher", "file_path", filePath)
	return &ADIWatcher{
		filePath: filePath,
		watcher:  watcher,
		callback: callback,
	}, nil
}

func (w *ADIWatcher) Start() error {
	if err := w.watcher.Add(w.filePath); err != nil {
		return err
	}

	slog.Info("Starting file monitoring", "file_path", w.filePath)
	go w.watch()
	return nil
}

func (w *ADIWatcher) watch() {
	for {
		select {
		case event, ok := <-w.watcher.Events:
			if !ok {
				return
			}
			if event.Op&fsnotify.Write == fsnotify.Write {
				w.processNewLines()
			}
		case err, ok := <-w.watcher.Errors:
			if !ok {
				return
			}
			slog.Error("File watcher error", "error", err)
			panic(err)
		}
	}
}

func (w *ADIWatcher) processNewLines() {
	file, err := os.Open(w.filePath)
	if err != nil {
		slog.Error("Cannot open file", "file_path", w.filePath, "error", err)
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var lastLine string
	for scanner.Scan() {
		lastLine = scanner.Text()
	}

	if err := scanner.Err(); err != nil {
		slog.Error("Error while reading file", "file_path", w.filePath, "error", err)
		return
	}

	if lastLine != "" && strings.Contains(lastLine, "<eor>") {
		w.callback(lastLine)
	}
}

func (w *ADIWatcher) Close() error {
	slog.Info("Closing file watcher", "file_path", w.filePath)
	return w.watcher.Close()
}
