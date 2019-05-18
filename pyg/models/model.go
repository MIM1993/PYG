package models

import (
	"github.com/astaxie/beego/orm"
	_ "github.com/go-sql-driver/mysql"
	"time"
)

//用户表
type User struct {
	Id        int
	Name      string     `orm:"size(40);unique"`
	Pwd       string     `orm:"size(40)"`
	Phone     string     `orm:"size(11)"`
	Email     string     `orm:"null"`
	Active    bool       `orm:"default(false)"`
	Addresses []*Address `orm:"reverse(many)"`
	OrderInfo   []*OrderInfo `orm:"reverse(many)"`
}

//地址表
type Address struct {
	Id        int
	Receiver  string `orm:"size(40)"`
	Addr      string `orm:"size(100)"`
	PostCode  string
	Phone     string `orm:"size(11)"`
	User      *User  `orm:"rel(fk)"`
	IsDefault bool   `orm:"default(false)"` //默认地址
	OrderInfo   []*OrderInfo `orm:"reverse(many)"`
}

//首页三级联动类型表
type TpshopCategory struct {
	Id         int
	CateName   string `orm:"default('')"`
	Pid        int    `orm:"default(0)"`
	IsShow     int    `orm:"default(1)"`
	CreateTime int    `orm:"null"`
	UpdateTime int    `orm:"null"`
	DeleteTime int    `orm:"null"`
}

//商品SPU表
type Goods struct {
	Id       int
	Name     string      `orm:"size(20)"`  //商品名称
	Detail   string      `orm:"size(200)"` //详细描述
	GoodsSKU []*GoodsSKU `orm:"reverse(many)"`
}

//商品类型表
type GoodsType struct {
	Id                   int
	Name                 string //种类名称
	Logo                 string //logo
	Image                string //图片
	GoodsSKU             []*GoodsSKU             `orm:"reverse(many)"`
	IndexTypeGoodsBanner []*IndexTypeGoodsBanner `orm:"reverse(many)"`
}

//商品SKU表
type GoodsSKU struct {
	Id                   int
	Goods                *Goods                  `orm:"rel(fk)"`      //商品SPU
	GoodsType            *GoodsType              `orm:"rel(fk)"`      //商品所属种类
	Name                 string                                       //商品名称
	Desc                 string                                       //商品简介
	Price                int                                          //商品价格
	Unite                string                                       //商品单位
	Image                string                                       //商品图片
	Stock                int                     `orm:"default(1)"`   //商品库存
	Sales                int                     `orm:"default(0)"`   //商品销量
	Status               int                     `orm:"default(1)"`   //商品状态
	Time                 time.Time               `orm:"auto_now_add"` //添加时间
	GoodsImage           []*GoodsImage           `orm:"reverse(many)"`
	IndexGoodsBanner     []*IndexGoodsBanner     `orm:"reverse(many)"`
	IndexTypeGoodsBanner []*IndexTypeGoodsBanner `orm:"reverse(many)"`
	OrderGoods   []*OrderGoods `orm:"reverse(many)"`
}

//商品图片表
type GoodsImage struct {
	Id       int
	Image    string                    //商品图片
	GoodsSKU *GoodsSKU `orm:"rel(fk)"` //商品SKU
}

//首页轮播商品展示表
type IndexGoodsBanner struct {
	Id       int
	GoodsSKU *GoodsSKU `orm:"rel(fk)"`    //商品sku
	Image    string                       //商品图片
	Index    int       `orm:"default(0)"` //展示顺序
}

//首页分类商品展示表
type IndexTypeGoodsBanner struct {
	Id          int
	GoodsType   *GoodsType `orm:"rel(fk)"`    //商品类型
	GoodsSKU    *GoodsSKU  `orm:"rel(fk)"`    //商品sku
	DisplayType int        `orm:"default(1)"` //展示类型 0代表文字，1代表图片
	Index       int        `orm:"default(0)"` //展示顺序
}

//首页促销商品展示表
type IndexPromotionBanner struct {
	Id    int
	Name  string `orm:"size(20)"`   //活动名称
	Url   string `orm:"size(50)"`   //活动链接
	Image string                    //活动图片
	Index int    `orm:"default(0)"` //展示顺序
}

//订单表
type OrderInfo struct {
	Id 				int
	OrderId         string  `orm:"unique"`
	User 			*User	`orm:"rel(fk)"`		//用户
	Address 		*Address`orm:"rel(fk)"`		//地址
	PayMethod 		int							//付款方式
	TotalCount 	int		`orm:"default(1)"`	//商品数量
	TotalPrice 	int							//商品总价
	TransitPrice 	int							//运费
	Orderstatus 	int 	`orm:"default(1)"`	//订单状态
	TradeNo 		string	`orm:"default('')"`	//支付编号
	Time			time.Time `orm:"auto_now_add"`		//订单时间

	OrderGoods   []*OrderGoods `orm:"reverse(many)"`
}

//订单商品表
type OrderGoods struct {//订单商品表
	Id 			int
	OrderInfo 	*OrderInfo	`orm:"rel(fk)"`	//订单
	GoodsSKU 	*GoodsSKU	`orm:"rel(fk)"`	//商品
	Count 		int		`orm:"default(1)"`	//商品数量
	Price 		int							//商品价格
	Comment 	string	`orm:"default('')"` //评论
}

func init() {
	//注册数据库
	orm.RegisterDataBase("default", "mysql", "root:123456@tcp(127.0.0.1:3306)/pyg?charset=utf8")
	//注册表结构
	orm.RegisterModel(new(User), new(Address), new(TpshopCategory),new(IndexPromotionBanner),new(IndexTypeGoodsBanner),new(IndexGoodsBanner),new(GoodsImage),new(GoodsSKU),new(GoodsType),new(Goods),new(OrderInfo),new(OrderGoods))
	//生成表
	orm.RunSyncdb("default", false, true)
}
