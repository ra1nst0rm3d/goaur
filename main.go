package main

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/levigross/grequests"
	"github.com/tidwall/gjson"
)

func help() {
	fmt.Println("ra1nst0rm3d AUR helper")
	fmt.Println("Usage:	")

}

func stringToTime(s string) (time.Time, error) {
	sec, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return time.Time{}, err
	}
	return time.Unix(sec, 0), nil
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
	count := gjson.Get(json, "resultscount").Int()
	fmt.Println("All results:", count)
	for i := 0; int64(i) < count; i++ {
		str := "results." + string(i) + ".Name"
		fmt.Println(i)
		fmt.Println("Name:", gjson.Get(json, str))
		str = "results." + string(i) + ".Version"
		fmt.Println("Ver.:", gjson.Get(json, str))
		str = "results." + string(i) + ".OutOfDate"
		if gjson.Get(json, str).Int() == int64(0) {
			fmt.Println("OutOfDate: null")
		} else {
			timestamp := gjson.Get(json, str).String()
			fmt.Println("OutOfDate:", stringToTime(timestamp))
		}
		fmt.Println(" ")
	}

	// --------------------------------------------------------------

}
