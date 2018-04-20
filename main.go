package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"sort"
	"strings"

	"github.com/colorless49/ping"
)

//SSAccount is one account
type SSAccount struct {
	Health   string
	IP       string
	Port     string
	Password string
	Method   string
	Verified string
	Geo      string
	PingTime int
}

/*
{
	"local_port": 1081,
	"server_password": [
		["127.0.0.1:8387", "foobar"],
		["127.0.0.1:8388", "barfoo", "aes-128-cfb"]
	]
}
**/
type clientMultiServer struct {
	LocalPort      int         `json:"local_port"`
	ServerPassword [][3]string `json:"server_password"`
}

func main() {

	cms := clientMultiServer{}
	var brookCmd string
	cms.LocalPort = 1080
	/* 去掉直接从网站抓取
	ssa, err := readFromWeb()
	if err != nil {
		fmt.Println("read from web error:", err)
		fmt.Println("Then read from data.txt ...")
		ssa = readFromFile("data.txt")
	}*/

	//根据data.txt文件内容读取
	ssa := readFromFile("data.txt")

	sort.Slice(ssa, func(i, j int) bool { return ssa[i].PingTime < ssa[j].PingTime })
	for i, v := range ssa {
		if brookCmd == "" && v.Method == "aes-256-cfb" {
			fmt.Print("brook CMD ----")
			fmt.Println(v)
			brookCmd = "brook ssclient -l 127.0.0.1:1080 -i 127.0.0.1 -s " + v.IP + ":" + v.Port + " -p " + v.Password + " --http"
		}
		if i == 10 {
			break
		}
		fmt.Println(v)
		cms.ServerPassword = append(cms.ServerPassword, [3]string{v.IP + ":" + v.Port, v.Password, v.Method})
	}
	jsonstr, _ := json.Marshal(cms)
	if ioutil.WriteFile("client-multi-server.json", jsonstr, 0644) != nil {
		fmt.Println("写入client-multi-server.json失败。")
	}
	if ioutil.WriteFile("brook.bat", []byte(brookCmd), 0644) != nil {
		fmt.Println("写入brook.bat失败。")
	}

	fmt.Println("Then end.")
}

//解析从网站上copy出来的数据文件，如data.txt。在有验证码的情况下使用
func readFromFile(path string) []SSAccount {
	ss := make([]SSAccount, 0, 70)

	f, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	r := bufio.NewReader(f)
	for {
		b, _, err := r.ReadLine()
		if err != nil {
			if err == io.EOF {
				break
			}
			panic(err)
		}
		s, err := line2SSaccount(string(b))
		if err != nil {
			continue
		}
		if supportEncryption(s.Method) {
			t, err := testTime(s.IP)
			if err != nil {
				continue
			}
			fmt.Println("Ping ", s.IP, " duration ", t, " ms")
			s.PingTime = t

			ss = append(ss, s)
		}
	}
	return ss
}

func line2SSaccount(line string) (SSAccount, error) {
	ssa := SSAccount{}
	s := strings.TrimSpace(line)
	el := strings.Split(s, "\t")
	if len(el) == 7 {
		ssa.Health = el[0] //strconv.Atoi(el[0])
		ssa.IP = el[1]
		ssa.Port = el[2]
		ssa.Password = el[3]
		ssa.Method = el[4]
		ssa.Verified = el[5]
		ssa.Geo = el[6]
	} else {
		return ssa, errors.New("解析的参数个数不对。")
	}
	return ssa, nil
}

func testTime(address string) (int, error) {

	duration, err := ping.PingDuration(address, 1)

	return duration, err
}

func supportEncryption(method string) bool {
	encrypt := "aes-128-cfb, aes-192-cfb, aes-256-cfb, aes-128-ctr, aes-192-ctr, aes-256-ctr, des-cfb, bf-cfb, cast5-cfb, rc4-md5, chacha20, chacha20-ietf, salsa20,"
	return strings.Contains(encrypt, strings.ToLower(method))
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}
