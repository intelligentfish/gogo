package asset

import "github.com/go-redis/redis"

var (
	Util = &util{}
)

// 资产工具
type util struct {
}

// interfaceArrayToAssetArray 接口数组转换为资产数组
func (object *util) interfaceArrayToAssetArray(result interface{}, asset IAsset) (assets []IAsset, err error) {
	assets = make([]IAsset, len(result.([]interface{})))
	for idx, result := range result.([]interface{}) {
		target := asset.New()
		if err = target.Unmarshal(result.(string)); nil != err {
			return
		}
		assets[idx] = target
	}
	return
}

// UpdateAsset 更新资产
func (object *util) UpdateAsset(c *redis.Client, asset IAsset) (err error) {
	var s string
	if s, err = asset.Marshal(); nil != err {
		return
	}
	_, err = c.Eval(`
	redis.call('hset', KEYS[2], KEYS[3], 1)
	redis.call('hset', KEYS[1], KEYS[3], ARGV[1])
	return 1
`, []string{asset.GetAssetKey(), asset.GetAssetDirtyFlagKey(), asset.GetHashKey()},
		[]string{s}).Result()
	return
}

// GetAsset 获取资产
func (object *util) GetAsset(c *redis.Client, asset IAsset) (err error) {
	var s string
	if s, err = c.HGet(asset.GetAssetKey(), asset.GetHashKey()).Result(); nil != err {
		return
	}
	err = asset.Unmarshal(s)
	return
}

// GetAllAsset 获取所有资产
func (object *util) GetAllAsset(c *redis.Client, asset IAsset) (assets []IAsset, err error) {
	var result interface{}
	result, err = c.Eval(`
	local cursor = 0
	local assets = {}
	repeat
		local result = redis.call('hscan', KEYS[1], cursor, 'count', '1000')
		cursor = tonumber(result[1])
		local fields = result[2]
		if fields and 0 < #fields then
			local bulk = redis.call('hmget', KEYS[1], unpack(fields))
			for i, v in ipairs(bulk) do
				if v then
					assets[#assets+1] = v
				end
			end
		end
	until 0 == cursor
	return assets
	`, []string{asset.GetAssetKey()}, []string{}).Result()
	if nil != err {
		return
	}
	return object.interfaceArrayToAssetArray(result, asset)
}

// GetDirtyUserAsset 获取脏的用户资产
func (object *util) GetDirtyUserAsset(c *redis.Client, asset IAsset) (assets []IAsset, err error) {
	var result interface{}
	result, err = c.Eval(`
	local cursor = 0
	local assets = {}
	repeat
		local result = redis.call('hscan', KEYS[2], cursor, 'count', '1000')
		cursor = tonumber(result[1])
		local fields = result[2]
		if fields and 0 < #fields then
			local ids = {}
			for i = 1, #fields, 2 do
				if '1' == fields[i+1] then
					ids[#ids+1] = fields[i]
				end
			end
			if ids and 0 < #ids then
				local tbUnDirty = {}
				local bulk = redis.call('hmget', KEYS[1], unpack(ids))
				for i, v in ipairs(bulk) do
					if v then
						assets[#assets+1] = v
						tbUnDirty[#tbUnDirty+1] = ids[i]
						tbUnDirty[#tbUnDirty+1] = '0'
					end
				end
				redis.call('hmset', KEYS[2], unpack(tbUnDirty))
			end
		end
	until 0 == cursor
	return assets
	`, []string{asset.GetAssetKey(), asset.GetAssetDirtyFlagKey()},
		[]string{}).Result()
	if nil != err {
		return
	}
	return object.interfaceArrayToAssetArray(result, asset)
}
