package main

import (
	"encoding/json"
	"fmt"
	"iluckin.cn/valve/source"
	"time"
)

func main() {
	s, e := source.NewQuerier("103.53.124.109:6895", 1 * time.Second)
	defer s.Close()

	if e != nil {
		panic(e)
	}

	info, e := s.GetPlayerInfo()
	if e != nil {
		panic(e)
	}

	jsonBytes, e := json.Marshal(&info)

	fmt.Println(string(jsonBytes))
}
