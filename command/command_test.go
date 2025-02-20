package command

import "testing"

//
// @Author yfy2001
// @Date 2025/2/20 09 40
//

func TestRunCommand(t *testing.T) {
	err := RunCommand("echo", "hello world")
	if err != nil {
		t.Log(err.Error())
		return
	}

}

func TestRunCommandOutput(t *testing.T) {
	output, err := RunCommandOutput("echo", "hello world")
	if err != nil {
		t.Log(err.Error())
		return
	}
	t.Log(output)
}
