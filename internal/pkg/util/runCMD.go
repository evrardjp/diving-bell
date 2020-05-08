package util

import (
	"bufio"
	"bytes"
	"io"
	"os"
	"os/exec"
	"os/user"
	"path"
	"strings"

	"k8s.io/klog"
)

type CmdResults struct {
	ExitCode int
	Output   string
	Error    error
}

var Debug bool

func pathExpansion(filePath string) string {
	usr, err := user.Current()
	if err != nil {
		klog.Fatalf("getting current user failed: %s", err)
	}

	if strings.HasPrefix(filePath, "~/") {
		filePath = path.Join(usr.HomeDir, filePath[2:])
	}

	return filePath
}

// RunShell runs a shell cmd
func RunShell(shellCmd string) bool {
	args := strings.Fields(shellCmd)
	cmd := exec.Command(args[0], args[1:len(args)]...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		klog.Fatal(err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		klog.Fatal(err)
	}

	cmd.Start()

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		klog.Infoln(scanner.Text())
	}

	errScanner := bufio.NewScanner(stderr)
	for errScanner.Scan() {
		klog.Warningln(errScanner.Text())
	}

	cmd.Wait()

	klog.Infoln()

	return cmd.ProcessState.ExitCode() == 0
}

// RunShellOutput run a shell command and get the output
func RunShellOutput(shellCmd string) CmdResults {
	args := strings.Fields(shellCmd)
	cmd := exec.Command(args[0], args[1:len(args)]...)

	out, err := cmd.Output()

	return CmdResults{ExitCode: cmd.ProcessState.ExitCode(), Output: string(out), Error: err}
}

// RunShellAt run a shell command at a given location
func RunShellAt(shellCmd string, dirPath string) CmdResults {
	buf := new(bytes.Buffer)
	args := strings.Fields(shellCmd)
	mWriter := io.MultiWriter(buf, os.Stdout)
	cmd := exec.Command(args[0], args[1:len(args)]...)
	cmd.Dir = pathExpansion(dirPath)
	cmd.Stdout = mWriter
	cmd.Stderr = mWriter

	cmd.Run()

	klog.Infoln()

	return CmdResults{ExitCode: cmd.ProcessState.ExitCode(), Output: buf.String()}
}
