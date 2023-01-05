.DEFAULT_GOAL := build
bin=k8s-watch

build:
	@go build -o ${bin} -ldflags="-X 'main.buildInfo=${shell date '+%Y_%m_%d'}-${shell git branch --show-current}'"
