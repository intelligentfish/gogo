package asset

import (
	"encoding/json"
	"fmt"
	"github.com/jinzhu/gorm"
)

// UserAsset 用户资产
type UserAsset struct {
	gorm.Model `json:"-"`
	Nickname   string `json:"nickname"`
	Avatar     string `json:"avatar"`
	Balance    int64  `json:"balance"`
}

// TableName 表名
func (object *UserAsset) TableName() string {
	return "user_assets"
}

// GetHashKey Hash存储Key
func (object *UserAsset) GetHashKey() string {
	return fmt.Sprintf("user:%d", object.ID)
}

// GetAssetKey 资产Key
func (object *UserAsset) GetAssetKey() string {
	return "user_asset"
}

// GetAssetDirtyFlagKey 资产脏标志Key
func (object *UserAsset) GetAssetDirtyFlagKey() string {
	return "user_asset_dirty_flag"
}

// Marshal 序列化
func (object *UserAsset) Marshal() (s string, err error) {
	var raw []byte
	if raw, err = json.Marshal(object); nil != err {
		return
	}
	s = string(raw)
	return
}

// Unmarshal 反序列化
func (object *UserAsset) Unmarshal(s string) (err error) {
	err = json.Unmarshal([]byte(s), object)
	return
}

// New 新建对象
func (object *UserAsset) New() IAsset {
	return &UserAsset{}
}

// Clone 复制对象
func (object *UserAsset) Clone() IAsset {
	copy := &UserAsset{
		Nickname: object.Nickname,
		Avatar:   object.Avatar,
		Balance:  object.Balance,
	}
	copy.ID = object.ID
	return copy
}

// String 字符串描述
func (object *UserAsset) String() string {
	raw, _ := json.Marshal(object)
	return string(raw)
}
