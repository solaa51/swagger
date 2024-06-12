# 获取本地机器当前用户
USER = $(shell id -u -n)
# 编译时间
DATE = $(shell date '+%Y-%m-%d %H:%M:%S')

.PHONY: build
VERSION = v1.1.5
REMARK = "新增kv库"
build:
	git add .
	git commit -m $(REMARK)
	git tag -a $(VERSION) -m $(REMARK)
	git push origin $(VERSION)
