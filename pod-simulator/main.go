package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"time"
)

func main() {
	echoDelay := 3 * time.Second
	if delay, err := strconv.Atoi(os.Getenv("ECHO_DELAY")); err == nil {
		echoDelay = time.Duration(delay) * time.Second
	}

	// Source - https://www.geeksforgeeks.org/how-to-read-a-file-line-by-line-to-string-in-golang/

	// os.Open() opens specific file in
	// read-only mode and this return
	// a pointer of type os.
	file, err := os.Open("/var/logs/simulation.log")

	if err != nil {
		panic("failed to open simulation.log")

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
