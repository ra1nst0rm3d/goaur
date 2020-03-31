package main

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"time"

	"github.com/tidwall/gjson"
)

// Main constants
const dirname = "/tmp/ra1n-helper"
const giturl = "https://aur.archlinux.org/"
const rpc = "https://aur.archlinux.org/rpc/?v=5&type=search&arg="
const red = "\u001b[31m"
const green = "\u001b[32m"
const reset = "\u001b[0m"

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

//---------------------MAIN FUNCTION-------------------------
func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		help()
		return
	}
	var url, json string
	var i int
	var count int64
	var err error
	var response *http.Response
	buf := new(bytes.Buffer)
	if len(args) != 0 && args[0] == "--resume" {
		goto makepkg
	}
	//--------------------CONNECTING TO RPC------------------------
	response, err = http.Get(rpc + args[0])
	if err != nil {
		fmt.Println("[ERR] Cannot connect to RPC interface. Check ur Internet connection")
	}
	buf.ReadFrom(response.Body)
	json = buf.String()
	//----------------SWAP TO JSON----------------------------

	count = gjson.Get(json, "resultcount").Int()
	fmt.Println("All results:", count)
	fmt.Println()
	//----------------PARSING---------------------------------
	for i := 0; int64(i) < count; i++ {
		fmt.Print(i, ". ")
		str := "results." + strconv.Itoa(i) + ".Name"
		fmt.Print("Name: ")
		fmt.Println(green + gjson.Get(json, str).String() + reset)
		str = "results." + strconv.Itoa(i) + ".Description"
		fmt.Print("Desc.: ")
		fmt.Println(green + gjson.Get(json, str).String() + reset)
		str = "results." + strconv.Itoa(i) + ".Version"
		fmt.Print("Ver.: ")
		fmt.Println(green + gjson.Get(json, str).String() + reset)
		str = "results." + strconv.Itoa(i) + ".OutOfDate"
		timestamp := gjson.Get(json, str).String()
		time, err := stringToTime(timestamp)
		if err != nil {
			fmt.Println("Out of date: none")
		} else {
			fmt.Print("Out of date: ")
			fmt.Println(red + time.String() + reset)
		}
		str = "results." + strconv.Itoa(i) + ".LastModified"
		timestamp = gjson.Get(json, str).String()
		time, _ = stringToTime(timestamp)
		fmt.Print("Last modified: ")
		fmt.Println(green + time.String() + reset)
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
	if proc, err := Start("git", "clone", giturl+gjson.Get(json, "results."+strconv.Itoa(i)+".Name").String()+".git", dirname); err == nil {
		proc.Wait()
	}
	// --------------------------------------------------------------
makepkg:
	os.Chdir(dirname)
	fmt.Print(green + "Maybe you want to edit PKGBUILD?[y/N] " + reset)
ret:
	fmt.Scanf("%s", &url)
	switch url {
	case "":
		break
	case "y":
		if proc, err := Start("nano", "PKGBUILD"); err == nil {
			proc.Wait()
		}
		break
	default:
		fmt.Print(red + "Failed to understand you, retry: " + reset)
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
