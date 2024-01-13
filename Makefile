VERSION = v1.0.6
REMARK = "支付完成后推送消息给社区"
build:
	GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o center-order main.go
	#git tag -a release-$(VERSION) -m "$(REMARK)"
	#git push --tags
