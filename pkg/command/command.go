package command

import (
	"bufio"
	"fmt"
	"log_utils"
	"os/exec"
	"runtime"
	"strings"
)

//
// @Author yfy2001
// @Date 2024/12/17 16 30
//

func NewCommand(command string, args ...string) *exec.Cmd {
	log_utils.Info.Println("创建命令:", command, strings.Join(args, " "))

	var cmd *exec.Cmd

	if runtime.GOOS == "windows" {
		// 在 Windows 上运行命令
		cmd = exec.Command("cmd", append([]string{"/C", command}, args...)...)
	} else {
		// 在 Linux/Unix 上运行命令
		cmd = exec.Command(command, args...)
	}
	return cmd
}

func RunCommand(command string, args ...string) (string, error) {

	cmd := NewCommand(command, args...)

	// 获取标准输出管道
	stdout, err := cmd.StdoutPipe()
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return "Failed", err
	}

	// 启动命令
	if err := cmd.Start(); err != nil {
		return "Failed", err
	}

	// 标准输出
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			log_utils.Info.Println(scanner.Text())
		}
	}()

	// 标准错误
	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			log_utils.Error.Println(scanner.Text())
		}
	}()

	// 等待命令执行完成
	if err := cmd.Wait(); err != nil {
		return "Failed", err
	}

	return "Success", nil
}

func RunCommandOutput(command string, args ...string) (string, error) {
	cmd := NewCommand(command, args...)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("command execution failed: %w", err)
	}
	return string(output), nil
}