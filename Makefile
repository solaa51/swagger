# 获取本地机器当前用户
USER = $(shell id -u -n)
# 编译时间
DATE = $(shell date '+%Y-%m-%d %H:%M:%S')

.PHONY: build
VERSION = v1.1.7
REMARK = "修复http请求参数校验bug"
build:
	git add .
	git commit -m $(REMARK)
	git tag -a $(VERSION) -m $(REMARK)
	git push origin $(VERSION)
