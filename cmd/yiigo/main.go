package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/shenghui0779/yiigo"
	"github.com/shenghui0779/yiigo/cmd/internal"
	"github.com/spf13/cobra"
	"golang.org/x/mod/modfile"
)

func main() {
	cmd := &cobra.Command{
		Use:   "yiigo",
		Short: "项目脚手架",
		Long:  "项目脚手架，用于快速创建Go项目",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if cmd.Use == "new" && len(args) != 0 {
				if err := os.MkdirAll(args[0], 0o775); err != nil {
					log.Fatalln("os.MkdirAll:", internal.FmtErr(err))
				}
			}
		},
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("欢迎使用yiigo脚手架")
		},
	}
	// 注册命令
	cmd.AddCommand(project(), app(), ent())
	// 执行
	if err := cmd.Execute(); err != nil {
		log.Fatalln("cmd.Execute:", internal.FmtErr(err))
	}
}

func project() *cobra.Command {
	var grpc bool
	var mod string
	var apps []string
	cmd := &cobra.Command{
		Use:   "new",
		Short: "创建项目",
		Example: yiigo.CmdExamples(
			"-- HTTP --",
			"yiigo new demo",
			"yiigo new demo --mod=xxx.yyy.com",
			"yiigo new demo --apps=foo,bar",
			"yiigo new demo --apps=foo --apps=bar",
			"yiigo new demo --mod=xxx.yyy.com --apps=foo --apps=bar",
			"-- gRPC --",
			"yiigo new demo --grpc",
			"yiigo new demo --mod=xxx.yyy.com --grpc",
			"yiigo new demo --apps=foo,bar --grpc",
			"yiigo new demo --apps=foo --apps=bar --grpc",
			"yiigo new demo --mod=xxx.yyy.com --apps=foo --apps=bar --grpc",
		),
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return errors.New("必须指定一个项目名称")
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			workDir := args[0]
			if len(mod) == 0 {
				mod = workDir
			}
			// 判断目录是否为空
			if path, ok := internal.IsDirEmpty(workDir); !ok {
				fmt.Printf("👿 目录(%s)不为空，请确认！\n", path)
				return
			}
			// 创建项目文件
			fmt.Println("🍺 创建项目文件")
			if grpc {
				internal.InitGrpcProject(workDir, mod, apps...)
			} else {
				internal.InitHttpProject(workDir, mod, apps...)
			}
			// go mod init
			fmt.Println("🍺 执行 go mod init")
			modInit := exec.Command("go", "mod", "init", mod)
			modInit.Dir = workDir
			if err := modInit.Run(); err != nil {
				log.Fatalln("🐛 go mod init 执行失败:", internal.FmtErr(err))
			}
			// go mod tidy
			fmt.Println("🍺 执行 go mod tidy")
			modTidy := exec.Command("go", "mod", "tidy")
			modTidy.Dir = workDir
			modTidy.Stderr = os.Stderr
			if err := modTidy.Run(); err != nil {
				log.Fatalln("🐛 go mod tidy 执行失败:", internal.FmtErr(err))
			}
			fmt.Println("🍺 项目创建完成！请阅读README")
		},
	}
	// 注册参数
	cmd.Flags().BoolVar(&grpc, "grpc", false, "创建gRPC项目")
	cmd.Flags().StringVar(&mod, "mod", "", "设置Module名称（默认为项目名称）")
	cmd.Flags().StringSliceVar(&apps, "apps", []string{}, "创建多应用项目")
	return cmd
}

func app() *cobra.Command {
	var grpc bool
	cmd := &cobra.Command{
		Use:   "app",
		Short: "创建应用",
		Example: yiigo.CmdExamples(
			"-- HTTP --",
			"yiigo app hello",
			"-- gRPC --",
			"yiigo app hello --grpc",
		),
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return errors.New("必须指定一个App名称")
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("🍺 解析 go.mod")
			// 读取 go.mod 文件
			data, err := os.ReadFile("go.mod")
			if err != nil {
				log.Fatalln("🐛 读取go.mod文件失败:", internal.FmtErr(err))
			}
			// 解析 go.mod 文件
			f, err := modfile.Parse("go.mod", data, nil)
			if err != nil {
				log.Fatalln("🐛 解析go.mod文件失败:", internal.FmtErr(err))
			}
			// 创建应用文件
			fmt.Println("🍺 创建应用文件")
			if grpc {
				for _, name := range args {
					if path, ok := internal.IsDirEmpty("pkg/app/" + name); !ok {
						fmt.Printf("👿 目录(%s)不为空，请确认！\n", path)
						return
					}
					internal.InitGrpcApp(".", f.Module.Mod.Path, name)
				}
			} else {
				for _, name := range args {
					if path, ok := internal.IsDirEmpty("pkg/app/" + name); !ok {
						fmt.Printf("👿 目录(%s)不为空，请确认！\n", path)
						return
					}
					internal.InitHttpApp(".", f.Module.Mod.Path, name)
				}
			}
			// go mod tidy
			fmt.Println("🍺 执行 go mod tidy")
			modTidy := exec.Command("go", "mod", "tidy")
			modTidy.Stderr = os.Stderr
			if err := modTidy.Run(); err != nil {
				log.Fatalln("🐛 go mod tidy 执行失败:", internal.FmtErr(err))
			}
			fmt.Println("🍺 应用创建完成！请阅读README")
		},
	}
	// 注册参数
	cmd.Flags().BoolVar(&grpc, "grpc", false, "创建gRPC应用")
	return cmd
}

func ent() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ent",
		Short: "创建Ent实例",
		Example: yiigo.CmdExamples(
			"-- 默认实例 --",
			"yiigo ent",
			"-- 指定名称 --",
			"yiigo ent foo",
			"yiigo ent foo bar",
		),
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("🍺 解析 go.mod")
			// 读取 go.mod 文件
			data, err := os.ReadFile("go.mod")
			if err != nil {
				log.Fatalln("🐛 读取go.mod文件失败:", internal.FmtErr(err))
			}
			// 解析 go.mod 文件
			f, err := modfile.Parse("go.mod", data, nil)
			if err != nil {
				log.Fatalln("🐛 解析go.mod文件失败:", internal.FmtErr(err))
			}
			// 创建Ent文件
			fmt.Println("🍺 创建Ent文件")
			if len(args) != 0 {
				for _, name := range args {
					if path, ok := internal.IsDirEmpty("pkg/ent/" + name); !ok {
						fmt.Printf("👿 目录(%s)不为空，请确认！\n", path)
						return
					}
					internal.InitEnt(".", f.Module.Mod.Path, name)
				}
			} else {
				if path, ok := internal.IsDirEmpty("pkg/ent"); !ok {
					fmt.Printf("👿 目录(%s)不为空，请确认！\n", path)
					return
				}
				internal.InitEnt(".", f.Module.Mod.Path, "")
			}
			// go mod tidy
			fmt.Println("🍺 执行 go mod tidy")
			modTidy := exec.Command("go", "mod", "tidy")
			modTidy.Stderr = os.Stderr
			if err := modTidy.Run(); err != nil {
				log.Fatalln("🐛 go mod tidy 执行失败:", internal.FmtErr(err))
			}
			// ent generate
			fmt.Println("🍺 执行 ent generate")
			if len(args) != 0 {
				for _, name := range args {
					entGen := exec.Command("go", "generate", "./pkg/ent/"+name)
					if err := entGen.Run(); err != nil {
						log.Fatalln("🐛 ent generate 执行失败:", internal.FmtErr(err))
					}
				}
			} else {
				entGen := exec.Command("go", "generate", "./pkg/ent")
				if err := entGen.Run(); err != nil {
					log.Fatalln("🐛 ent generate 执行失败:", internal.FmtErr(err))
				}
			}
			// go mod tidy
			fmt.Println("🍺 执行 go mod tidy")
			modClean := exec.Command("go", "mod", "tidy")
			modClean.Stderr = os.Stderr
			if err := modClean.Run(); err != nil {
				log.Fatalln("🐛 go mod tidy 执行失败:", internal.FmtErr(err))
			}
			fmt.Println("🍺 Ent实例创建完成！请阅读README")
		},
	}
	return cmd
}
