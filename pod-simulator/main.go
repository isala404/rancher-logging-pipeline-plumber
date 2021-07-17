package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"time"
)

func main() {
	var logDir string
	var echoDelay time.Duration
	flag.DurationVar(&echoDelay, "delay", 3*time.Second, "interval between 2 log echos")
	flag.StringVar(&logDir, "log-dir", "", "absolute path for the log file")
	flag.Parse()

	if len(logDir) == 0 {
		fmt.Println("Usage: ./pod-simulator -log-dir")
		flag.PrintDefaults()
		os.Exit(1)
	}

	// Source - https://www.geeksforgeeks.org/how-to-read-a-file-line-by-line-to-string-in-golang/

	// os.Open() opens specific file in
	// read-only mode and this return
	// a pointer of type os.
	file, err := os.Open(logDir)

	if err != nil {
		panic(fmt.Sprintf("failed to %s", logDir))
	}

	// The bufio.NewScanner() function is called in which the
	// object os.File passed as its parameter and this returns a
	// object bufio.Scanner which is further used on the
	// bufio.Scanner.Split() method.
	scanner := bufio.NewScanner(file)

	// The bufio.ScanLines is used as an
	// input to the method bufio.Scanner.Split()
	// and then the scanning forwards to each
	// new line using the bufio.Scanner.Scan()
	// method.
	scanner.Split(bufio.ScanLines)
	var text []string

	for scanner.Scan() {
		text = append(text, scanner.Text())
	}

	// The method os.File.Close() is called
	// on the os.File object to close the file
	file.Close()

	// ------------------------------------------------------------------------------------------

	for ; ; {
		for _, line := range text {
			fmt.Println(line)
			time.Sleep(echoDelay)
		}
	}
}
