VERSION = v1.0.6
REMARK = "调试"
build:
	#GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o center-order main.go
	#git tag -a release-$(VERSION) -m "$(REMARK)"
	#git push --tags
	git commit -m "测试自动提交"
	git push
