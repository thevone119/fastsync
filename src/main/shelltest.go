package main

import (
	"fmt"
	"github.com/hpcloud/tail"
	"os"
	"strings"
)

func main() {
	//bolt 库
	//t, err := tail.TailFile("e:test/test2.txt", tail.Config{Follow: true})
	filename := "e:/test/test2.txt"
	t, err := tail.TailFile(filename, tail.Config{Follow: true})
	if err != nil {
		fmt.Println(err)
	}
	for line := range t.Lines {
		fmt.Println(line.Text)
	}

}

func runCommand(commandStr string) error {
	commandStr = strings.TrimSuffix(commandStr, "\n")
	arrcmd := strings.Fields(commandStr)
	switch arrcmd[0] {
	case "syncup":
		fmt.Fprintln(os.Stdout, "syncup")
	case "testsync":
		fmt.Fprintln(os.Stdout, "testsync")
	}
	return nil
}
