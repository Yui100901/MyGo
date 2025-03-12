package builder

import (
	"github.com/Yui100901/MyGo/command"
	"github.com/Yui100901/MyGo/command/docker"
	"github.com/Yui100901/MyGo/log_utils"
)

//
// @Author yfy2001
// @Date 2024/12/17 17 00
//

type BuildConfig struct {
	NodeConfig NodeConfig
}

type NodeConfig struct {
	NodePackage        string //node打包工具
	NodePackageVersion string //node打包工具版本
	BuildTag           string
}

// Builder 接口
type Builder interface {
	Build(config *BuildConfig) error
}

// Maven 构建器结构体
type Maven struct {
	Path string
}

func NewMaven(path string) *Maven {
	return &Maven{Path: path}
}

func (m *Maven) Build(config *BuildConfig) error {
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

func (g *Gradle) Build(config *BuildConfig) error {
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

func (p *Python) Build(config *BuildConfig) error {
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

func (n *Node) Build(config *BuildConfig) error {
	log_utils.Info.Println("构建Node项目")
	var nodeCmd string
	switch config.NodeConfig.NodePackage {
	default:
		nodeCmd = "npm"
		err := NpmInstall(config)
		if err != nil {
			return err
		}
	case "pnpm":
		nodeCmd = "pnpm"
		err := PnpmInstall(config)
		if err != nil {
			return err
		}
	}
	buildCommand := "build"
	if config.NodeConfig.BuildTag != "" {
		buildCommand += ":" + config.NodeConfig.BuildTag
	}
	return command.RunCommand(nodeCmd, "run", buildCommand)
}

func NpmInstall(config *BuildConfig) error {
	return command.RunCommand("npm", "install", "--registry=https://registry.npmmirror.com")

}

func PnpmInstall(config *BuildConfig) error {
	var targetPnpm string
	if config.NodeConfig.NodePackageVersion != "" {
		targetPnpm = "pnpm" + "@" + config.NodeConfig.NodePackageVersion
	}
	if err := command.RunCommand("npm", "install", "-g", targetPnpm, "--registry=https://registry.npmmirror.com"); err != nil {
		return err
	}
	return command.RunCommand("pnpm", "install")
}

// Go 构建器结构体
type Go struct {
	Path string
}

func NewGo(path string) *Go {
	return &Go{Path: path}
}

func (g *Go) Build(config *BuildConfig) error {
	log_utils.Info.Println("构建Go项目")
	if err := command.RunCommand("go", "env", "-w", "GO111MODULE=on"); err != nil {
		return err
	}
	if err := command.RunCommand("go", "env", "-w", "GOPROXY=https://goproxy.cn,direct"); err != nil {
		return err
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

func (c *C) Build(config *BuildConfig) error {
	log_utils.Info.Println("构建C项目")
	if err := command.RunCommand("cmake", ".."); err != nil {
		return err
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

func (r *Rust) Build(config *BuildConfig) error {
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

func (d *Docker) Build(config *BuildConfig) error {
	log_utils.Info.Println("构建Docker项目")
	return docker.BuildImage(d.Name)
}
