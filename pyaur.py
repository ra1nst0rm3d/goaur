import sys
import json
from datetime import datetime
from urllib.request import urlopen
from termcolor import colored
from shutil import rmtree
import os
aurrpc = "https://aur.archlinux.org/rpc/?v=5&type=search&arg="
dirname = "/tmp/pyaura"
giturl = "https://aur.archlinux.org/"
class Package:
    def __init__(self):
        self.rpcurl = aurrpc
        self.args = sys.argv[1:]
        self.jsonDict = {}
    def FetchJSON(self, askedP):
        jsonContent = urlopen(self.rpcurl+askedP).read()
        self.jsonDict = json.loads(jsonContent)
    def DecodeJSON(self, word):
        return self.jsonDict.get(word)

def wrapper():
    args = sys.argv[1:]
    package = Package()
    package.FetchJSON(args[0])
    resultcount = package.DecodeJSON("resultcount")
    print("Got "+str(resultcount)+" results")
    # I don't know why, but it works, don't touch it!
    tmpdict = eval(str(package.DecodeJSON("results"))[1:][:-1])
    # Ha, sped it up 10000x times by just moving THAT previous line out of cycle
    # Eval is pretty CPU demanding though, tried literal_eval and it was literal_HELL with speed
    for i in range(resultcount):
        # This works too, DONT YOU DARE TOUCHING IT!
        resultdict = eval(str(tmpdict[i]))
        print(str(i)+". Name: "+colored(resultdict.get("Name"), "green"))
        print("Desc.: "+colored(resultdict.get("Description"), "green"))
        print("Ver.: "+colored(resultdict.get("Version"), "green"))
        if resultdict.get("OutOfDate") is None:
            print("Out of date: None")
        else:
            # That is fragile too, what it does is it converts timestamp to UTC, than UTC to date by strftime
            print("Out of date: "+colored(datetime.utcfromtimestamp(resultdict.get("OutOfDate")).strftime('%Y-%m-%d %H:%M:%S'), "red"))
        print("Last modified: "+colored(datetime.utcfromtimestamp(resultdict.get("LastModified")).strftime('%Y-%m-%d %H:%M:%S'), "green"))
    print("")
    number = int(input("Choose one candidate[0-{}]: ".format(resultcount)))
    if os.path.exists(dirname):
        rmtree(dirname)
    os.mkdir(dirname)
    os.system("git clone "+giturl+eval(str(tmpdict[number])).get("Name")+".git "+dirname)
    os.chdir(dirname)
    while True:
        url = input("MaYbE yOu WaNt tO eDiT PKGBUILD?????[y/n] ")
        if url == "y":
            os.system("nano PKGBUILD")
            break
        elif url == "n":
            break
        else:
            print("y/n")
    if "--resume" in args:
        os.system("makepkg -si --nocheck")
    else:
        os.system("makepkg -i -s")
if __name__ == "__main__":
    wrapper()


