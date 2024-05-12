package cache

import (
	"reflect"
	"unsafe"

	"github.com/LagrangeDev/LagrangeGo/client/entity"
	"github.com/RomiChan/syncx"
)

type cacheType uint32

const (
	cacheTypeCache cacheType = 1 << iota
	cacheTypeFriend
	cacheTypeGroupInfo
	cacheTypeGroupMember
)

func typenameof[T any]() string {
	return reflect.ValueOf((*T)(nil)).Type().String()
}

var cacheTypesMap = map[string]cacheType{
	typenameof[Cache]():              cacheTypeCache,
	typenameof[entity.Friend]():      cacheTypeFriend,
	typenameof[entity.Group]():       cacheTypeGroupInfo,
	typenameof[entity.GroupMember](): cacheTypeGroupMember,
}

type Cache struct {
	m         syncx.Map[uint64, unsafe.Pointer]
	refreshed syncx.Map[cacheType, struct{}]
}

func hasRefreshed[T any](c *Cache) bool {
	typ := cacheTypesMap[reflect.ValueOf((*T)(nil)).Type().String()]
	if typ == 0 {
		return false
	}
	_, ok := c.refreshed.Load(typ)
	return ok
}

func refreshAllCacheOf[T any](c *Cache, newcache map[uint32]*T) {
	typstr := reflect.ValueOf((*T)(nil)).Type().String()
	typ := cacheTypesMap[typstr]
	if typ == 0 {
		return
	}
	c.refreshed.Store(typ, struct{}{})
	key := uint64(typ) << 32
	dellst := make([]uint64, 0, 64)
	c.m.Range(func(k uint64, v unsafe.Pointer) bool {
		if k&key != 0 {
			if _, ok := newcache[uint32(k)]; !ok {
				dellst = append(dellst, k)
			}
		}
		return true
	})
	for k, v := range newcache {
		c.m.Store(key|uint64(k), unsafe.Pointer(v))
	}
	for _, k := range dellst {
		c.m.Delete(k)
	}
}

func setCacheOf[T any](c *Cache, k uint32, v *T) {
	typstr := reflect.ValueOf(v).Type().String()
	typ := cacheTypesMap[typstr]
	if typ == 0 {
		return
	}
	key := uint64(typ)<<32 | uint64(k)
	c.m.Store(key, unsafe.Pointer(v))
}

func getCacheOf[T any](c *Cache, k uint32) (v *T, ok bool) {
	typstr := reflect.ValueOf(v).Type().String()
	typ := cacheTypesMap[typstr]
	if typ == 0 {
		return
	}
	key := uint64(typ)<<32 | uint64(k)
	unsafev, ok := c.m.Load(key)
	if ok {
		v = (*T)(unsafev)
	}
	return
}

func rangeCacheOf[T any](c *Cache, iter func(k uint32, v *T) bool) {
	typ := cacheTypesMap[reflect.ValueOf((*T)(nil)).Type().String()]
	if typ == 0 {
		return
	}
	key := uint64(typ) << 32
	c.m.Range(func(k uint64, v unsafe.Pointer) bool {
		if k&key != 0 {
			return iter(uint32(k), (*T)(v))
		}
		return true
	})
}
