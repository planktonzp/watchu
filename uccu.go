package main

import (
	"bytes"
	"encoding/json"
	flag "flag"
	"fmt"
	"github.com/containous/traefik/log"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
)

const regular = `^(1[0-9])\d{9}$`

type WatchOnit struct {
	Proc      string
	Args      []string
	Contacts  []string
	APIADDR   string
	HeartBeat int64
}

type POSTMSG struct {
	Mobiles []string `json:"mobiles"`
	Content string   `json:"content"`
}

type RESPMSG struct {
	Status string `json:"status"`
	Msg    string `json:"msg"`
}

var u WatchOnit

func validate(mobileNum string) bool {
	reg := regexp.MustCompile(regular)
	return reg.MatchString(mobileNum)
}

func FromCmd() {
	var (
		cmd     string
		arg     string
		contact string
		api     string
		num     string
		hb      int64
	)

	flag.StringVar(&cmd, "cmd", "", "需要监控的程序")
	flag.StringVar(&arg, "arg", "", "程序启动的参数")
	flag.StringVar(&contact, "tel", "", "告警联系人电话,多个时用逗号分开")
	flag.StringVar(&api, "api", "", "短信api地址")
	flag.Int64Var(&hb, "hb", 60, "心跳频率,单位:秒")

	flag.Parse()

	fmt.Printf("cmd:%s", cmd)
	fmt.Printf("u.Proc:%s", u.Proc)
	u.Proc = cmd

	u.Args = append(u.Args, cmd)
	for _, a := range strings.Split(arg, " ") {
		u.Args = append(u.Args, a)
	}

	for _, num = range strings.Split(contact, ",") {
		tf := validate(num)
		if tf {
			u.Contacts = append(u.Contacts, num)
		} else {
			fmt.Println("请输入合法的电话号码,多个时以逗号分开")
		}
	}

	u.APIADDR = api

	u.HeartBeat = hb

	return
}

func MsgOrNot(a string) string {
	if len(u.Contacts) != 0 {
		var postmsg POSTMSG
		postmsg.Mobiles = u.Contacts
		if a == u.Proc {
			postmsg.Content = fmt.Sprintf("%s退出,请悉知", u.Proc)
		} else {
			postmsg.Content = a
		}

		bytesData, _ := json.Marshal(postmsg)
		reader := bytes.NewReader(bytesData)
		url := fmt.Sprintf("%succu", u.APIADDR)
		req, err := http.NewRequest("POST", url, reader)
		if err != nil {
			log.Error(err)
			return "请求无响应,请检查您输入的地址"
		}

		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		client := http.Client{}
		response, err := client.Do(req)
		if err != nil {
			log.Error(err)
			return "返回信息异常，请检查api"
		}

		body, _ := ioutil.ReadAll(response.Body)
		var respmsg RESPMSG
		e := json.Unmarshal(body, &respmsg)
		if e != nil {
			fmt.Println(e.Error())
			return "短信接口返回异常"
		}

		return respmsg.Msg
	} else {
		return "没写联系人，我也不知道联系谁"
	}
}

func uccu() {

	Attr := &os.ProcAttr{
		Files: []*os.File{os.Stdin, os.Stdout, os.Stderr},
	}
	//一旦监控的程序或者参数提交错误 是不是会引起这个程序无限重启导致死循环.... 不太明白这里为啥不用signal控制重启...
	p, err := os.StartProcess(u.Proc, u.Args, Attr)

	log.Info(p)

	if err != nil {
		log.Error(err)
		MsgOrNot(u.Proc)
	}
	r, err := p.Wait()
	if err != nil {
		log.Error(err)
	}
	//Wait退出，不管是人为原因还是异常状态都发短信通知
	log.Info(r)
	MsgOrNot(u.Proc)

	time.Sleep(time.Duration(u.HeartBeat) * time.Second)
}

func main() {
	//获取参数
	FromCmd()
	//测试短信接口是否正常
	MsgOrNot("api测试短信")

	//一旦主程序退出,发短信
	defer MsgOrNot("uccu主进程退出")
	for {
		uccu()
	}
}
