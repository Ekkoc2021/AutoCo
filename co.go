package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

// 命令行直接启动不附带参数，则将使用默认的账号密码完成连接
var user string = "0921****" //默认账号
var password string = "****" //默认密码

func main() {
	log.Println(
		"\n---------------------小E-v1.1----------------------------------\n" +
			"|命令行用法:                                                  |\n" +
			"| 1,co 默认连接                                               |\n" +
			"| 2,co [任意] 切换到cumt_stu并连接                            |\n" +
			"| 3,co [账号] [密码] 使用账号密码直接连接                     |\n" +
			"| 4,co [任意] [账号] [密码] 账号密码连接并且切换wifi到cumt_stu|" +
			"\n---------------------------------------------------------------\n")

	for !analysis() {
		log.Print("输入1重使用，输入其他退出程序！")
		var input string
		fmt.Scan(&input)
		if input != "1" {
			os.Exit(0)
		}
	}
	log.Println("运行完成")
	if len(os.Args) > 1 {
		log.Println("7s后程序自动结束,可以手动退出!")
		time.Sleep(7 * time.Second)
	}
}


func analysis() bool {
	if len(os.Args) == 1 {
		return verify(user, password)

	}
	if len(os.Args) == 2 {
		//切换wifi
		if !connect() {
			return false
		}
		//登录验证!
		return verify(user, password)

	}
	if len(os.Args) == 3 {
		return verify(os.Args[1], os.Args[2])

	}
	if len(os.Args) == 4 {
		//切换wifi
		if !connect() {
			return false
		}
		//登录验证!
		return verify(os.Args[2], os.Args[3])
	}
	return false
}

/*
切换wifi到cumt_
*/
func connect() bool {
	log.Println("切换wifi到CUMT_Stu...")
	//切换wifi到cumt_stu
	command := exec.Command("cmd", "/C", "netsh wlan connect name=CUMT_Stu")
	_, err := command.Output()
	if err != nil {
		log.Println("切换失败：可能是wifi功能未开启!")
		return false
	}
	log.Println("切换成功! 等待系统WIFI切换完成...")
	return true
}

type Resp struct {
	Result  string `json:"result"`
	Msg     string `json:"msg"`
	RetCode string `json:"ret_code"`
}

/*
	   验证身份
		user :用户名称
		password:用户密码
*/
func verify(user string, password string) (myr bool) {

	log.Println("认证中...")
	time.Sleep(2 * time.Second) //电脑启动,或者wifi切换,系统可能反应不过来,容易导致验证失败
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	//req, err := http.NewRequest("GET", "http://10.2.5.251:801/eportal/?c=Portal&a=login&callback=dr1678184325643&login_method=1&user_account=09213643%40unicom&user_password=bv623977&wlan_user_ip=10.3.221.67&wlan_user_mac=14857fcc574f&wlan_ac_ip=&wlan_ac_name=NAS&jsVersion=3.0&_=1678184294511", nil)
	req, err := http.NewRequest("GET", "http://10.2.5.251:801/eportal/?c=Portal&a=login&login_method=1&user_account="+user+"%40unicom&user_password="+password, nil)
	if err != nil {

		log.Println("错误:wifi未连接或未打开！\n")
		myr = false
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Println("错误:wifi未连接或未打开！")
		return false
	}
	defer resp.Body.Close()

	bodyText, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("错误:wifi未连接或未打开！")
		return false
	}

	s := string(bodyText)
	re := s[strings.Index(s, "{"):strings.Index(s, ")")]
	res := new(Resp)
	json.Unmarshal([]byte(re), &res)
	if res.RetCode == "2" {
		log.Println("认证失败!\n响应数据：", re, "\n导致错误原因可能有:\n1,账号或密码错误。\n2,当前时间不允许上网冲浪!\n3,没有切换wifi到CUMT_Stu。\n4,当前电脑已完成认证。")
		return myr
	}
	if res.RetCode == "0" {
		log.Println("认证失败!,"+"\n响应数据：", re, "\n导致错误原因可能有:已完成认证或没有切换wifi。")
		return myr
	}
	log.Print(res.Msg)
	return true
}
