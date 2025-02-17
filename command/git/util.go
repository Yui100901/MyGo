package git

import (
	"github.com/Yui100901/MyGo/command"
)

//
// @Author yfy2001
// @Date 2024/12/17 16 57
//

func CloneDefault(url, branch, dir string) error {
	args := []string{"clone", "--branch", branch, url, dir}
	return command.RunCommand("git", args...)
}

func CloneSingleBranch(url, branch, dir string) error {
	args := []string{"clone",
		"--single-branch", "--branch", branch,
		url, dir}
	return command.RunCommand("git", args...)
}

func CloneLatest(url, branch, dir string) error {
	args := []string{"clone",
		"--single-branch", "--branch", branch,
		"--depth", "1", url, dir}
	return command.RunCommand("git", args...)
}

func Pull() error {
	return command.RunCommand("git", "pull")
}
