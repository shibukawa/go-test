package main

import (
	"fmt"
	"github.com/fatih/color"
	"os/exec"
	"os"
	"sync"
	"syscall"
	"errors"
	"bufio"
	"strings"
)

func main() {
	args := append([]string{"test"}, os.Args[1:]...)
	command := exec.Command("go", args...)
	command.Dir, _ = os.Getwd()
	command.Env = os.Environ()

	var outputWait sync.WaitGroup
	outputWait.Add(2)

	stdout, err := command.StdoutPipe()
	if err != nil {
		panic(err)
	}
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			text := scanner.Text()
			if strings.HasPrefix(text, "=== RUN") {
				fmt.Fprintf(os.Stdout, color.BlueString("%s\n", text))
			} else if strings.HasPrefix(text, "--- FAIL") {
					fmt.Fprintf(os.Stdout, color.YellowString("%s\n", text))
			} else if strings.Contains(text, "FAIL") {
				fmt.Fprintf(os.Stdout, color.RedString("%s\n", text))
			} else if text == "PASS" || strings.HasPrefix(text, "--- PASS") || strings.HasPrefix(text, "ok") {
				fmt.Fprintf(os.Stdout, color.GreenString("%s\n", text))
			} else {
				fmt.Fprintf(os.Stdout, "%s\n", text)
			}
		}
		outputWait.Done()
	}()

	stderr, err := command.StderrPipe()
	if err != nil {
		panic(err)
	}
	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			fmt.Fprintf(os.Stderr, color.RedString("%s\n", scanner.Text()))
		}
		outputWait.Done()
	}()

	command.Start()
	err = command.Wait()

	stdout.Close()
	stderr.Close()

	// http://qiita.com/hnakamur/items/5e6f22bda8334e190f63
	var exitStatus int
	if err != nil {
		if e2, ok := err.(*exec.ExitError); ok {
			if s, ok := e2.Sys().(syscall.WaitStatus); ok {
				exitStatus = s.ExitStatus()
			} else {
				panic(errors.New("Unimplemented for system where exec.ExitError.Sys() is not syscall.WaitStatus."))
			}
		}
	}
	outputWait.Wait()
	os.Exit(exitStatus)
}