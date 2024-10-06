package internal

import (
	"embed"
	"fmt"
	"log"
	"text/template"

	"github.com/shenghui0779/yiigo"
	"github.com/shenghui0779/yiigo/cmd/internal/grpc"
	"github.com/shenghui0779/yiigo/cmd/internal/http"
)

type Params struct {
	Module  string
	AppPkg  string
	AppName string
}

func InitHttpProject(root, mod string, apps ...string) {
	// 创建项目
	initProject(root, mod, apps, http.Project, http.FS)
	// 创建App(单应用)
	if len(apps) == 0 {
		initApp(root, mod, "", http.App, http.FS)
		return
	}
	// 创建App(多应用)
	for _, name := range apps {
		initApp(root, mod, name, http.App, http.FS)
	}
}

func InitHttpApp(root, mod, name string) {
	initApp(root, mod, name, http.App, http.FS)
}

func InitGrpcProject(root, mod string, apps ...string) {
	// 创建项目
	initProject(root, mod, apps, grpc.Project, grpc.FS)
	// 创建App(单应用)
	if len(apps) == 0 {
		initApp(root, mod, "", grpc.App, grpc.FS)
		return
	}
	// 创建App(多应用)
	for _, name := range apps {
		initApp(root, mod, name, grpc.App, grpc.FS)
	}
}

func InitGrpcApp(root, mod, name string) {
	initApp(root, mod, name, grpc.App, grpc.FS)
}

func initProject(root, mod string, apps []string, tmpls []map[string]string, fs embed.FS) {
	params := &Params{Module: mod}
	// 创建项目
	for _, tmpl := range tmpls {
		output := root + "/" + tmpl["output"]
		buildTmpl(fs, tmpl["name"], tmpl["path"], output, params)
	}
}

func initApp(root, mod, name string, tmpls []map[string]string, fs embed.FS) {
	prefix := root + "/pkg/app"
	params := &Params{
		Module:  mod,
		AppPkg:  "app",
		AppName: root,
	}
	if len(name) != 0 {
		prefix += "/" + name
		params.AppPkg = "app/" + name
		params.AppName = name
	}
	for _, tmpl := range tmpls {
		output := prefix + "/" + tmpl["output"]
		buildTmpl(fs, tmpl["name"], tmpl["path"], output, params)
	}
}

func buildTmpl(fs embed.FS, name, path, output string, params *Params) {
	// 模板解析
	t, err := template.New(name).ParseFS(fs, path)
	if err != nil {
		log.Fatalln(err)
	}
	// 文件创建
	f, err := yiigo.CreateFile(output)
	if err != nil {
		log.Fatalln(err)
	}
	defer f.Close()
	// 模板执行
	if err = t.Execute(f, &params); err != nil {
		log.Fatalln(err)
	}
	fmt.Println(output)
}
