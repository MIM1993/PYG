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
	"github.com/smartwalle/alipay"
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

	o.Begin()            //开启事务
	o.Insert(&orderInfo) //插入订单表
	//将字符串转换为切片
	goodsSlice := strings.Split(goodsIds[1:len(goodsIds)-1], " ") //注意断开方式是空格
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

		//初始库存
		oldStock := goodsSku.Stock
		fmt.Println("初始库存", oldStock)

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
			o.Rollback()
			return
		}

		time.Sleep(time.Second * 5) //手动加延迟

		//更新商品表
		o.Read(&goodsSku) //再次读取数据，确定当前数据与初始数据一样，在进行操作
		fmt.Println("当前库存", goodsSku.Stock)
		//高级更新
		qs := o.QueryTable("GoodsSKU").Filter("Id", id).Filter("Stock", oldStock)
		_, err := qs.Update(orm.Params{"Stock": goodsSku.Stock - count, "Sales": goodsSku.Sales + count})
		if err != nil {
			resp["errnum"] = 8
			resp["errmsg"] = "购买失败，请重新排队！"
			o.Rollback()
			return
		}

		//将数据插入订单商品表
		_, err = o.Insert(&orderGoods)
		if err != nil {
			resp["errnum"] = 6
			resp["errmsg"] = "订单商品表插入失败"
			o.Rollback()
			return
		}
		//订单完成，删除redis数据库中的信息
		_, err = conn.Do("hdel", "cart_"+name.(string), id)
		if err != nil {
			resp["errnum"] = 7
			resp["errmsg"] = "删除订单redis中完成订单失败"
			o.Rollback()
			return
		}
	}

	//返回数据
	o.Commit() //提交事务
	resp["errnum"] = 5
	resp["errmsg"] = "OK"
}

//支付订单
func (this *OrderController) HandlePay() {
	orderId, err := this.GetInt("OrderId")
	if err != nil {
		this.Redirect("/user/userOrder", 302)
		return
	}
	//处理数据  获取订单信息
	o := orm.NewOrm()
	var orderInfo models.OrderInfo
	orderInfo.Id = orderId
	o.Read(&orderInfo)

	pubilckey := `MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAxo7V3LO3OLCPJYw2+8X1
+bvyRnHxJpHsCO3nTkKVtce1s+FP8lFgva7ujYWxJocU8/6G0m3QQ1JIcR7FQV2A
l8FcS9Q/LAgifnDfwt3weELxJxpjVCIuHw0Nnc6coXcmgEHi01U/ss0gWdz/2RGI
JZNTByUTnT8/PK7m+6FvzrfCUFcU74PUqpTsorx/SW7r6TJiP2sKHQOkLj5Zt5BL
k/PLnGubUQwM00oJI63fn0ZOdAufGcSKutuxSnOeDYGwjWdr6KsCZ+tL7aT02UNX
2YXtWb55v+ZsuJXZ1hgbUIl9q1I3MsWnchrFmzUDVpZPLabQUomDJhsbH3SuGn7G
8QIDAQAB`
	privatekey := `MIIEpQIBAAKCAQEAxo7V3LO3OLCPJYw2+8X1+bvyRnHxJpHsCO3nTkKVtce1s+FP
8lFgva7ujYWxJocU8/6G0m3QQ1JIcR7FQV2Al8FcS9Q/LAgifnDfwt3weELxJxpj
VCIuHw0Nnc6coXcmgEHi01U/ss0gWdz/2RGIJZNTByUTnT8/PK7m+6FvzrfCUFcU
74PUqpTsorx/SW7r6TJiP2sKHQOkLj5Zt5BLk/PLnGubUQwM00oJI63fn0ZOdAuf
GcSKutuxSnOeDYGwjWdr6KsCZ+tL7aT02UNX2YXtWb55v+ZsuJXZ1hgbUIl9q1I3
MsWnchrFmzUDVpZPLabQUomDJhsbH3SuGn7G8QIDAQABAoIBAQCdveH3KSs5NUMz
wDX6RWXJ1d9+yYycaLcMzPvCt7E6LgOTeT9LMg1aBDxuYDTBd/VUdfPj/uvCX/8/
JwPsjvzXEv1hHKhnMbs9miyaIjmlQQFWYGdi8piTgIo9wWO7/u2uXSl3XTVytfWq
jqEPcRcpSuZeOb1gYlu5uPW2GKW7oeiREL5NRKIYhU04CqHZZKbya/kyYXOCc1VQ
Jph7vwDPRzW785lwZ/eNlQsxHst4749uxD0lyTT/hegd4Qs8usnq9L9mg6oz166N
ifKOad71o6HSOc7tW3y+2FO9z6t5p0iJg1wTJBBNp13V0r7W/zYUzYZNVEZNwsIJ
zyjgEdn1AoGBAOqiEFruB0EAvTA56IN3EHCpJmBmnZqMOROe5+whODVoA+Hn+/Uj
cWrT2gOXLtX8UqSh4ZQ0wg31Xf3sgbUEJbCmQYfmTQxkm5IUWGSQJ5clPq1WOB68
gJjGPmvC3PaCqOgaDS/+7UaVOvTSpyoJMmzTJUnlt0kMJxQPNCdYOTKzAoGBANij
wymXMBUTiPzZXhTFJ6GkJq2zoxtCqDdW/rT4N9axUtTnyXM1jsDyeoF0ZL0t+RM5
PU12+1gaMMbsM2h9S6mSnOS0xw1MCc9okjEaE2vwUOjY3MjaDd5R2kB60GGL36Om
inFF48JGA77/eS4B7gZlAq13Wvf2rIgPsMgPO6HLAoGBAJZLkaZtaoAc9RL7RRFR
J1rDPy3pDXN99uG2nGEJNnQfWL1QWHjCZ1hCtBgws3Om2VlTyKei09sVHUwWP0+w
hGQPW1KuFxSlLXH5TlW8cV/EH83wuKoXnVg/RKTKgbf2IjPq3B+ucFcSKF3CwgmP
Mpm69tvLJgjInA+hXYsGXD05AoGAem4KBD4h+TdCCosZRSmqXQo0paPIgMZp5GRw
q3s/q0ApLJMVYNuaomYmX3SU3ejWQ0vs3hNotgCPIkwoPVvZ36owqnuTpmsbfDB/
teojobN9NioglA3PGp4tdpUxoH740zZyNNOnhIom4dDk/eAvUUPSgI/cRBgOpxIe
+ODK9YMCgYEAzNrh4do6IXq/Eww1RMoT5mz+shtGKMlnVds4gmAmeyG8ODD70fLX
6HG73IfHsHqFmCWWYxm7G1l4sytA29yUCJwFiBggvpc41qNlM/imivJCuiGXXrt8
RtRBKz62/5XEvb7iBkOi7MDpdreQNDNXzXlkC6sF9+RxXSzkX72Qd4g=`
	//支付    appId, aliPublicKey, privateKey string, isProduction bool
	client := alipay.New("2016093000634589", pubilckey, privatekey, false)
	var p = alipay.TradePagePay{}
	p.NotifyURL = "http://192.168.11.141:8080/payok"
	p.ReturnURL = "http://192.168.11.141:8080/payok"
	p.Subject = "品优购"
	p.OutTradeNo = orderInfo.OrderId
	p.TotalAmount = strconv.Itoa(orderInfo.TotalPrice)
	p.ProductCode = "FAST_INSTANT_TRADE_PAY"
	url, err := client.TradePagePay(p)
	if err != nil {
		fmt.Println("支付失败")
	}
	payUrl := url.String()
	this.Redirect(payUrl, 302)
}
