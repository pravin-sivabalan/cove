package cove

import (
	"fmt"
	"log"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
)

type FileWatcher struct {
	watcher  *fsnotify.Watcher
	filename string
	onChange func()
}

func NewFileWatcher(filename string, onChange func()) (*FileWatcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create watcher: %w", err)
	}

	fw := &FileWatcher{
		watcher:  watcher,
		filename: filename,
		onChange: onChange,
	}

	// Add the file to the watcher
	err = watcher.Add(filename)
	if err != nil {
		watcher.Close()
		return nil, fmt.Errorf("failed to watch file: %w", err)
	}

	// Start watching in a goroutine
	go fw.watchLoop()

	return fw, nil
}

func (fw *FileWatcher) watchLoop() {
	for {
		select {
		case event, ok := <-fw.watcher.Events:
			if !ok {
				return
			}

			// Check if this is our file and if it was modified
			if filepath.Clean(event.Name) == filepath.Clean(fw.filename) {
				if event.Has(fsnotify.Write) {
					if fw.onChange != nil {
						fw.onChange()
					}
				}
			}

		case err, ok := <-fw.watcher.Errors:
			if !ok {
				return
			}
			log.Printf("File watcher error: %v", err)
		}
	}
}

func (fw *FileWatcher) Close() error {
	return fw.watcher.Close()
}