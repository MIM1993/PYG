package controllers

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"pyg/pyg/models"
	"strconv"
	"github.com/gomodule/redigo/redis"
	"fmt"
	"strings"
	"time"
)

type OrderController struct {
	beego.Controller
}

//展示订单页面
func (this *OrderController) ShowOrder() {
	//获取数据
	goodsIds := this.GetStrings("checkGoods")
	//校验数据
	if len(goodsIds) == 0 {
		this.Redirect("/user/showCart", 302)
		return
	}
	//校验是否登陆
	name := this.GetSession("name")
	if name == nil {
		this.Redirect("/user/showCart", 302)
		return
	}
	//处理数据
	//获取用户所用收获地址
	o := orm.NewOrm()
	var addrs []models.Address
	o.QueryTable("Address").RelatedSel("User").Filter("User__Name", name.(string)).All(&addrs)

	//连接redis，获取商品数量
	conn, _ := redis.Dial("tcp", "127.0.0.1:6379")
	//获取商品信息，获取总价和总件数
	var goods []map[string]interface{}
	//总价和总件数
	var totalPrice, totalCount int
	for _, v := range goodsIds {
		//定义行容器
		temp := make(map[string]interface{})
		var goodsSku models.GoodsSKU
		id, _ := strconv.Atoi(v)
		goodsSku.Id = id
		o.Read(&goodsSku)
		//获取商品数量
		count, _ := redis.Int(conn.Do("hget", "cart_"+name.(string), id))
		littleCount := goodsSku.Price * count

		temp["goodsSku"] = goodsSku
		temp["littleCount"] = littleCount
		temp["count"] = count
		totalPrice += littleCount
		totalCount += 1
		goods = append(goods, temp)
	}

	//返回数据
	this.Data["addrs"] = addrs
	this.Data["goods"] = goods
	this.Data["totalPrice"] = totalPrice
	this.Data["totalCount"] = totalCount
	this.Data["truePrice"] = totalPrice + 10
	this.Data["goodsIds"] = goodsIds
	this.TplName = "place_order.html"
}

//提交订单操作
func (this *OrderController) HandlePushOrder() {
	//获取数据
	//地址ID
	addrId, err1 := this.GetInt("addrId")
	//支付ID
	payId, err2 := this.GetInt("payId")
	//商品ID
	goodsIds := this.GetString("goodsIds")
	//总数和总价格
	totalCount, err3 := this.GetInt("totalCount")
	totalPrice, err4 := this.GetInt("totalPrice")

	resp := make(map[string]interface{})
	defer respFunc(&this.Controller, resp)
	//校验数据
	if err1 != nil || err2 != nil || err3 != nil || err4 != nil || goodsIds == "" {
		fmt.Println("获取数据错误")
		resp["errnum"] = 1
		resp["errmsg"] = "获取数据错误"
		this.Redirect("/user/showCart", 302)
		return
	}
	//获取session
	name := this.GetSession("name")
	if name == nil {
		fmt.Println("用户未登陆")
		resp["errnum"] = 2
		resp["errmsg"] = "用户未登陆"
		this.Redirect("/login", 302)
		return
	}
	//处理数据 插入数据库
	//获取用户对象和地址对象
	o := orm.NewOrm()
	var user models.User //用户对象
	user.Name = name.(string)
	o.Read(&user, "Name")

	var address models.Address //地址对象
	address.Id = addrId
	o.Read(&address)

	var orderInfo models.OrderInfo //订单对象
	orderInfo.User = &user
	orderInfo.Address = &address
	orderInfo.PayMethod = payId
	orderInfo.TotalCount = totalCount
	orderInfo.TotalPrice = totalPrice
	orderInfo.TransitPrice = 10
	orderInfo.OrderId = time.Now().Format("20060102150405" + strconv.Itoa(user.Id))
	o.Insert(&orderInfo) //插入订单表
	//将字符串转换为切片
	goodsSlice := strings.Split(goodsIds[1:len(goodsIds)-1], " ")  //注意断开方式是空格
	//循环切片
	//连接redis
	conn, err := redis.Dial("tcp", "127.0.0.1:6379")
	defer conn.Close()
	if err != nil {
		resp["errnum"] = 3
		resp["errmsg"] = "redis连接错误"
		return
	}
	for _, v := range goodsSlice {
		id, _ := strconv.Atoi(v) //转为int类型
		var goodsSku models.GoodsSKU
		goodsSku.Id = id
		o.Read(&goodsSku)

		//获取商品数量和价格
		count, _ := redis.Int(conn.Do("hget", "cart_"+name.(string), id))
		littleprice := goodsSku.Price * count

		//插入订单商品表
		var orderGoods models.OrderGoods
		orderGoods.GoodsSKU = &goodsSku
		orderGoods.OrderInfo = &orderInfo
		orderGoods.Count = count
		orderGoods.Price = littleprice
		//判断库存是否大于订单数量，如果大于更新库存和销量
		if goodsSku.Stock < count {
			resp["errnum"] = 4
			resp["errmsg"] = "库存数量不足"
			return
		}
		goodsSku.Stock -= count
		goodsSku.Sales += count
		//更新商品表
		o.Update(&goodsSku)
		//将数据插入订单商品表
		_,err :=o.Insert(&orderGoods)
		if err!=nil{
			resp["errnum"] = 6
			resp["errmsg"] = "订单商品表插入失败"
			return
		}
		//订单完成，删除redis数据库中的信息
		conn.Do("hdel","cart_"+name.(string),id)
	}

	//返回数据
	resp["errnum"]=5
	resp["errmsg"]="OK"
}
