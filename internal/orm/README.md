数据库orm调用基础库

配置示例

```
[[dbConfig]]
uName = "irentOnlyRead" #必须唯一 方便索引数据库
mark = "正式库只读"
host = "rm-bp1t8o27rta4z4895go.mysql.rds.aliyuncs.com"
user = "read_only"
pass = "r!Gd$8vehqg7Z"
port = "3306"
name = "irent"
# ssh tunnel配置
tunnelSSHHost = "47.108.200.228"
tunnelSSHPort = "22"
tunnelSSHUser = "root"
tunnelSSHPassword = ""
# 秘钥验证 RSA PRIVATE KEY
tunnelSSHKey = "test.pem"
# 秘钥生成时的密码 一般都没有
tunnelSSHPassphrase = ""
# 自定义协议名称
tunnelSSHNetName = "mysql-ssh-tunnel"
```