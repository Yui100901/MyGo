package docker

import (
	"fmt"
	"github.com/Yui100901/MyGo/command"
	"github.com/Yui100901/MyGo/log_utils"
	"strings"
)

//
// @Author yfy2001
// @Date 2024/12/17 16 47
//

// ContainerStop 停止docker容器
func ContainerStop(containers ...string) error {
	log_utils.Info.Println("停止容器", containers)
	args := append([]string{"stop"}, containers...)
	_, err := command.RunCommand("docker", args...)
	return err
}

// ContainerKill 强制停止docker容器
func ContainerKill(containers ...string) error {
	log_utils.Info.Println("强制停止容器", containers)
	args := append([]string{"kill"}, containers...)
	_, err := command.RunCommand("docker", args...)
	return err
}

// ContainerRemove 删除docker容器
func ContainerRemove(containers ...string) error {
	log_utils.Info.Println("删除容器", containers)
	args := append([]string{"rm"}, containers...)
	_, err := command.RunCommand("docker", args...)
	return err
}

// ContainerInspect 获取容器详细信息
func ContainerInspect(names ...string) (string, error) {
	log_utils.Info.Println("获取容器详细信息", names)
	args := append([]string{"container", "inspect"}, names...)
	output, err := command.RunCommandOutput("docker", args...)
	return output, err
}

// ImageListFormatted 获取docker镜像列表
func ImageListFormatted() (string, error) {
	log_utils.Info.Println("列出格式化的镜像列表")
	args := []string{"images", "--format", "{{.Repository}}:{{.Tag}}"}
	output, err := command.RunCommandOutput("docker", args...)
	return output, err
}

// ImageRemove 删除docker镜像
func ImageRemove(images ...string) error {
	log_utils.Info.Println("删除镜像", images)
	args := append([]string{"rmi"}, images...)
	_, err := command.RunCommand("docker", args...)
	return err
}

// BuildImage 构建Docker镜像
func BuildImage(name string) error {
	log_utils.Info.Println("构建镜像", name)
	args := []string{"build", "-t", name, "."}
	_, err := command.RunCommand("docker", args...)
	return err
}

// Save 导出Docker镜像
func Save(name, path string) error {
	log_utils.Info.Println("导出镜像", name)
	sanitizedFilename := strings.ReplaceAll(name, ":", "_")
	sanitizedFilename = strings.ReplaceAll(sanitizedFilename, "/", "_")
	filename := fmt.Sprintf("%s/%s.tar", path, sanitizedFilename)
	args := []string{"save", "-o", filename, name}
	_, err := command.RunCommand("docker", args...)
	return err
}

// Load 导入Docker镜像
func Load(path string) error {
	log_utils.Info.Println("导入镜像", path)
	args := []string{"load", "-i", path}
	_, err := command.RunCommand("docker", args...)
	return err
}

// ImagePrune 清理docker镜像
func ImagePrune() error {
	log_utils.Info.Println("清理镜像")
	args := []string{"image", "prune", "-f"}
	_, err := command.RunCommand("docker", args...)
	return err
}

// DefaultRun 默认启动Docker容器
func DefaultRun(name string, ports ...string) error {
	log_utils.Info.Println("默认启动", name)
	args := []string{
		"run",
		"-d",
		"--name", name,
		"-v", "/etc/localtime:/etc/localtime:ro",
	}
	for _, p := range ports {
		args = append(args, "-p", p+":"+p)
	}
	args = append(args, name+":latest")
	_, err := command.RunCommand("docker", args...)
	return err
}

// ContainerRerun 重新创建Docker容器
func ContainerRerun(name string, ports ...string) error {
	if err := ContainerStop(name); err != nil {

	}
	if err := ContainerRemove(name); err != nil {

	}
	if err := DefaultRun(name, ports...); err != nil {

	}
	return nil
}
