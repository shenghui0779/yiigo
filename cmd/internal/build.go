package internal

import (
	"embed"
	"fmt"
	"io/fs"
	"log"
	"path/filepath"
	"strings"
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
	initProject(root, mod, http.FS)
	// 创建App(单应用)
	if len(apps) == 0 {
		initApp(root, mod, "", http.FS)
		return
	}
	// 创建App(多应用)
	for _, name := range apps {
		initApp(root, mod, name, http.FS)
	}
}

func InitHttpApp(root, mod, name string) {
	initApp(root, mod, name, http.FS)
}

func InitGrpcProject(root, mod string, apps ...string) {
	// 创建项目
	initProject(root, mod, grpc.FS)
	// 创建App(单应用)
	if len(apps) == 0 {
		initApp(root, mod, "", grpc.FS)
		return
	}
	// 创建App(多应用)
	for _, name := range apps {
		initApp(root, mod, name, grpc.FS)
	}
}

func InitGrpcApp(root, mod, name string) {
	initApp(root, mod, name, grpc.FS)
}

func initProject(root, mod string, fsys embed.FS) {
	params := &Params{Module: mod}
	// root
	files, _ := fs.ReadDir(fsys, ".")
	for _, v := range files {
		if v.IsDir() || filepath.Ext(v.Name()) == ".go" {
			continue
		}
		output := buildOutput(root, v.Name(), "")
		buildTmpl(fsys, v.Name(), output, params)
	}
	// pkg/lib
	_ = fs.WalkDir(fsys, "pkg/lib", func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() || filepath.Ext(path) == ".go" {
			return nil
		}
		output := buildOutput(root, path, "")
		buildTmpl(fsys, path, output, params)
		return nil
	})
}

func initApp(root, mod, name string, fsys embed.FS) {
	params := &Params{
		Module:  mod,
		AppPkg:  "app",
		AppName: root,
	}
	if len(name) != 0 {
		params.AppPkg = "app/" + name
		params.AppName = name
	}
	_ = fs.WalkDir(fsys, "pkg/app", func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			return nil
		}
		if filepath.Ext(path) == ".go" {
			return nil
		}
		output := buildOutput(root, path, name)
		buildTmpl(fsys, path, output, params)
		return nil
	})
}

func buildOutput(root, path, appName string) string {
	var builder strings.Builder

	builder.WriteString(root)
	builder.WriteString("/")

	dir, name := filepath.Split(path)

	// dockerfile
	if name == "Dockerfile" {
		if len(appName) != 0 {
			builder.WriteString(appName)
			builder.WriteString(".dockerfile")
		} else {
			builder.WriteString("Dockerfile")
		}
		return builder.String()
	}

	// dir
	if len(dir) != 0 {
		builder.WriteString(dir)
	}

	// name
	switch ext := filepath.Ext(path); ext {
	case ".yiigo":
		builder.WriteString(name[:len(name)-6])
		builder.WriteString(".go")
	case "":
		if strings.Contains(name, "ignore") {
			builder.WriteString(".")
		}
		builder.WriteString(name)
	default:
		builder.WriteString(name)
	}

	output := builder.String()
	if len(appName) != 0 {
		output = strings.ReplaceAll(output, "app", "app/"+appName)
	}
	return output
}

func buildTmpl(fsys embed.FS, path, output string, params *Params) {
	b, _ := fsys.ReadFile(path)
	// 模板解析
	t, err := template.New(path).Parse(string(b))
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
