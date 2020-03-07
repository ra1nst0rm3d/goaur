package main

import (
	"fmt"
	"os"

	"github.com/levigross/grequests"
	"github.com/tidwall/gjson"
)

func help() {
	fmt.Println("ra1nst0rm3d AUR helper")
	fmt.Println("Usage:	")

}
func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		help()
		return
	}
	pack := &grequests.RequestOptions{
		Params: map[string]string{"v": "5",
			"type": "search",
			"arg":  args[0]},
	}
	resp, err := grequests.Get("https://aur.archlinux.org/rpc/", pack)
	if err != nil {
		fmt.Println("Failed to send response with error", err)
		return
	}
	var json string
	json = resp.String()
	count := gjson.Get(json, "resultscount")
	val := gjson.Get(json, "results.0.ID")
	println("First ID:", val.String())

}
