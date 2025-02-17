package builder

import (
	"fmt"
	"github.com/Yui100901/MyGo/command/docker"
	"github.com/Yui100901/MyGo/command/git"
	"github.com/Yui100901/MyGo/log_utils"
	"os"
	"path/filepath"
)

//
// @Author yfy2001
// @Date 2024/12/17 17 12
//

// Repository 结构体定义: 存储仓库信息
type Repository struct {
	URL    string `json:"url"`
	Branch string `json:"branch"`
}

func NewRepository(url, branch string) *Repository {
	return &Repository{URL: url, Branch: branch}
}

// Clone 克隆仓库到指定路径
func (repo *Repository) Clone(path string) {
	if err := git.CloneLatest(repo.URL, repo.Branch, path); err != nil {
		log_utils.Error.Println(err)
	}
}

// Update 拉取最新的仓库更改
func (repo *Repository) Update() {
	if err := git.Pull(); err != nil {
		log_utils.Error.Println("Update project failed,try to build!")
	}
}

// Project 结构体定义: 存储构建器信息
type Project struct {
	Path         string        `json:"path"`
	Name         string        `json:"name"`
	Ports        []string      `json:"ports"`
	Repository   Repository    `json:"repository"`
	BuildMessage string        `json:"build_message"`
	BuilderList  []BuilderItem `json:"-"`
}

type BuilderItem struct {
	FileType string
	Builder  Builder
}

func NewProject(path, name string, ports []string, url, branch string) *Project {
	repo := NewRepository(url, branch)
	project := &Project{
		Path:         path,
		Name:         name,
		Ports:        ports,
		Repository:   *repo,
		BuildMessage: "",
	}
	project.initInfo()
	return project
}

// 初始化构建器信息
func (p *Project) initInfo() {
	log_utils.Info.Printf("初始化构建器！\n")
	log_utils.Info.Printf("项目路径：%s，项目名：%s\n", p.Path, p.Name)
	log_utils.Info.Printf("项目地址：%s，项目分支：%s\n", p.Repository.URL, p.Repository.Branch)
}

// GetSourceCode 克隆或拉取仓库
func (p *Project) GetSourceCode() {
	if _, err := os.Stat(p.Path); os.IsNotExist(err) {
		// 项目目录不存在
		if err := os.MkdirAll(p.Path, 0755); err != nil {
			log_utils.Error.Fatalf("创建路径失败：%v", err)
		} else {
			log_utils.Error.Printf("路径创建成功：%s\n", p.Path)
		}
	}
	// 项目目录存在
	if _, err := os.Stat(filepath.Join(p.Path, ".git")); err == nil {
		// .git存在，进入项目目录，并获取最新代码
		log_utils.Info.Println("拉取最新代码")
		os.Chdir(p.Path)
		p.Repository.Update()
	} else {
		// .git不存在
		if p.Repository.URL != "" {
			// 项目地址不为空
			log_utils.Info.Printf("克隆仓库 %s\n", p.Path)
			p.Repository.Clone(p.Path)
		}
	}
}

// InitBuilder 初始化构建器
func (p *Project) InitBuilder() {
	pathStr := p.Path
	builderList := []BuilderItem{
		{"pom.xml", NewMaven(pathStr)},
		{"build.gradle", NewGradle(pathStr)},
		{"requirements.txt", NewPython(pathStr)},
		{"package.json", NewNode(pathStr)},
		{"go.mod", NewGo(pathStr)},
		{"CMakeLists.txt", NewC(pathStr)},
		{"Cargo.toml", NewRust(pathStr)},
		{"Dockerfile", NewDocker(pathStr, p.Name)},
	}
	for _, builderItem := range builderList {
		if _, err := os.Stat(filepath.Join(pathStr, builderItem.FileType)); err == nil {
			log_utils.Info.Printf("发现文件 %s。\n", builderItem.FileType)
			p.BuilderList = append(p.BuilderList, builderItem)
		}
	}
}

// Build 构建项目
func (p *Project) Build(config BuildConfig) {
	os.Chdir(p.Path)
	if len(p.BuilderList) == 0 {
		log_utils.Error.Println("没有找到任何可构建的文件！")
		return
	}
	for _, builderItem := range p.BuilderList {
		if err := builderItem.Builder.Build(config); err != nil {
			log_utils.Warn.Println("构建产生警告或错误！")
		}
	}
	log_utils.Info.Printf("构建项目 %s 结束。\n", p.Name)
	p.BuildMessage = fmt.Sprintf("%s", p.Name)
}

// DeployToDocker 部署到docker
func (p *Project) DeployToDocker() {
	if !p.hasDockerfile() {
		log_utils.Info.Printf("项目 %s 没有对应的 Dockerfile 文件，无法部署！\n", p.Name)
		return
	}
	var portList []string
	for _, port := range p.Ports {
		portList = append(portList, port)
	}
	if err := docker.ContainerRerun(p.Name, portList...); err != nil {
		log_utils.Error.Fatalf("启动docker容器出错: %v", err)
	}
}

// 判断是否存在 Dockerfile
func (p *Project) hasDockerfile() bool {
	for _, item := range p.BuilderList {
		if item.FileType == "Dockerfile" {
			return true
		}
	}
	return false
}
