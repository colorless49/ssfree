package main

import (
	"bufio"
	"io"
	"os"
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

func main() {

}

func readData(path string) {
	ss := &SSAccount{}

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
	}
}

func lineToSSaccount(line string) SSAccount {
	return SSAccount{}
}
