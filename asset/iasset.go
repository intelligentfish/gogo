package asset

// 资产接口
type IAsset interface {
	// 获取Hash Key
	GetHashKey() string
	// 获取资产Key
	GetAssetKey() string
	// 获取资产脏标志Key
	GetAssetDirtyFlagKey() string
	// 序列化
	Marshal() (s string, err error)
	// 反序列化
	Unmarshal(s string) (err error)
	// 新建对象
	New() IAsset
	// 克隆对象
	Clone() IAsset
}
