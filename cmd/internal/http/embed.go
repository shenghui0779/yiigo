package http

import "embed"

//go:embed all:*
var FS embed.FS

var Project = []map[string]string{
	{
		"name":   "config_toml.tmpl",
		"path":   "config_toml.tmpl",
		"output": "config.toml.example",
	},
	{
		"name":   "dockerignore.tmpl",
		"path":   "dockerignore.tmpl",
		"output": ".dockerignore",
	},
	{
		"name":   "gitignore.tmpl",
		"path":   "gitignore.tmpl",
		"output": ".gitignore",
	},
	{
		"name":   "gomod.tmpl",
		"path":   "gomod.tmpl",
		"output": "go.mod",
	},
	{
		"name":   "gosum.tmpl",
		"path":   "gosum.tmpl",
		"output": "go.sum",
	},
	{
		"name":   "pkg_lib_binding.tmpl",
		"path":   "pkg_lib_binding.tmpl",
		"output": "pkg/lib/binding.go",
	},
	{
		"name":   "pkg_lib_embed.tmpl",
		"path":   "pkg_lib_embed.tmpl",
		"output": "pkg/lib/embed.go",
	},
	{
		"name":   "pkg_lib_util.tmpl",
		"path":   "pkg_lib_util.tmpl",
		"output": "pkg/lib/util.go",
	},
	{
		"name":   "pkg_lib_db_db.tmpl",
		"path":   "pkg_lib_db_db.tmpl",
		"output": "pkg/lib/db/db.go",
	},
	{
		"name":   "pkg_lib_db_mixin.tmpl",
		"path":   "pkg_lib_db_mixin.tmpl",
		"output": "pkg/lib/db/mixin.go",
	},
	{
		"name":   "pkg_lib_db_redis.tmpl",
		"path":   "pkg_lib_db_redis.tmpl",
		"output": "pkg/lib/db/redis.go",
	},
	{
		"name":   "pkg_lib_identity_identity.tmpl",
		"path":   "pkg_lib_identity_identity.tmpl",
		"output": "pkg/lib/identity/identity.go",
	},
	{
		"name":   "pkg_lib_log_init.tmpl",
		"path":   "pkg_lib_log_init.tmpl",
		"output": "pkg/lib/log/init.go",
	},
	{
		"name":   "pkg_lib_log_log.tmpl",
		"path":   "pkg_lib_log_log.tmpl",
		"output": "pkg/lib/log/log.go",
	},
	{
		"name":   "pkg_lib_middleware_cors.tmpl",
		"path":   "pkg_lib_middleware_cors.tmpl",
		"output": "pkg/lib/middleware/cors.go",
	},
	{
		"name":   "pkg_lib_middleware_log.tmpl",
		"path":   "pkg_lib_middleware_log.tmpl",
		"output": "pkg/lib/middleware/log.go",
	},
	{
		"name":   "pkg_lib_middleware_monitor.tmpl",
		"path":   "pkg_lib_middleware_monitor.tmpl",
		"output": "pkg/lib/middleware/monitor.go",
	},
	{
		"name":   "pkg_lib_middleware_recovery.tmpl",
		"path":   "pkg_lib_middleware_recovery.tmpl",
		"output": "pkg/lib/middleware/recovery.go",
	},
	{
		"name":   "pkg_lib_result_code.tmpl",
		"path":   "pkg_lib_result_code.tmpl",
		"output": "pkg/lib/result/code.go",
	},
	{
		"name":   "pkg_lib_result_result.tmpl",
		"path":   "pkg_lib_result_result.tmpl",
		"output": "pkg/lib/result/result.go",
	},
	{
		"name":   "README.tmpl",
		"path":   "README.tmpl",
		"output": "README.md",
	},
}

var App = []map[string]string{
	{
		"name":   "pkg_app_api_controller_demo.tmpl",
		"path":   "app/pkg_app_api_controller_demo.tmpl",
		"output": "api/controller/demo.go",
	},
	{
		"name":   "pkg_app_api_middleware_auth.tmpl",
		"path":   "app/pkg_app_api_middleware_auth.tmpl",
		"output": "api/middleware/auth.go",
	},
	{
		"name":   "pkg_app_api_router_app.tmpl",
		"path":   "app/pkg_app_api_router_app.tmpl",
		"output": "api/router/app.go",
	},
	{
		"name":   "pkg_app_api_service_demo_create.tmpl",
		"path":   "app/pkg_app_api_service_demo_create.tmpl",
		"output": "api/service/demo/create.go",
	},
	{
		"name":   "pkg_app_api_service_demo_test.tmpl",
		"path":   "app/pkg_app_api_service_demo_test.tmpl",
		"output": "api/service/demo/demo_test.go",
	},
	{
		"name":   "pkg_app_cmd_hello.tmpl",
		"path":   "app/pkg_app_cmd_hello.tmpl",
		"output": "cmd/hello.go",
	},
	{
		"name":   "pkg_app_cmd_init.tmpl",
		"path":   "app/pkg_app_cmd_init.tmpl",
		"output": "cmd/init.go",
	},
	{
		"name":   "pkg_app_cmd_migrate.tmpl",
		"path":   "app/pkg_app_cmd_migrate.tmpl",
		"output": "cmd/migrate.go",
	},
	{
		"name":   "pkg_app_cmd_root.tmpl",
		"path":   "app/pkg_app_cmd_root.tmpl",
		"output": "cmd/root.go",
	},
	{
		"name":   "pkg_app_ent_db.tmpl",
		"path":   "app/pkg_app_ent_db.tmpl",
		"output": "ent/db.go",
	},
	{
		"name":   "pkg_app_ent_generate.tmpl",
		"path":   "app/pkg_app_ent_generate.tmpl",
		"output": "ent/generate.go",
	},
	{
		"name":   "pkg_app_ent_gitignore.tmpl",
		"path":   "app/pkg_app_ent_gitignore.tmpl",
		"output": "ent/.gitignore",
	},
	{
		"name":   "pkg_app_ent_schema_demo.tmpl",
		"path":   "app/pkg_app_ent_schema_demo.tmpl",
		"output": "ent/schema/demo.go",
	},
	{
		"name":   "pkg_app_main.tmpl",
		"path":   "app/pkg_app_main.tmpl",
		"output": "main.go",
	},
	{
		"name":   "dockerfile.tmpl",
		"path":   "app/dockerfile.tmpl",
		"output": "Dockerfile",
	},
	{
		"name":   "config_toml.tmpl",
		"path":   "config_toml.tmpl",
		"output": "config.toml",
	},
	{
		"name":   "pkg_app_web_dist_index.tmpl",
		"path":   "app/pkg_app_web_dist_index.tmpl",
		"output": "web/dist/index.html",
	},
	{
		"name":   "pkg_app_web_dist_style.tmpl",
		"path":   "app/pkg_app_web_dist_style.tmpl",
		"output": "web/dist/style.css",
	},
	{
		"name":   "pkg_app_web_embed.tmpl",
		"path":   "app/pkg_app_web_embed.tmpl",
		"output": "web/embed.go",
	},
}
