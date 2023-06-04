package login

import (
	Encoder "AutoCo/passwordEncoder"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"syscall"
)

var Yys = []string{"联通", "移动", "电信"}

type LoginInfo struct {
	Username string `json:"账号"`
	Password string `json:"密码"`
	Yys      string `json:"运营商"`
	Time     string `json:"时间"`
}

func (loginInfo *LoginInfo) InputInfo() {
	loginInfo.InputUsername()
	loginInfo.InputPassword()
	loginInfo.InputYys()
}

func (loginInfo *LoginInfo) InputUsername() {
	fmt.Print("校园网账号:")
	fmt.Scan(&loginInfo.Username)
}
func (loginInfo *LoginInfo) InputPassword() {
	fmt.Print("校园网密码:")
	fmt.Scan(&loginInfo.Password)
}
func (loginInfo *LoginInfo) InputYys() {
	fmt.Print("运营商(电信,联通,移动):")
	fmt.Scan(&loginInfo.Yys)
}

func (loginInfo *LoginInfo) YysIsRight() bool {
	var Yys = []string{"联通", "移动", "电信"}
	for _, element := range Yys {
		if loginInfo.Yys == element {
			return true
		}
	}
	return false
}

/*
更新数据
*/
func (info *LoginInfo) Update(fileName string) {
	info.InputInfo()
	info.WriteInfoInFile(fileName)
}

/*
判断数据是否正确
:运营商是否正确
:账号密码是否为空
*/
func (info *LoginInfo) DataIsRight() bool {
	if info.Username == "" || info.Password == "" || !info.YysIsRight() {
		return false
	}
	return true
}
func (info *LoginInfo) WriteInfoInFile(fileName string) {
	openFile, e := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 777)
	err := syscall.Chmod(fileName, 0777)
	if err != nil {
		fmt.Println(err)
		return
	}

	// 读取或者创建文件
	if e != nil {
		fmt.Println(e)
	}

	//加密
	data, _ := Encoder.RsaEncrypt([]byte(info.Password))
	data2, _ := Encoder.RsaEncrypt([]byte(info.Username))

	// 结构体复制
	info2 := LoginInfo{}
	info2.Username = base64.StdEncoding.EncodeToString((data2))
	info2.Yys = info.Yys
	info2.Password = base64.StdEncoding.EncodeToString((data))

	dataByte, _ := json.Marshal(info2) //格式化数据
	// 写入数据
	openFile.WriteString(string(dataByte))
	openFile.Close()

}

func (info *LoginInfo) GetYysCode() int {
	if info.Yys == "移动" {
		return 0
	}
	if info.Yys == "联通" {
		return 1
	}
	if info.Yys == "电信" {
		return 2
	}
	return 1
}
func (info *LoginInfo) ReadInfoInFile(fileName string) {
	openFile, e := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE, 777)
	// 修改权限 :如果没有文件就会抛出异常
	syscall.Chmod(fileName, 0777)
	// 读取或者创建文件
	if e != nil {
		fmt.Println(e)
	}

	// 读取数据
	buf := make([]byte, 1024)
	loginfile := ""
	for {
		len, _ := openFile.Read(buf)
		if len == 0 {
			break
		}
		loginfile = loginfile + string(buf[:len])
	}

	//r := []rune(loginfile) // go中string 中文站3个,因为站一个,统计长度要用rune类型
	// 格式化json数据
	json.Unmarshal([]byte(loginfile), &info)
	openFile.Close()

	//解码
	decodeString, _ := base64.StdEncoding.DecodeString(info.Password)
	origData, _ := Encoder.RsaDecrypt(decodeString)
	info.Password = string(origData)

	decodeString1, _ := base64.StdEncoding.DecodeString(info.Username)
	origData1, _ := Encoder.RsaDecrypt(decodeString1)
	info.Username = string(origData1)

}
