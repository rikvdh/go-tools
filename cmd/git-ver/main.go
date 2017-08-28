package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"code.gitea.io/git"
	version "github.com/hashicorp/go-version"
	flag "github.com/spf13/pflag"
)

var commitregex = regexp.MustCompile("\"([a-z0-9]*)\"")

func commitHash() string {
	cmd := git.NewCommand("log", "-n", "1", "--pretty=\"%h\"")
	b, err := cmd.RunInDirBytes(".")
	if err != nil {
		panic("error while running git command")
	}
	commitHash := commitregex.FindStringSubmatch(string(b))
	if len(commitHash) != 2 {
		panic("commithash fetching failed")
	}
	return commitHash[1]
}

func exactTag() string {
	cmd := git.NewCommand("describe", "--tags", "--exact-match")
	b, err := cmd.RunInDirBytes(".")
	if err != nil {
		return "0.0.0"
	}

	verstr := strings.Trim(string(b), "\n")
	v, err := version.NewVersion(verstr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "version error: %v\n", err)
		return "0.0.0"
	}
	segments := v.Segments()
	if len(segments) >= 3 {
		return fmt.Sprintf("%d.%d.%d", segments[0], segments[1], segments[2])
	}
	return "0.0.0"
}

func latestTag() string {
	cmd := git.NewCommand("describe", "--abbrev=0", "--tags")
	b, err := cmd.RunInDirBytes(".")
	if err != nil {
		return "0.0.0"
	}

	verstr := strings.Trim(string(b), "\n")
	v, err := version.NewVersion(verstr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "version error: %v\n", err)
		return "0.0.0"
	}
	segments := v.Segments()
	if len(segments) >= 3 {
		return fmt.Sprintf("%d.%d.%d", segments[0], segments[1], segments[2])
	}
	return "0.0.0"
}

func workDirIsDirty() bool {
	cmd := git.NewCommand("diff", "--shortstat")
	b, err := cmd.RunInDirBytes(".")
	if err != nil || len(b) > 0 {
		return true
	}

	cmd2 := git.NewCommand("diff", "--shortstat", "--cached")
	b2, err2 := cmd2.RunInDirBytes(".")
	if err2 != nil || len(b2) > 0 {
		return true
	}

	return false
}

var (
	date bool
	nodate bool
	nodirty bool
)

func init() {
	flag.BoolVarP(&date, "date", "d", false, "always include date in version")
	flag.BoolVarP(&nodate, "no-date", "n", false, "never include date in version (goes above -d)")
	flag.BoolVarP(&nodate, "no-dirty", "f", false, "never include dirty flag")
	flag.Parse()
}

func main() {
	var appVersion, dateString, commit string

	dateString = "+" + time.Now().UTC().Format("20060102150405")

	if tag := exactTag(); tag != "0.0.0" {
		appVersion = tag
	} else {
		commit = commitHash()
		appVersion = latestTag()
		date = true
	}
	if date && !nodate {
		appVersion += dateString
	}
	if len(commit) > 0 {
		appVersion += "+git." + commit
	}
	if workDirIsDirty() && !nodirty {
		appVersion += "-dirty"
	}
	fmt.Println(appVersion)
}
