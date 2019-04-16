package tcppc

import (
	"github.com/jehiah/go-strftime"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

type RotWriter struct {
	// Filename format w/ time indicators of strftime.
	FileNameFmt string
	// Rotation interval in second.
	RotInt int64
	// Rotation interval offset in second.
	RotOffset int64
	// Location used as timezone in FileNameFmt.
	Location *time.Location
	// Current file object.
	file *os.File
	// Last rotation time.
	lstRotTime int64
	// True if this writer is closed.
	// This flag is used to make the update goroutine exit.
	closed bool
	// Number of session data written to file.
	numSessions int
	// Mutex object for exclusive control of writing data to file.
	mutex sync.RWMutex
}

func NewWriter(fileNameFmt string, rotInt, rotOffset int, loc *time.Location) *RotWriter {
	w := &RotWriter{
		FileNameFmt: fileNameFmt,
		RotInt:      int64(rotInt),
		RotOffset:   int64(rotOffset),
		Location:    loc,
		file:        nil,
		lstRotTime:  0,
		closed:      false,
	}

	go func() {
		for {
			if w.closed {
				break
			}

			w.update()
			time.Sleep(100 * time.Millisecond)
		}
	}()

	return w
}

func (w *RotWriter) findFileName(ts int64) string {
	// Convert unix time to native time.
	tmpTime := time.Unix(ts, 0).In(w.Location)

	// Fill format of date and time in FileNameFmt.
	fileNameFmt := w.FileNameFmt
	fileName := strftime.Format(fileNameFmt, tmpTime)

	return fileName
}

func (w *RotWriter) update() {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	curTime := time.Now().Unix()

	if curTime > w.lstRotTime && w.RotInt > 0 && (curTime%w.RotInt) == w.RotOffset {
		if w.file != nil {
			w.file.Close()
			w.file = nil

			log.Printf("Wrote %d session data.\n", w.numSessions)
		}
	}

	if w.file == nil {
		fileName := w.findFileName(curTime)
		dirName := filepath.Dir(fileName)

		// Create directories if not exists.
		if !fileExists(dirName) {
			err := os.MkdirAll(dirName, 0755)
			if err == nil {
				log.Printf("Create directories: %s\n", dirName)
			} else {
				log.Fatalf("Failed to create directories: %s (%s)", dirName, err)
			}
		}

		file, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
		if err == nil {
			log.Printf("Created a session file: %s\n", fileName)
		} else {
			log.Fatalf("Failed to create a session file: %s (%s)\n", fileName, err)
		}

		w.lstRotTime = curTime
		w.numSessions = 0
		w.file = file
	}
}

func (w *RotWriter) Write(data []byte) (n int, err error) {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	w.numSessions += 1

	// Write data to file with '\n'.
	return w.file.Write(append(data, 0x0a))
}

func (w *RotWriter) Close() error {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	w.closed = true

	return w.file.Close()
}
