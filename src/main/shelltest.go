package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main(){
	//bolt 库
	reader:=bufio.NewReader(os.Stdin)
	for{
		fmt.Print("$")
		cmdString,err:=reader.ReadString('\n')
		if err!=nil{
			fmt.Fprintln(os.Stderr,err)
		}
		runCommand(cmdString)
	}
}

func runCommand(commandStr string) error{
	commandStr=strings.TrimSuffix(commandStr,"\n")
	arrcmd:=strings.Fields(commandStr)
	switch arrcmd[0] {
	case "syncup":
		fmt.Fprintln(os.Stdout,"syncup")
	case "testsync":
		fmt.Fprintln(os.Stdout,"testsync")
	}
	return nil
}