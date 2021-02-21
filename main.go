package main

import (
	"encoding/json"
	"fmt"
	"iluckin.cn/valve/source"
	"time"
)

func main() {
	s, e := source.NewQuerier("103.205.253.202:33246", 1*time.Second)
	defer s.Close()

	if e != nil {
		panic(e)
	}

	info, e := s.GetInfo()
	if e != nil {
		panic(e)
	}

	jsonBytes, e := json.Marshal(&info)

	fmt.Println(string(jsonBytes))

}
