package main

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"time"

	"github.com/levigross/grequests"
	"github.com/tidwall/gjson"
	"gopkg.in/src-d/go-git.v4"
)

const dirname = "/tmp/ra1n-helper"
const giturl = "https://aur.archlinux.org/"

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

func Start(args ...string) (p *os.Process, err error) {
	if args[0], err = exec.LookPath(args[0]); err == nil {
		var procAttr os.ProcAttr
		procAttr.Files = []*os.File{os.Stdin,
			os.Stdout, os.Stderr}
		p, err := os.StartProcess(args[0], args, &procAttr)
		if err == nil {
			return p, nil
		}
	}
	return nil, err
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
	// -------------------------------------------------------
	var json string
	json = resp.String()
	count := gjson.Get(json, "resultcount").Int()
	fmt.Println("All results:", count)
	fmt.Println()
	// -------------------------------------------------------
	for i := 0; int64(i) < count; i++ {
		str := "results." + strconv.Itoa(i) + ".Name"
		fmt.Println("Name:", gjson.Get(json, str))
		str = "results." + strconv.Itoa(i) + ".Description"
		fmt.Println("Desc.:", gjson.Get(json, str))
		str = "results." + strconv.Itoa(i) + ".Version"
		fmt.Println("Ver.:", gjson.Get(json, str))
		str = "results." + strconv.Itoa(i) + ".OutOfDate"
		timestamp := gjson.Get(json, str).String()
		time, err := stringToTime(timestamp)
		if err != nil {
			fmt.Println("Out of date: null")
		} else {
			fmt.Println("Out of date:", time)
		}
		str = "results." + strconv.Itoa(i) + ".LastModified"
		timestamp = gjson.Get(json, str).String()
		time, _ = stringToTime(timestamp)
		fmt.Println("Last modified: ", time)
		fmt.Println()
	}
	fmt.Println(" ")
	// --------------------------------------------------------------
	fmt.Print("Choose once [1-", count, "]: ")
	var i int
	fmt.Scanf("%d", &i)
	// --------------------------------------------------------------
	os.RemoveAll(dirname)
	os.Mkdir(dirname, 0777)
	// --------------------------------------------------------------

	url := giturl + gjson.Get(json, "results."+strconv.Itoa(i)+".Name").String() + ".git"
	_, err = git.PlainClone(dirname, false, &git.CloneOptions{
		URL:      url,
		Progress: os.Stdout,
	})
	if err != nil {
		fmt.Println("[ERR] Cloning error! Maybe you disconnected from Internet?")
	}
	// --------------------------------------------------------------
	os.Chdir(dirname)
	if proc, err := Start("makepkg", "-i", "-s"); err == nil {
		proc.Wait()
	}
}
