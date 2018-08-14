package main

import (
	"fmt"
	"os"
	//"syscall"
	"runtime/debug"
	"time"
)

func TryE() {
	errs := recover()
	if errs == nil {
		return
	}
	exeName := os.Args[0] //获取程序名称

	now := time.Now()  //获取当前时间
	pid := os.Getpid() //获取进程ID

	time_str := now.Format("20060102150405")                          //设定时间格式
	fname := fmt.Sprintf("%s-%d-%s-dump.log", exeName, pid, time_str) //保存错误信息文件名:程序名-进程ID-当前时间（年月日时分秒）
	fmt.Println("dump to file ", fname)

	f, err := os.Create(fname)
	if err != nil {
		return
	}
	defer f.Close()

	f.WriteString(fmt.Sprintf("%v\r\n", errs)) //输出panic信息
	f.WriteString("========\r\n")

	f.WriteString(string(debug.Stack())) //输出堆栈信息
}

func main() {

	//	log_file, err := os.OpenFile("/tmp/output.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0766)
	//	if err != nil {
	//		log.Fatal(err)
	//	}
	//	defer log_file.Close()
	//	log_file.Write([]byte("I'm waitting for a bug\n"))
	//	log_time := time.Now().Format("2006-01-02 15:04:0" +
	//		"5")
	//
	//	c := make(chan os.Signal)
	//	signal.Notify(c)
	//	//监听指定信号
	//	//signal.Notify(c, syscall.SIGHUP, syscall.SIGUSR2)
	//
	//	//阻塞直至有信号传入
	//	s := <-c
	//	fmt.Println("get signal:", s)
	//	log_content := strings.Join([]string{"====", log_time, "====", s.String(), "\n"}, "")
	//	buf := []byte(log_content)
	//	log_file.Write(buf)
	//	fmt.Println(log_content)
	defer TryE()
	fmt.Println(time.Now())
	panic(-2)
	fmt.Println("panic restore now")
}
