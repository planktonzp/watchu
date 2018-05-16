package main

import (
	"flag"
	"fmt"
	"github.com/containous/traefik/log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
	"io/ioutil"
)

const regular = `^(13[0-9]|14[579]|15[0-3,5-9]|16[6]|17[0135678]|18[0-9]|19[89])\d{8}$`

type WatchOnit struct {
	Proc      string
	Args      []string
	Contacts  []string
	HeartBeat int64
}

var u WatchOnit

func validate(mobileNum string) bool {
	reg := regexp.MustCompile(regular)
	return reg.MatchString(mobileNum)
}

func FromCmd(CmdArgs WatchOnit) WatchOnit {
	var contact, arg, num string

	//不确定是不是应该写一个-h或者用别的方式来设计.... 明天问
	flag.StringVar(&CmdArgs.Proc, "cmd", "", "需要监控的程序")
	flag.StringVar(&arg, "args", "", "程序启动的参数")
	flag.StringVar(&contact, "tel", "", "告警联系人电话,多个时用逗号分开")
	flag.Int64Var(&CmdArgs.HeartBeat, "hb", 60, "心跳频率,单位:秒")

	CmdArgs.Args[0] = arg
	numbers := strings.Split(contact, ",")

	for num = range numbers {
		if validate(num) {
			CmdArgs.Contacts = append(CmdArgs.Contacts, num)
		} else {
			fmt.Println("请输入合法的电话号码,多个时以逗号分开")
		}
	}

	return CmdArgs
}

func MsgOrNot(CmdArgs WatchOnit) {
	var n string
	for n = range CmdArgs.Contacts {
		if n != "" {
			//api明天问问是什么
			url := fmt.Sprintf("http://MSG_API/%s/content", n)

			req, err := http.NewRequest("POST", url, nil)

			if err != nil {
				log.Error(err)
				return
			}
			//不明白这个地方到底该怎么改... 明天记得问
			response, _ := http.Client.Do(req)
			body,_ := ioutil.ReadAll(response.Body)

			//假设返回的body为"发送成功"或"发送失败"
			fmt.Printf("号码%s短信%s",n,body)
		}
	}
	//不知道是不是需要一个返回值,明天问
}

func main() {

	FromCmd(u)
	flag.Parse()

	Attr := &os.ProcAttr{
		Files: []*os.File{os.Stdin, os.Stdout, os.Stderr},
	}

	//一旦监控的程序或者参数提交错误 是不是会引起这个程序无限重启导致死循环.... 不太明白这里为啥不用signal控制重启...
	for true {
		p, err := os.StartProcess(u.Proc, u.Args, Attr)

		log.Info(p)

		if err != nil {
			log.Error(err)
			return
		}
		r, err := p.Wait()
		if err != nil {
			log.Error(err)
			return
		}
		log.Info(r)
		//  重启后发告警短信
		MsgOrNot(u)

		time.Sleep(time.Duration(u.HeartBeat) * time.Second)
	}
}
