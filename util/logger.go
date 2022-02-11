package util

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"
)

const (
	LOG_FILE_DIR = "./tmp"
)

var (
	Log *log.Logger
)

func init() {
	// Set location of log file
	logpath := fmt.Sprintf("%s/%s.log", LOG_FILE_DIR, time.Now().Format("2006-01-02"))

	flag.Parse()

	file, err1 := os.Create(logpath)

	if err1 != nil {
		panic(err1)
	}

	Log = log.New(file, "", log.LstdFlags|log.Lshortfile)

	Log.Println("LogFile : " + logpath)
}
