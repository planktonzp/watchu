package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	//"syscall"
	"strings"
	"time"
)

func main() {

	log_file, err := os.OpenFile("/tmp/output.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0766)
	if err != nil {
		log.Fatal(err)
	}
	defer log_file.Close()
	log_file.Write([]byte("I'm waitting for a bug\n"))
	log_time := time.Now().Format("2006-01-02 15:04:05")

	c := make(chan os.Signal)
	signal.Notify(c)
	//监听指定信号
	//signal.Notify(c, syscall.SIGHUP, syscall.SIGUSR2)

	//阻塞直至有信号传入
	s := <-c
	fmt.Println("get signal:", s)
	log_content := strings.Join([]string{"====", log_time, "====", s.String(), "\n"}, "")
	buf := []byte(log_content)
	log_file.Write(buf)
	fmt.Println(log_content)
}
