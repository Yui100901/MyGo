package docker

import (
	"fmt"
	"path/filepath"
	"strings"
)

//
// @Author yfy2001
// @Date 2024/12/17 16 53
//

// 定义 Mount 结构体
type Mount struct {
	Source      string `json:"Source"`
	Destination string `json:"Destination"`
	Mode        string `json:"Mode"`
}

// 定义 PortBinding 结构体
type PortBinding struct {
	HostPort string `json:"HostPort"`
}

// 定义 RestartPolicy 结构体
type RestartPolicy struct {
	Name string `json:"Name"`
}

// 定义 HostConfig 结构体
type HostConfig struct {
	PortBindings    map[string][]PortBinding `json:"PortBindings"`
	RestartPolicy   RestartPolicy            `json:"RestartPolicy"`
	AutoRemove      bool                     `json:"AutoRemove"`
	Privileged      bool                     `json:"Privileged"`
	PublishAllPorts bool                     `json:"PublishAllPorts"`
}

// 定义 Config 结构体
type Config struct {
	User  *string  `json:"User"`
	Env   []string `json:"Env"`
	Cmd   []string `json:"Cmd"`
	Image string   `json:"Image"`
}

// 定义 ContainerInfo 结构体
type ContainerInfo struct {
	Name       string     `json:"Name"`
	Config     Config     `json:"Config"`
	HostConfig HostConfig `json:"HostConfig"`
	Mounts     []Mount    `json:"Mounts"`
}

// ContainerInfo 的方法实现
func (ci *ContainerInfo) ParseContainerName() string {
	return strings.TrimPrefix(ci.Name, "/")
}

func (ci *ContainerInfo) ParsePrivileged() bool {
	return ci.HostConfig.Privileged
}

func (ci *ContainerInfo) ParsePublishAllPorts() bool {
	return ci.HostConfig.PublishAllPorts
}

func (ci *ContainerInfo) ParseAutoRemove() bool {
	return ci.HostConfig.AutoRemove
}

func (ci *ContainerInfo) ParseUser() string {
	if ci.Config.User != nil && *ci.Config.User != "" {
		return *ci.Config.User
	}
	return ""
}

func (ci *ContainerInfo) ParseEnvs() []string {
	return ci.Config.Env
}

func (ci *ContainerInfo) ParseMounts() []string {
	var mounts []string
	for _, mount := range ci.Mounts {
		if !filepath.IsAbs(mount.Destination) {
			// 非绝对路径时挂载匿名卷
			mounts = append(mounts, mount.Destination)
		} else {
			volume := fmt.Sprintf("%s:%s%s", mount.Source, mount.Destination,
				func() string {
					if mount.Mode == "" {
						return ""
					}
					return fmt.Sprintf(":%s", mount.Mode)
				}())
			mounts = append(mounts, volume)
		}
	}
	return mounts
}

func (ci *ContainerInfo) ParsePortBindings() []string {
	var portBindings []string
	for port, bindings := range ci.HostConfig.PortBindings {
		for _, binding := range bindings {
			portBindings = append(portBindings, fmt.Sprintf("%s:%s", binding.HostPort, port))
		}
	}
	return portBindings
}

func (ci *ContainerInfo) ParseRestartPolicy() string {
	return fmt.Sprintf("--restart=%s", ci.HostConfig.RestartPolicy.Name)
}

func (ci *ContainerInfo) ParseImage() string {
	return ci.Config.Image
}

// 定义 DockerCommand 结构体
type DockerCommand struct {
	ContainerName   string
	Privileged      bool
	PublishAllPorts bool
	AutoRemove      bool
	RestartPolicy   string
	User            string
	Envs            []string
	Mounts          []string
	PortBindings    []string
	Image           string
}

// 从 ContainerInfo 创建 DockerCommand 实例
func NewDockerCommand(info *ContainerInfo) *DockerCommand {
	return &DockerCommand{
		ContainerName:   info.ParseContainerName(),
		Privileged:      info.ParsePrivileged(),
		PublishAllPorts: info.ParsePublishAllPorts(),
		AutoRemove:      info.ParseAutoRemove(),
		RestartPolicy:   info.ParseRestartPolicy(),
		User:            info.ParseUser(),
		Envs:            info.ParseEnvs(),
		Mounts:          info.ParseMounts(),
		PortBindings:    info.ParsePortBindings(),
		Image:           info.ParseImage(),
	}
}

// 将 DockerCommand 转换为命令行参数
func (dc *DockerCommand) ToCommand() []string {
	var command []string
	command = append(command, "docker", "run", "-d")
	command = append(command, "--name", dc.ContainerName)
	if dc.Privileged {
		command = append(command, "--privileged")
	}
	if dc.PublishAllPorts {
		command = append(command, "-P")
	}
	if dc.AutoRemove {
		command = append(command, "--rm")
	}
	if dc.User != "" {
		command = append(command, "-u", dc.User)
	}
	for _, env := range dc.Envs {
		command = append(command, "-e", env)
	}
	for _, mount := range dc.Mounts {
		command = append(command, "-v", mount)
	}
	for _, portBinding := range dc.PortBindings {
		command = append(command, "-p", portBinding)
	}
	command = append(command, dc.Image)
	return command
}