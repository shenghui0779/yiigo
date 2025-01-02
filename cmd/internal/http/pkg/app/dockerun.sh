#!/bin/bash
docker rm -f {{.AppName}}
docker rmi -f img_{{.AppName}}

{{- if eq .DockerF "Dockerfile"}}
docker build -t img_{{.AppName}} .
{{- else}}
docker build -f {{.DockerF}} -t img_{{.AppName}} .
{{- end}}
docker image prune -f

docker run -d --name={{.AppName}} --restart=always --privileged -p 10086:8000 -v /data/{{.AppName}}:/data img_{{.AppName}}
