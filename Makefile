# 获取本地机器当前用户
USER = $(shell id -u -n)
# 编译时间
DATE = $(shell date '+%Y-%m-%d %H:%M:%S')

.PHONY: build
VERSION = v1.1.6
REMARK = "singleFlight独立成新的库，方便分组调用"
build:
	git add .
	git commit -m $(REMARK)
	git tag -a $(VERSION) -m $(REMARK)
	git push origin $(VERSION)
