package main

import "os"

type CMD struct {
	u_proc      string
	u_argv       []string
	u_attr       *os.ProcAttr
	u_pid        *os.Process
}

func main() {
	attr:=&os.ProcAttr{
		Files: []*os.File{os.Stdin,os.Stdout,os.Stderr},
	}
}

