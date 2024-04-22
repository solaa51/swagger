package model

// SysAdmin 管理员用户
type SysAdmin struct {
	Id         int64  `json:"id" gorm:"column:id;size:10;autoIncrement"`                               //int
	Username   string `json:"username" gorm:"column:username;size:32"`                                 //varchar(32) 用户名
	Password   string `json:"password" gorm:"column:password;size:32"`                                 //varchar(32)
	CreateTime string `json:"create_time" gorm:"column:create_time;size:20;default:CURRENT_TIMESTAMP"` //datetime 创建时间
	Roles      string `json:"roles" gorm:"column:roles;size:64"`                                       //varchar(64) 关联角色
	UpdateTime string `json:"update_time" gorm:"column:update_time;size:20;default:CURRENT_TIMESTAMP"` //datetime 更新时间
	LastTime   string `json:"last_time" gorm:"column:last_time;size:20;default:CURRENT_TIMESTAMP"`     //datetime 最后登录时间
	LastIp     string `json:"last_ip" gorm:"column:last_ip;size:64"`                                   //varchar(64) 最后登录IP
	IsDel      int64  `json:"is_del" gorm:"column:is_del;size:3;default:0"`                            //tinyint 是否删除：0正常 1删除
	Status     int64  `json:"status" gorm:"column:status;size:3;default:0"`                            //tinyint 状态： 0正常 1禁用
	RealName   string `json:"real_name" gorm:"column:real_name;size:20"`                               //varchar(20) 真实姓名
	Telephone  string `json:"telephone" gorm:"column:telephone;size:20"`                               //varchar(20) 手机号码
}
