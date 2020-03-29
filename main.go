package main

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"time"

	"github.com/fatih/color"
	"github.com/levigross/grequests"
	"github.com/tidwall/gjson"
	"gopkg.in/src-d/go-git.v4"
)

const dirname = "/tmp/ra1n-helper"
const giturl = "https://aur.archlinux.org/"

func help() {
	fmt.Println("ra1nst0rm3d AUR helper")
	fmt.Println("Usage:	name_of_package: Search for name")
	fmt.Println("--resume			Resume build(don't cloning)")

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
	var url, json string
	var i int
	var count int64
	var resp *grequests.Response
	var err error
	var pack *grequests.RequestOptions
	if len(args) != 0 && args[0] == "--resume" {
		goto makepkg
	}
	pack = &grequests.RequestOptions{
		Params: map[string]string{"v": "5",
			"type": "search",
			"arg":  args[0]},
	}
	resp, err = grequests.Get("https://aur.archlinux.org/rpc/", pack)
	if err != nil {
		color.Red("Failed to send response with error", err)
		return
	}
	// -------------------------------------------------------
	json = resp.String()
	count = gjson.Get(json, "resultcount").Int()
	fmt.Println("All results:", count)
	fmt.Println()
	// -------------------------------------------------------
	for i := 0; int64(i) < count; i++ {
		fmt.Print(i, ". ")
		str := "results." + strconv.Itoa(i) + ".Name"
		fmt.Print("Name: ")
		color.Green(gjson.Get(json, str).String())
		str = "results." + strconv.Itoa(i) + ".Description"
		fmt.Print("Desc.: ")
		color.Green(gjson.Get(json, str).String())
		str = "results." + strconv.Itoa(i) + ".Version"
		fmt.Print("Ver.: ")
		color.Green(gjson.Get(json, str).String())
		str = "results." + strconv.Itoa(i) + ".OutOfDate"
		timestamp := gjson.Get(json, str).String()
		time, err := stringToTime(timestamp)
		if err != nil {
			fmt.Println("Out of date: none")
		} else {
			fmt.Print("Out of date: ")
			color.Red(time.String())
		}
		str = "results." + strconv.Itoa(i) + ".LastModified"
		timestamp = gjson.Get(json, str).String()
		time, _ = stringToTime(timestamp)
		fmt.Print("Last modified: ")
		color.Green(time.String())
		fmt.Println()
	}
	fmt.Println(" ")
	// --------------------------------------------------------------
	fmt.Print("Choose once [0-", count-1, "]: ")
	fmt.Scanf("%d", &i)
	// --------------------------------------------------------------
	os.RemoveAll(dirname)
	os.Mkdir(dirname, 0777)
	// --------------------------------------------------------------
	url = giturl + gjson.Get(json, "results."+strconv.Itoa(i)+".Name").String() + ".git"
	_, err = git.PlainClone(dirname, false, &git.CloneOptions{
		URL:      url,
		Progress: os.Stdout,
	})
	if err != nil {
		color.Red("[ERR] Cloning error! Maybe you disconnected from Internet?")
		return
	}
	// --------------------------------------------------------------
makepkg:
	os.Chdir(dirname)
	color.Set(color.FgGreen)
	fmt.Print("Maybe you want to edit PKGBUILD?[y/n] ")
	color.Unset()
ret:
	fmt.Scanf("%s", &url)
	if url == "y" {
		if proc, err := Start("nano", "PKGBUILD"); err == nil {
			proc.Wait()
		}
	} else if url != "n" {
		color.Red("Failed to understand you, retry...")
		goto ret
	}
	if args[0] == "--resume" {
		if proc, err := Start("makepkg", "-si", "--nocheck"); err == nil {
			proc.Wait()
		}
	} else {
		if proc, err := Start("makepkg", "-i", "-s"); err == nil {
			proc.Wait()
		}
	}
}
