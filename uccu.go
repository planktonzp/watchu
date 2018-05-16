package main

import "os"

type CMD struct {
	u_proc      string
	u_argv       []string
	u_attr       *os.ProcAttr
	u_pid        *os.Process
}

func main(u CMD) {
	u.u_attr =&os.ProcAttr{
		Files: []*os.File{os.Stdin,os.Stdout,os.Stderr},
	}
	p, err := os.StartProcess(u.u_proc, u.u_argv,u.u_attr}, attr)
}

