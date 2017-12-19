package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/colorless49/ping"
)

//SSAccount is one account
type SSAccount struct {
	Health   int
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
*/
type ClientMultiServer struct {
	LocalPort      int         `json:"local_port"`
	ServerPassword [][3]string `json:"server_password"`
}

func main() {

	cms := ClientMultiServer{}
	cms.LocalPort = 1080
	ssa := readFromWeb()
	//ssa := readFromFile("data.txt")
	sort.Slice(ssa, func(i, j int) bool { return ssa[i].PingTime < ssa[j].PingTime })
	for i, v := range ssa {
		if i == 10 {
			break
		}
		fmt.Println(v)
		cms.ServerPassword = append(cms.ServerPassword, [3]string{v.IP + ":" + v.Port, v.Password, v.Method})
	}
	jsonstr, _ := json.Marshal(cms)
	ioutil.WriteFile("client-multi-server.json", jsonstr, 0644)
	fmt.Println("Then end.")
}

func readFromWeb() []SSAccount {
	now := (time.Now().UnixNano()) / 1000000
	url := "https://free-ss.site/ss.php?_=" + strconv.FormatInt(now, 10)
	fmt.Println(url)

	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	ssAounts := make(map[string][][]interface{})
	e2 := json.Unmarshal(body, &ssAounts)
	if e2 != nil {
		panic(e2)
	}
	ss := make([]SSAccount, 0, 70)
	for _, v := range ssAounts["data"] {
		idx := v[0].(float64)
		accout := SSAccount{}
		accout.Health = int(idx)
		accout.IP = v[1].(string)
		accout.Port = v[2].(string)
		accout.Password = v[3].(string)
		accout.Method = v[4].(string)
		accout.Verified = v[5].(string)
		accout.Geo = v[6].(string)
		if accout.Health == 100 && supportEncryption(accout.Method) {
			t, err := testTime(accout.IP)
			if err != nil {
				continue
			}
			fmt.Println("Ping ", accout.IP, " duration ", t, " ms")
			accout.PingTime = t
			ss = append(ss, accout)
		}
	}
	return ss
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
		if s.Health == 100 && supportEncryption(s.Method) {
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
		ssa.Health, _ = strconv.Atoi(el[0])
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
