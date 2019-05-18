package controllers

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"pyg/pyg/models"
	"strconv"
	"github.com/gomodule/redigo/redis"
)

type OrderController struct {
	beego.Controller
}

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
		temp["count"]=count
		totalPrice += littleCount
		totalCount += 1
		goods=append(goods,temp)
	}

	//返回数据
	this.Data["addrs"] = addrs
	this.Data["goods"]=goods
	this.Data["totalPrice"]=totalPrice
	this.Data["totalCount"]=totalCount
	this.Data["truePrice"]=totalPrice+10
	this.TplName="place_order.html"
}
