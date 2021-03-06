package main

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"strings"

	"golang.org/x/sys/windows/registry"

	"github.com/gonutz/w32"
	"github.com/google/go-github/github"
)

func main() {

	version, err := checkGameVersion()
	if err != nil {
		fmt.Print("Warcraft III Not detected in system!")
		os.Exit(1)
	}

	fmt.Printf("Detected Warcraft III version %s", version)

	client := github.NewClient(nil)

	mainCommands := []string{"require"}

	if len(os.Args) < 2 {
		fmt.Print(`Ketch - Wurst Dependency Manager

Commands:
init - create Wurst map repository
up - updates dependencies (basing on wurst.deps file)
require - adds dependency`)
		os.Exit(1)
	}
	mainCommand := os.Args[1]

	nodesToCheck := []string{
		"wurst.dependencies",
		"_build",
		"_build/dependencies",
		"_build/objectEditingOutput",
		"_build/blizzard.j",
		"_build/common.j",
		"_build/compiled.j.txt",
		"_build/WurstRunMap.w3x",
	}

	err = checkFiles(nodesToCheck)

	if err != nil {
		fmt.Print(err)
		os.Exit(1)
	}

	if stringInSlice(mainCommand, mainCommands) {

		switch mainCommand {
		case "init":
			fmt.Printf("Initializing Wurst Repo")
		case "require":

			if len(os.Args) >= 3 {
				addDependency(os.Args[2], client)
			} else {
				fmt.Print("You need to provide dependency")
			}

		}

	} else {
		fmt.Printf("Unknown command \"%s\"", mainCommand)
	}

}

func addDependency(urll string, client *github.Client) {

	parsedURL, _ := url.Parse(urll)

	urlData := strings.Split(parsedURL.Path, "/")
	if len(urlData) != 3 {
		fmt.Println("Dependency should be in github.com/owner/repo form")
		os.Exit(1)
	}
	owner := urlData[1]
	repo := urlData[2]
	ctx := context.Background()
	response, err := client.Repositories.DownloadContents(ctx, owner, repo, "wurst.build", nil)
	if err != nil {
		fmt.Print("Could not find repository or it is not wurst code repository")
		os.Exit(1)
	}
	/*buf := new(bytes.Buffer)
	buf.ReadFrom(response)
	fmt.Print(buf.String())*/ //TODO: Gazzilion things, dependencies of dependencies etc, for now it's just a check if repo has a wurst code in it
	response.Close()
	fmt.Println("cloning")
	out := exec.Command("git", "clone", "https://"+urll+".git", "_build/dependencies/"+repo).Run()
	fmt.Print(out)

}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func checkFiles(nodes []string) error {
	var err error

	for _, n := range nodes {
		if err = checkIfExists(n); err != nil {
			return err
		}
	}

	return err
}

func checkIfExists(directory string) error {
	var errorMsg error
	if _, err := os.Stat(directory); os.IsNotExist(err) {
		//TODO: Check if node is dir or file and return proper error message
		return errors.New(directory + " directory missing!")
	}
	return errorMsg
}

func checkGameVersion() (version string, err error) {
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\WOW6432Node\Blizzard Entertainment\Warcraft III`, registry.ALL_ACCESS)
	s, _, err := k.GetStringValue("GamePath")

	size := w32.GetFileVersionInfoSize(s)
	if size <= 0 {
		panic("GetFileVersionInfoSize failed")
	}

	info := make([]byte, size)
	ok := w32.GetFileVersionInfo(s, info)
	if !ok {
		panic("GetFileVersionInfo failed")
	}

	fixed, ok := w32.VerQueryValueRoot(info)
	if !ok {
		panic("VerQueryValueRoot failed")
	}
	ver := fixed.FileVersion()
	version = fmt.Sprintf(
		"%d.%d.%d.%d\n",
		ver&0xFFFF000000000000>>48,
		ver&0x0000FFFF00000000>>32,
		ver&0x00000000FFFF0000>>16,
		ver&0x000000000000FFFF>>0,
	)

	return version, err
}
