# Internal项目结构说明
```
├─cmd: 项目启动入口，这里存放main.go文件，项目外部存放dockerfile或shell启动脚本，方便部署
├─client:存放对外暴露的Client相关方法接口与对象
├─event:存放event相关的方法接口与对象
├─internal:不希望外部了解的内部相关实现，TODO：内部结构待二次优化，暂时先按这样丢进去
│  ├─cache
│  ├─entity
│  ├─info
│  └─pkg:内部使用的工具方法
├─message
├─pkg：对外暴露的工具方法
│  ├─highway
│  ├─oidb
│  ├─pb
│  │  ├─action
│  │  ├─login
│  │  ├─message
│  │  ├─service
│  │  │  ├─highway
│  │  │  └─oidb
│  │  └─system
│  ├─tlv
│  └─wtlogin
│      ├─loginState
│      └─qrcodeState
└─utils：暂时不知道干啥的，先不动
    ├─binary
    ├─crypto
    │  └─ecdh
    ├─platform
    └─proto
```
## 一点开发私货：
可以合理使用interface，对外全部以interface FactoryMethod进行暴露，而不是直接暴露struct对象，对象使用变量通过Get/Set拿取


附带一个FactoryMethod使用样例：

工厂：可以对于所有的Client添加一个ClientFactory的FactoryMethod，管理所有的Client
```golang
type (
	Factory interface { // 工厂接口
		Demo() demo.Store
	}
	DataStore struct { // 工厂实体所用到的一些配置实例
		DB *gorm.DB
	}
)

var _ Factory = (*DataStore)(nil)

func (ds *DataStore) Demo() demo.Store { // 加载对应的实体，在本项目中可抽象为QQClient
	return demo.NewDemo(ds)
}

func (ds DataStore) GetMySQL() *gorm.DB { // 实体通过工厂反射获取自己所需要的配置
	return ds.DB
}
```
实体：
```golang
type (
	Store interface {
		Hello()
	}
	demoMethod interface {
		GetMySQL() *gorm.DB
	}
	demoStore struct {
		db *gorm.DB
	}
)

func NewDemo(ds demoMethod) *demoStore {
	return &demoStore{
		db: ds.GetMySQL(),
	}
}

func (ds *demoStore) Hello() {
	print("hello world")
}
```
使用：
```golang
func main {
    sro := &store.DataStore{
        DB: nil, // DB初始化，加载Factory所需相关的配置
    }
    sro.Demo().Hello()
}
```
## 好处在哪里？
1. 项目结构更加清晰。
2. 节约框架学习成本。接入方只需要关心Factory能够给他们提供什么实例即可，无需额外去找到底有几个Client需要调用
3. 便于拓展、维护。单个Client开发者无需关心整体的工具实现，只需根据Factory提供的工具来开发维护自己负责的Client，开发完成后直接去Factory里注册即可，而无需关注整体代码情况