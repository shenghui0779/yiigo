package main

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"golang.org/x/mod/modfile"

	"github.com/shenghui0779/yiigo"
	"github.com/shenghui0779/yiigo/cmd/internal"
)

func main() {
	cmd := &cobra.Command{
		Use:   "yiigo",
		Short: "项目脚手架",
		Long:  "项目脚手架，用于快速创建Go项目",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if cmd.Use == "new" && len(args) != 0 {
				if err := os.MkdirAll(args[0], 0o775); err != nil {
					log.Fatalln(err)
				}
			}
		},
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("欢迎使用yiigo脚手架")
		},
	}
	// 注册命令
	cmd.AddCommand(project(), app())
	// 执行
	if err := cmd.Execute(); err != nil {
		log.Fatalln("Error cmd execute", zap.Error(err))
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
			"yiigo new demo",
			"yiigo new demo --mod=xxx.yyy.com",
			"yiigo new demo --apps=foo,bar",
			"yiigo new demo --apps=foo --apps=bar",
			"yiigo new demo --mod=xxx.yyy.com --apps=foo --apps=bar",
		),
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return errors.New("必须指定一个项目名称")
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			if len(mod) == 0 {
				mod = args[0]
			}
			if grpc {
				fmt.Println("暂不支持gRPC项目")
				return
			}
			internal.InitHttpProject(args[0], mod, apps...)
			fmt.Println("项目创建完成！不要忘记执行<ent/generate.go>")
		},
	}
	// 注册参数
	cmd.Flags().BoolVar(&grpc, "grpc", false, "创建gRPC项目（暂不支持）")
	cmd.Flags().StringVar(&mod, "mod", "", "设置Module名称（默认为项目名称）")
	cmd.Flags().StringSliceVar(&apps, "apps", []string{}, "创建多应用项目")
	return cmd
}

func app() *cobra.Command {
	var grpc bool
	cmd := &cobra.Command{
		Use:   "app",
		Short: "新增应用",
		Example: yiigo.CmdExamples(
			"yiigo app foo",
		),
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return errors.New("必须指定一个App名称")
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			// 读取 go.mod 文件
			data, err := os.ReadFile("go.mod")
			if err != nil {
				log.Fatalln("读取go.mod文件失败「", err, "」请在Go项目根目录下执行命令")
			}
			// 解析 go.mod 文件
			f, err := modfile.Parse("go.mod", data, nil)
			if err != nil {
				log.Fatalln("解析go.mod文件失败:", err)
			}
			if grpc {
				fmt.Println("暂不支持gRPC应用")
				return
			}
			internal.InitHttpApp(".", f.Module.Mod.Path, args[0])
			fmt.Println("应用创建完成！不要忘记执行<ent/generate.go>")
		},
	}
	// 注册参数
	cmd.Flags().BoolVar(&grpc, "grpc", false, "新增gRPC应用（暂不支持）")
	return cmd
}
