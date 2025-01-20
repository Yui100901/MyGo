package git

import "github.com/Yui100901/MyGo/pkg/command"

//
// @Author yfy2001
// @Date 2024/12/17 16 57
//

func CloneDefault(url, branch, dir string) error {
	args := []string{"clone", "--branch", branch, url, dir}
	_, err := command.RunCommand("git", args...)
	return err
}

func CloneSingleBranch(url, branch, dir string) error {
	args := []string{"clone",
		"--single-branch", "--branch", branch,
		url, dir}
	_, err := command.RunCommand("git", args...)
	return err
}

func CloneLatest(url, branch, dir string) error {
	args := []string{"clone",
		"--single-branch", "--branch", branch,
		"--depth", "1", url, dir}
	_, err := command.RunCommand("git", args...)
	return err
}

func Pull() error {
	_, err := command.RunCommand("git", "pull")
	return err
}
