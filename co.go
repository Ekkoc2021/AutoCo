package main

import (
	login "AutoCo/login"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

var defaultYysC int = 1 //自定义运营商
var filename = "info.json"
var info = login.LoginInfo{}

/*
移动:cmcc
联通:unicom
电信:telecom
*/
var yys [3]string = [3]string{"cmcc", "unicom", "telecom"}

func main() {
	fmt.Println(
		"--------------欢迎使用:AutoCo v2.1-----------------\n" +
			"| 使用参数usage查看具体命令行使用详情,如: co usage|\n" +
			"| GitHub：https://github.com/Ekkoc2021/AutoCo |\n" +
			"| 反馈邮箱：189890049@qq.com                  |" +
			"\n---------------------------------------------------")
	for !analysis() {
		log.Print("输入 1 ：重新尝试登录，输入 2 ：修改默认账号密码，输入其他退出程序！")
		var input string
		fmt.Print("请输入：")
		fmt.Scan(&input)
		if input != "1" && input != "2" {
			os.Exit(0)
		}
		if input == "2" {
			info.Update(filename)
		}
	}
}

// 完成通过文件登录的逻辑,最终返回一个登录数据集:数据不正确,第一次登录要求向文件中数据,同时修改默认的登录密码账号
func loginWithFile(filename string) {
	// 读取数据
	info.ReadInfoInFile(filename)

	//初步校验数据
	for !info.DataIsRight() {
		//数据格式不正确:更新数据
		log.Println(":第一次运行或者运营商输入不在范围内!")
		info.Update(filename)
	}
	//更新完毕:运营商映射问题()
	defaultYysC = info.GetYysCode()
}

/*
解析命令行参数,并完成对应逻辑
*/
func analysis() bool {
	if len(os.Args) == 1 {
		//使用默认数据登录
		// 初始化默认数据
		loginWithFile(filename)
		// 默认是使用联通
		return isVerify(info.Username, info.Password, defaultYysC)
	}
	if len(os.Args) == 2 {
		var commend string = os.Args[1]
		if commend == "usage" {
			usage()
			os.Exit(0) //结束程序运行
		}
		if !connect() { //切换wifi
			return false
		}
		//使用默认数据登录
		// 初始化默认数据
		loginWithFile(filename)
		//登录验证!
		return isVerify(info.Username, info.Password, defaultYysC)
	}
	if len(os.Args) == 4 {
		info.Username = os.Args[1]
		info.Password = os.Args[2]
		c, err := strconv.Atoi(os.Args[3])
		if err != nil {
			log.Println("命令错误,无法解析")
			usage()
			time.Sleep(2 * time.Second)
			os.Exit(0)
		}
		defaultYysC = mapping(c) //对c进行处理
		return isVerify(info.Username, info.Password, defaultYysC)
	}
	if len(os.Args) == 5 {
		if !connect() { //切换wifi
			return false
		}
		info.Username = os.Args[2]
		info.Password = os.Args[3]
		c, err := strconv.Atoi(os.Args[4])
		if err != nil {
			usage()
			os.Exit(0)
		}
		defaultYysC = mapping(c) //对c进行处理
		return isVerify(info.Username, info.Password, defaultYysC)
	}
	log.Println("命令错误,无法解析")
	usage()
	time.Sleep(2 * time.Second)
	os.Exit(0)
	return false //执行不到
}

/*
用于对传入的运营商的数值的过滤,返回一个符合要求的值
*/
func mapping(c int) int {
	if c < 0 {
		return 0
	}
	if c > 2 {
		return 2
	}
	return c
}

/*
*
输出相关用法
*/
func usage() {
	//列出用法
	log.Println(
		"命令行用法:\n" +
			" 1,co 使用默认账号密码运营商进行连接，首次使用需要输入账号密码\n" +
			" 2,co [1] 切换到cumt_stu并使用默认账号密码运营商进行连接\n" +
			" 3,co [账号] [密码] [0~2] 指定账号密码运营商连接,0:移动,1:连通,2:电信\n" +
			" 4,co [1] [账号] [密码] [0~2] 切换wifi到cumt_stu然后通过指定账号密码运营商连接,0:移动,1:连通,2:电信 \n")
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
		log.Println("切换失败：可能是wifi功能未开启")
		return false
	}
	time.Sleep(3 * time.Second) //休眠2s,等待切换完成!
	log.Println("切换成功! 等待系统WIFI切换完成...")
	return true
}

/*
*
Resutl:10登录出错
*/
type Resp struct {
	Result  string `json:"result"`
	Msg     string `json:"msg"`
	RetCode string `json:"ret_code"`
}

// 2023/3/30 为了更加的通用,适配了移动,电信登录 并且打算重写verify这个函数，将认证和具体视图逻辑分离
/*
	    验证身份:拼装url,发送请求,封装请求数据
		user :用户名称
		password:用户密码
		yysN: 运营商编号
		返回一个封装请求json数据的结构体:Resp
*/
func verify2(user string, password string, yysN int) Resp {
	res := Resp{"0", "", "4"}
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	yysS := yys[yysN] //通过传入的运营商的索引获取对应的运营商
	req, err := http.NewRequest("GET", "http://10.2.5.251:801/eportal/?c=Portal&a=login&login_method=1&user_account="+user+"%40"+yysS+"&user_password="+password, nil)
	if err != nil {
		res.Msg = "未连接到CUMT_Stu"
		return res
	}
	resp, err := client.Do(req)
	if err != nil {
		res.Msg = "未连接到CUMT_Stu"
		return res
	}
	defer resp.Body.Close()

	bodyText, err := io.ReadAll(resp.Body)
	if err != nil {
		res.Msg = "未连接到CUMT_Stu"
		return res
	}

	s := string(bodyText)
	re := s[strings.Index(s, "{"):strings.Index(s, ")")]

	json.Unmarshal([]byte(re), &res)
	return res
}

/*
	    验证身份
		user :用户名称
		password:用户密码
		yysN: 运营商编号
		解析验证后的数据后,返回是否成功验证
*/
func isVerify(user string, password string, yysN int) bool {
	//msg 成功: 认证成功 重复认证: 无    账号密码错误运营商:dXNlcmlkIGVycm9yMQ==  未连接:未连接到CUMT_Stu
	//retCode 成功: 无		重复认证: 2    账号密码错误运营商:1					未连接:3
	//result 成功: 1		重复认证: 0	 账号密码错误运营商:0   					未连接:0

	log.Println("开始认证")
	resp := verify2(user, password, yysN)

	if resp.Result == "1" {
		log.Println(resp.Msg)
		return true
	}

	if resp.RetCode == "2" {
		//重复认证,也算认证成功!
		log.Println("认证成功")
		return true
	}

	if resp.RetCode == "1" {
		log.Println("认证失败:可能是运营商,密码,账号错误!还可能是当前时段不允许上网!")
		return false
	}

	if resp.RetCode == "4" {
		log.Println(resp.Msg)
		return false
	}
	log.Println("未知错误导致验证失败!该RetCode值未被记录!" +
		"\n		RetCode:" + resp.RetCode +
		"\n		Msg:" + resp.Msg +
		"\n		Result:" + resp.Result)
	return false
}
