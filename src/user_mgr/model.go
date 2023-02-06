package user_mgr

type UserModel struct {
	Id               int64             `json:"id" gorm:"primaryKey"`
	Email            string            `json:"email" gorm:"type:varchar(100); index; unique"`
	Password         string            `json:"password" gorm:"type:varchar(100)"`
	Token            string            `json:"token" gorm:"type:varchar(32); index"`
	Forbidden        bool              `json:"forbidden"`
	Roles            string            `json:"roles"`
	Roles_map        map[string]string `json:"roles_map" gorm:"-"`
	Permissions      string            `json:"permissions"`
	Permissions_map  map[string]string `json:"permissions_map" gorm:"-"`
	Register_ip      string            `json:"register_ip" gorm:"type:varchar(45)" `
	Created_unixtime int64             `json:"created_unixtime" gorm:"autoCreateTime"`
}
