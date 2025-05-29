package watcher

import (
	"io"
	"log/slog"
	"strings"

	"github.com/nxadm/tail"
)

type ADIWatcher struct {
	filePath string
	tailer   *tail.Tail
	callback func(string)
}

func NewADIWatcher(filePath string, callback func(string)) (*ADIWatcher, error) {
	slog.Info("Creating ADI file watcher", "file_path", filePath)
	t, err := tail.TailFile(filePath, tail.Config{
		Location: &tail.SeekInfo{
			Offset: 0,
			Whence: io.SeekEnd,
		},
		Follow: true,
	})
	if err != nil {
		return nil, err
	}
	return &ADIWatcher{
		filePath: filePath,
		tailer:   t,
		callback: callback,
	}, nil
}

func (w *ADIWatcher) Start() error {
	slog.Info("Starting file monitoring", "file_path", w.filePath)
	go w.watch()
	return nil
}

func (w *ADIWatcher) watch() {
	cache := ""
	for line := range w.tailer.Lines {
		cache += line.Text
		if strings.Contains(cache, "<eor>") {
			w.callback(cache)
			cache = ""
		}
	}
}

func (w *ADIWatcher) Close() {
	slog.Info("Closing file watcher", "file_path", w.filePath)
	w.tailer.Cleanup()
}
