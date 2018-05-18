package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/containous/traefik/log"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
)

const regular = `^(13[0-9]|14[579]|15[0-3,5-9]|16[6]|17[0135678]|18[0-9]|19[89])\d{8}$`

type WatchOnit struct {
	Proc      string
	Args      []string
	Contacts  []string
	HeartBeat int64
	APIADDR   string
}

type POSTMSG struct {
	Mobiles []string `json:"mobiles"`
	Content string   `json:"content"`
}

type RESPMSG struct {
	Status string `json:"status"`
	Msg    string `json:"msg"`
}

func validate(mobileNum string) bool {
	reg := regexp.MustCompile(regular)
	return reg.MatchString(mobileNum)
}

func FromCmd(CmdArgs WatchOnit) WatchOnit {
	var proc, contact, arg, num string

	//不确定是不是应该写一个-h或者用别的方式来设计.... 明天问
	flag.StringVar(&proc, "bin", "", "需要监控的程序")
	flag.StringVar(&arg, "args", "", "程序启动的参数")
	flag.StringVar(&contact, "tel", "", "告警联系人电话,多个时用逗号分开")
	flag.StringVar(&CmdArgs.APIADDR, "api", "", "短信api地址")
	flag.Int64Var(&CmdArgs.HeartBeat, "hb", 60, "心跳频率,单位:秒")

	CmdArgs.Proc = proc
	CmdArgs.Args = append(CmdArgs.Args, proc)
	CmdArgs.Args = append(CmdArgs.Args, arg)

	for _, num = range strings.Split(contact, ",") {
		if validate(num) {
			CmdArgs.Contacts = append(CmdArgs.Contacts, num)
		} else {
			fmt.Println("请输入合法的电话号码,多个时以逗号分开")
		}
	}

	return CmdArgs
}

func MsgOrNot(CmdArgs WatchOnit) string {
	if len(CmdArgs.Contacts) != 0 {
		var postmsg POSTMSG
		postmsg.Mobiles = CmdArgs.Contacts
		postmsg.Content = fmt.Sprintf("%s又挂啦,修不了啦", CmdArgs.Proc)

		bytesData, _ := json.Marshal(postmsg)
		reader := bytes.NewReader(bytesData)
		url := fmt.Sprintf("%s%s", CmdArgs.APIADDR, CmdArgs.Proc)
		req, err := http.NewRequest("POST", url, reader)
		if err != nil {
			fmt.Println(err.Error())
			log.Error(err)
			return fmt.Sprint("请求无响应,请检查您输入的地址")
		}
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		client := http.Client{}
		response, _ := client.Do(req)
		body, _ := ioutil.ReadAll(response.Body)
		var respmsg RESPMSG
		e := json.Unmarshal(body, &respmsg)
		if e != nil {
			fmt.Println(e.Error())
			return "短信接口返回异常"
		}
		return respmsg.Msg
	}
	return "没写联系人，我也不知道联系谁"
}

func uccu(c WatchOnit) {

	FromCmd(c)
	flag.Parse()

	Attr := &os.ProcAttr{
		Files: []*os.File{os.Stdin, os.Stdout, os.Stderr},
	}
	//一旦监控的程序或者参数提交错误 是不是会引起这个程序无限重启导致死循环.... 不太明白这里为啥不用signal控制重启...
	p, err := os.StartProcess(c.Proc, c.Args, Attr)

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

	time.Sleep(time.Duration(c.HeartBeat) * time.Second)
}
func main() {
	var u WatchOnit
	for {
		defer MsgOrNot(u)
		uccu(u)
	}
}
