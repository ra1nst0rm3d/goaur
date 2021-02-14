package main

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"time"

	"github.com/Jguer/go-alpm"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fastjson"
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
	var url string
	var i int
	var count int
	var err error
	var dst []byte
	var h *alpm.Handle
	var db *alpm.DB
	if len(args) != 0 && args[0] == "--resume" {
		goto makepkg
	}
	//------------------INIT ALPM----------------------------------
	h, err = alpm.Initialize("/", "/var/lib/pacman")
	if err != nil {
		fmt.Println(err)
		return
	}
	db, err = h.LocalDB()
	if err != nil {
		fmt.Println(err)
	}
	defer h.Release()
	//--------------------CONNECTING TO RPC------------------------
	count, dst, err = fasthttp.Get(dst, rpc+args[0])
	if err != nil {
		fmt.Println("[ERR] Cannot connect to RPC interface. Check ur Internet connection")
	}
	//----------------SWAP TO JSON----------------------------

	//count = gjson.Get(json, "resultcount").Int()
	count = fastjson.GetInt(dst, "resultcount")
	fmt.Println("All results:", count)
	fmt.Println()
	//----------------PARSING---------------------------------
	for i := 0; i < count; i++ {
		fmt.Print(i, ". ")
		fmt.Print("Name: ")
		fmt.Print(green + fastjson.GetString(dst, "results", strconv.Itoa(i), "Name") + reset + " ")
		for _, pkg := range db.PkgCache().Slice() {
			if fastjson.GetString(dst, "results", strconv.Itoa(i), "Name") == pkg.Name() {
				fmt.Println("<" + green + "INSTALLED" + reset + ">")
				break
			}
		}
		fmt.Println()
		fmt.Print("Desc.: ")
		fmt.Println(green + fastjson.GetString(dst, "results", strconv.Itoa(i), "Description") + reset)
		fmt.Print("Ver.: ")
		fmt.Println(green + fastjson.GetString(dst, "results", strconv.Itoa(i), "Version") + reset)
		timestamp := strconv.Itoa(fastjson.GetInt(dst, "results", strconv.Itoa(i), "OutOfDate"))
		time, err := stringToTime(timestamp)
		if timestamp == strconv.Itoa(0) || err != nil {
			fmt.Println("Out of date: none")
		} else {
			fmt.Print("Out of date: ")
			fmt.Println(red + time.String() + reset)
		}
		timestamp = strconv.Itoa(fastjson.GetInt(dst, "results", strconv.Itoa(i), "LastModified"))
		time, _ = stringToTime(timestamp)
		fmt.Print("Last modified: ")
		fmt.Println(green + time.String() + reset)
		fmt.Println()
	}
	fmt.Println(" ")
	// --------------------CHOOSING----------------------------------
	fmt.Print("Choose once [0-", count-1, "]: ")
	fmt.Scanf("%d", &i)
	// ---------------------MAKING WORK_DIR-------------------------------
	os.RemoveAll(dirname)
	os.Mkdir(dirname, 0777)
	// ---------------------CLONING REPO--------------------------
	if proc, err := Start("git", "clone", giturl+fastjson.GetString(dst, "results", strconv.Itoa(i), "Name")+".git", dirname); err == nil {
		proc.Wait()
	}
	// ---------------------NOW TRUE MAGIC!------------------------------
makepkg:
	os.Chdir(dirname)
	fmt.Print(green + "Maybe you want to edit PKGBUILD?[y/N] " + reset)
ret:
	fmt.Scanf("%s", &url)
	switch url {
	case "":
		break
	case "n":
		break
	case "y":
		if proc, err := Start("nano", "PKGBUILD"); err == nil {
			proc.Wait()
		}
		break
	default:
		fmt.Print(red + "Failed to understand you, retry: " + reset)
		url = ""
		goto ret
	}
	if args[0] == "--resume" {
		if proc, err := Start("makepkg", "-si", "--nocheck"); err == nil {
			proc.Wait()
		}
	} else {
		if proc, err := Start("makepkg", "-sirc", "--skippgpcheck"); err == nil {
			proc.Wait()
		}
	}
}
