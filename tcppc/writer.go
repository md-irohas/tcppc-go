package tcppc

import (
	"github.com/jehiah/go-strftime"
	"log"
	"os"
	"path/filepath"
	"strconv"
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
	// Off-set of rotation interval.
	RotOffset int64
	// Location used as timezone in FileNameFmt.
	Location *time.Location
	// Current file object.
	file *os.File
	// Last rotation time.
	lstRotTime int64
	// True if this writer is closed. This flag is used to make a goroutine
	// exit to update this writer.
	closed bool
	// Number of session data written to file.
	numSessions int
	// Mutex object for writing data to file.
	mutex sync.RWMutex
}

func NewWriter(fileNameFmt string, rotInt, rotOffset int64, loc *time.Location) *RotWriter {
	w := &RotWriter{FileNameFmt: fileNameFmt, RotInt: rotInt, RotOffset: rotOffset, Location: loc, file: nil, lstRotTime: 0, closed: false}

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

	// If the filename already exists, find an alternative filename.
	// e.g. foobar.pcap -> foobar-1.pcap
	for i := 1; fileExists(fileName); i++ {
		log.Printf("File already exists: %s\n", fileName)

		fileNameFmtExt := filepath.Ext(fileNameFmt)
		fileNameFmtBase := fileNameFmt[0 : len(fileNameFmt)-len(fileNameFmtExt)]

		newFileNameFmt := fileNameFmtBase + "-" + strconv.Itoa(i) + fileNameFmtExt
		fileName = strftime.Format(newFileNameFmt, tmpTime)
	}

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
		w.lstRotTime = curTime

		fileName := w.findFileName(curTime)
		dirName := filepath.Dir(fileName)

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

		w.numSessions = 0
		w.file = file
	}
}

func (w *RotWriter) Write(data []byte) (n int, err error) {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	w.numSessions += 1
	return w.file.Write(append(data, 0x0a))
}

func (w *RotWriter) Close() error {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	w.closed = true
	return w.file.Close()
}
