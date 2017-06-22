package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"text/template"
)

var (
	gitHash = "unset"

	versionTemplate = `package {{.PackageName}}

type versionInfo struct {
	gitReference string
	gitTag string
	gitBranch string
	gitUsername string
	gitUserEmail string
	versionString string
}

var (
	Version = versionInfo {
	  gitReference: "{{.Hash}}"
	  versionString: "{{.Version}}"
	  gitTag: "{{.GitTag}}"
	  gitBranch: "{{.GitBranch}}"

	  gitUsername: "{{.GitUsername}}"
	  gitUserEmail: "{{.GitUserEmail}}"
	}
)

func (v versionInfo) Version() string {
	return v.versionString
}

func (v versionInfo) BuildUser() string {
	return v.gitUsername + "<" + v.gitUserEmail + ">"
}

func (v versionInfo) GitTag() string {
	return v.gitTag
}

func (v versionInfo) GitBranch() string {
	return v.gitBranch
}

func (v versionInfo) GitRevision() string {
	return v.gitReference
}
`
)

var (
	packageNameFlag = flag.String("package.name", "main", "set the package name used in the template")
)

type versionInformation struct {
	Hash         string
	Version      string
	PackageName  string
	GitTag       string
	GitBranch    string
	GitUsername  string
	GitUserEmail string
}

func GenerateVersion(outputFile string, packageName string) {
	gitHashCmd := exec.Command("git", "rev-parse", "HEAD")
	out, err := gitHashCmd.CombinedOutput()
	if err != nil {
		log.Panicf("Failed to get git hash: %+v", err)
	}

	gitHash = strings.TrimSpace(string(out))

	gitTagCommand := exec.Command("git", "describe", "--tags", "--exact-match")

	gitVersion := "unset"

	var gitBranch string
	gitTag := "unset"

	out, err = gitTagCommand.CombinedOutput()
	if err == nil {
		gitTag = strings.TrimSpace(string(out))
	}
	gitBranchCommand := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	out, err = gitBranchCommand.CombinedOutput()
	if err != nil {
		log.Panicf("Failed to get branch name or tag name: %+v", err)
	}
	gitBranch = strings.TrimSpace(string(out))

	if gitTag != "unset" {
		gitVersion = gitTag
	} else {
		gitVersion = gitBranch
	}

	var gitUsername string
	var gitUserEmail string

	gitUsernameCmd := exec.Command("git", "config", "user.name")
	out, err = gitUsernameCmd.CombinedOutput()
	if err != nil {
		gitUsername = "unset"
	} else {
		gitUsername = strings.TrimSpace(string(out))
	}

	gitUserEmailCmd := exec.Command("git", "config", "user.email")
	out, err = gitUserEmailCmd.CombinedOutput()
	if err != nil {
		gitUserEmail = "unset"
	} else {
		gitUserEmail = strings.TrimSpace(string(out))
	}

	version := fmt.Sprintf("%s-%s", gitVersion, gitHash)

	versionTmpl := template.Must(template.New("version").Parse(versionTemplate))

	versionFile, err := os.OpenFile(outputFile, os.O_CREATE|os.O_WRONLY, 0777)
	if err != nil {
		log.Panicf("Unable to create version file: %+v", err)
	}

	err = versionTmpl.Execute(versionFile, versionInformation{
		Hash:         gitHash,
		Version:      version,
		PackageName:  packageName,
		GitTag:       gitVersion,
		GitBranch:    gitBranch,
		GitUsername:  gitUsername,
		GitUserEmail: gitUserEmail,
	})
	if err != nil {
		log.Panicf("Failed to execute template: %+v", err)
	}
}

func main() {
	flag.Parse()

	fileName := flag.Arg(0)

	if fileName == "" {
		log.Panicln("No output file set")
	}

	GenerateVersion(flag.Arg(0), *packageNameFlag)
}
