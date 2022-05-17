package clog

import (
	"io"
	"log"
	"os"
)

var (
	Logger *log.Logger
)

func InitLog(fileLoc string) {
	Logger = log.New(
		io.MultiWriter(os.Stdout, &customFileWriter{fileLoc: fileLoc}),
		"logger:",
		log.Lshortfile)
}

type customFileWriter struct {
	fileLoc string
}

func (cf *customFileWriter) Write(p []byte) (n int, err error) {
	if cf.fileLoc == "" {
		return 0, nil
	}

	err = os.WriteFile(cf.fileLoc, p, 0744)
	if err != nil {
		return 0, err
	}
	return len(p), nil
}
