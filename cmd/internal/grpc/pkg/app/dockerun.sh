#!/bin/bash
docker rm -f app_{{.AppName}}
docker rmi -f img_{{.AppName}}

{{- if eq .DockerF "Dockerfile"}}
docker build -t img_{{.AppName}} .
{{- else}}
docker build -f {{.DockerF}} -t img_{{.AppName}} .
{{- end}}
docker image prune -f

docker run -d --name=app_{{.AppName}} --restart=always --privileged -p 10085:50051 -p 10086:8000 -v /data/app_{{.AppName}}:/data img_{{.AppName}}
