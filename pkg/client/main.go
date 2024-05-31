package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type ResItem struct {
	Prefix string      `json:"prefix"`
	Data   interface{} `json:"data"`
}

func main() {
	url := "http://127.0.0.1:8080/v1/data/test/list"
	res, err := http.Get(url)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return
	}
	data := []*ResItem{}
	_ = json.Unmarshal(body, &data)
	fmt.Println(data[0].Data)
}
