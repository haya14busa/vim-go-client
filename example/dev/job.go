package main

import (
	"bufio"
	"os"
)

func main() {
	// io.Copy(os.Stdin, os.Stdout)

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		// log.Printf("stdin: %v\n", scanner.Text())

		os.Stdout.WriteString(`["expr", "1 + 2", -1]`)

		// os.Stdout.WriteString(fmt.Sprintf("%s\n", scanner.Text()))
	}
}
