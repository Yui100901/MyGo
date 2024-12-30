package builder

import (
	"github.com/Yui100901/MyGo/pkg/command"
	"github.com/Yui100901/MyGo/pkg/command/docker"
	"github.com/Yui100901/MyGo/pkg/log_utils"
)

//
// @Author yfy2001
// @Date 2024/12/17 17 00
//

// Builder 接口
type Builder interface {
	Build() (string, error)
}

// Maven 构建器结构体
type Maven struct {
	Path string
}

func NewMaven(path string) *Maven {
	return &Maven{Path: path}
}

func (m *Maven) Build() (string, error) {
	log_utils.Info.Println("构建Maven项目")
	return command.RunCommand("mvn", "clean", "package")
}

// Gradle 构建器结构体
type Gradle struct {
	Path string
}

func NewGradle(path string) *Gradle {
	return &Gradle{Path: path}
}

func (g *Gradle) Build() (string, error) {
	log_utils.Info.Println("构建Gradle项目")
	return command.RunCommand("gradle", "build")
}

// Python 构建器结构体
type Python struct {
	Path string
}

func NewPython(path string) *Python {
	return &Python{Path: path}
}

func (p *Python) Build() (string, error) {
	log_utils.Info.Println("构建Python项目")
	return command.RunCommand("pip", "install", "-r", "requirements.txt", "-i", "https://pypi.tuna.tsinghua.edu.cn/simple")
}

// Node 构建器结构体
type Node struct {
	Path string
}

func NewNode(path string) *Node {
	return &Node{Path: path}
}

func (n *Node) Build() (string, error) {
	log_utils.Info.Println("构建Node项目")
	if _, err := command.RunCommand("npm", "install", "--registry=https://registry.npmmirror.com"); err != nil {
		return "", err
	}
	return command.RunCommand("npm", "run", "build")
}

// Go 构建器结构体
type Go struct {
	Path string
}

func NewGo(path string) *Go {
	return &Go{Path: path}
}

func (g *Go) Build() (string, error) {
	log_utils.Info.Println("构建Go项目")
	if _, err := command.RunCommand("go", "env", "-w", "GO111MODULE=on"); err != nil {
		return "", err
	}
	if _, err := command.RunCommand("go", "env", "-w", "GOPROXY=https://goproxy.cn,direct"); err != nil {
		return "", err
	}
	return command.RunCommand("go", "build")
}

// C 构建器结构体
type C struct {
	Path string
}

func NewC(path string) *C {
	return &C{Path: path}
}

func (c *C) Build() (string, error) {
	log_utils.Info.Println("构建C项目")
	if _, err := command.RunCommand("cmake", ".."); err != nil {
		return "", err
	}
	return command.RunCommand("make")
}

// Rust 构建器结构体
type Rust struct {
	Path string
}

func NewRust(path string) *Rust {
	return &Rust{Path: path}
}

func (r *Rust) Build() (string, error) {
	log_utils.Info.Println("构建Rust项目")
	return command.RunCommand("cargo", "build", "--release")
}

// Docker 构建器结构体
type Docker struct {
	Path string
	Name string
}

func NewDocker(path, name string) *Docker {
	return &Docker{Path: path, Name: name}
}

func (d *Docker) Build() (string, error) {
	log_utils.Info.Println("构建Docker项目")
	return "", docker.BuildImage(d.Name)
}
