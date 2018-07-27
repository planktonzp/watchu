package main

import (
	"github.com/containous/traefik/log"
	"net/http"
	"os"
	"time"

	"regexp"
	"flag"
	"fmt"
	"strings"
	"bytes"
	"io/ioutil"
	"encoding/json"

)

const regular = `^(1[0-9])\d{9}$`

type WatchOnit struct {
	Proc      string
	Args      []string
	Contacts  []string
	APIADDR   string
	HeartBeat int64
}

type PostMsg struct {
	Mobiles []string `json:"mobiles"`
	Content string   `json:"content"`
}

type RespMsg struct {
	Status string `json:"status"`
	Msg    string `json:"msg"`
}

var u WatchOnit

func validate(mobileNum, rules string) bool {
	reg := regexp.MustCompile(rules)
	return reg.MatchString(mobileNum)
}

func GFromCmd() {
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

	var tmpvar []string = make([]string,0)
	startsymbol := false
	endsymbol := false
	u.Args = append(u.Args, cmd)
	for _, a := range strings.Split(arg, " ") {
		if a != " " {
			if !startsymbol {
				startsymbol = validate(a , `^"`)
			}
			//当元素以'开始时，进入临时数组并整合为一个字符串输入进u.Args
			if startsymbol {
				endsymbol = validate(a, `"$`)
				tmpvar = append(tmpvar, a)
				if endsymbol {
					str := strings.Replace(strings.Join(tmpvar," "),`"`,``,-1)
					u.Args = append(u.Args,str)
					startsymbol = false
					endsymbol = false
					tmpvar = make([]string,0)
				}
			} else {
				u.Args = append(u.Args,a)
			}
		}
	}

	for _, num = range strings.Split(contact, ",") {
		//检查电话号码是否合法
		tf := validate(num, regular)
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

func GMsgOrNot(a string) string {
	if len(u.Contacts) != 0 {
		var postmsg PostMsg
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
		var respmsg RespMsg
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

func GUccu() {

	Attr := &os.ProcAttr{
		Files: []*os.File{os.Stdin, os.Stdout, os.Stderr},
	}

	p, err := os.StartProcess(u.Proc, u.Args, Attr)

	log.Info(p)

	if err != nil {
		log.Error(err)
		GMsgOrNot(u.Proc)
	}
	r, err := p.Wait()
	if err != nil {
		log.Error(err)
	}
	//Wait退出，不管是人为原因还是异常状态都发短信通知
	log.Info(r)
	GMsgOrNot(u.Proc)

	time.Sleep(time.Duration(u.HeartBeat) * time.Second)
}

func main() {

	//获取参数
	GFromCmd()
	//测试短信接口是否正常
	GMsgOrNot("api测试短信")

	//一旦主程序退出,发短信
	defer GMsgOrNot("uccu主进程退出")
	for {
		GUccu()
		//异常退出
	}
}

