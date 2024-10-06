package grpc

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
		"name":   "pkg_lib_log_traceid.tmpl",
		"path":   "pkg_lib_log_traceid.tmpl",
		"output": "pkg/lib/log/trace_id.go",
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
		"name":   "pkg_lib_middleware_traceid.tmpl",
		"path":   "pkg_lib_middleware_traceid.tmpl",
		"output": "pkg/lib/middleware/trace_id.go",
	},
	{
		"name":   "pkg_lib_middleware_validator.tmpl",
		"path":   "pkg_lib_middleware_validator.tmpl",
		"output": "pkg/lib/middleware/validator.go",
	},
	{
		"name":   "pkg_lib_result_code.tmpl",
		"path":   "pkg_lib_result_code.tmpl",
		"output": "pkg/lib/result/code.go",
	},
	{
		"name":   "pkg_lib_result_status.tmpl",
		"path":   "pkg_lib_result_status.tmpl",
		"output": "pkg/lib/result/status.go",
	},
	{
		"name":   "pkg_lib_util_validator.tmpl",
		"path":   "pkg_lib_util_validator.tmpl",
		"output": "pkg/lib/util/validator.go",
	},
	{
		"name":   "README.tmpl",
		"path":   "README.tmpl",
		"output": "README.md",
	},
}

var App = []map[string]string{
	{
		"name":   "buf.tmpl",
		"path":   "app/buf.tmpl",
		"output": "buf.yaml",
	},
	{
		"name":   "buf_gen.tmpl",
		"path":   "app/buf_gen.tmpl",
		"output": "buf.gen.yaml",
	},
	{
		"name":   "proto_buf_validate.tmpl",
		"path":   "app/proto_buf_validate.tmpl",
		"output": "api/buf/validate/validate.proto",
	},
	{
		"name":   "proto_google_annotations.tmpl",
		"path":   "app/proto_google_annotations.tmpl",
		"output": "api/google/api/annotations.proto",
	},
	{
		"name":   "proto_google_http.tmpl",
		"path":   "app/proto_google_http.tmpl",
		"output": "api/google/api/http.proto",
	},
	{
		"name":   "proto_greeter.tmpl",
		"path":   "app/proto_greeter.tmpl",
		"output": "api/greeter.proto",
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
		"name":   "pkg_app_server_grpc.tmpl",
		"path":   "app/pkg_app_server_grpc.tmpl",
		"output": "server/grpc.go",
	},
	{
		"name":   "pkg_app_server_http.tmpl",
		"path":   "app/pkg_app_server_http.tmpl",
		"output": "server/http.go",
	},
	{
		"name":   "pkg_app_service_greeter.tmpl",
		"path":   "app/pkg_app_service_greeter.tmpl",
		"output": "service/greeter.go",
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
		"name":   "README.tmpl",
		"path":   "app/README.tmpl",
		"output": "README.md",
	},
}
