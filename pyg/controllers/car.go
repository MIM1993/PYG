package controllers

import (
	"github.com/astaxie/beego"
	"fmt"
	"github.com/gomodule/redigo/redis"
	"github.com/astaxie/beego/orm"
	"pyg/pyg/models"
)

type CartController struct {
	beego.Controller
}

//添加购物车
func (this *CartController) HandleAddCar() {
	//获取数据
	goodsId, err1 := this.GetInt("goodsId") //商品ID
	goodsCount, err2 := this.GetInt("goodsCount")
	fmt.Println(goodsId,goodsCount)
	fmt.Println(err1,"-----",err2)

	//校验数据
	resp := make(map[string]interface{})   //定义容器返回错误从信息
	defer respFunc(&this.Controller, resp) //最后调用函数将错误信息返回，继承父类的方法
	if err1 != nil || err2 != nil { //校验是否有错误
		fmt.Println("获取数据错误")
		resp["errnum"] = 1          //定义错误码
		resp["errmsg"] = "获取数据信息错误" //定义错误信息
		return
	}

	//检验是否登陆
	name := this.GetSession("name")
	if name == nil {
		fmt.Println("用户未登陆，请到登陆界面登陆")
		resp["errnum"] = 2                //定义错误码
		resp["errmsg"] = "用户未登陆，请到登陆界面登陆" //定义错误信息
		return
	}

	//处理数据
	//o := orm.NewOrm()
	//var goods models.GoodsSKU
	//goods.Id = goodsId
	//err := o.Read(&goods)
	//if err != nil {
	//	fmt.Println("商品id不存在")
	//	resp["errnum"] = 6         //定义错误码
	//	resp["errmsg"] = "商品id不存在" //定义错误信息
	//	return
	//}
	//

	//校验请求数量是否超出库存
	//if goods.Stock < goodsCount {
	//	fmt.Println("库存不足")
	//	resp["errnum"] = 7     //定义错误码
	//	resp["errmsg"] = "库存不足" //定义错误信息
	//	return
	//}

	//储存到redis
	conn, err := redis.Dial("tcp", "127.0.0.1:6379")
	if err != nil {
		fmt.Println("redis连接失败")
		resp["errnum"] = 3           //定义错误码
		resp["errmsg"] = "redis连接失败" //定义错误信息
		return
	}
	defer conn.Close()

	//读取已经存在的信息
	rep, err := conn.Do("hget", "cart_"+name.(string), goodsId)
	//助手函数
	oldgoodsCount, _ := redis.Int(rep, err)

	//存入redis数据库
	_, err = conn.Do("hset", "cart_"+name.(string), goodsId, goodsCount+oldgoodsCount)
	if err != nil {
		fmt.Println("添加商品失败")
		resp["errnum"] = 4        //定义错误码
		resp["errmsg"] = "添加商品失败" //定义错误信息
		return
	}

	resp["errnum"] = 5      //定义错误码
	resp["errmsg"] = "添加成功" //定义错误信息
}

//展示购物车操作
func (this *CartController)ShowCart(){
	//获取的当前用户，从redis中获取购物车数据
	//连接redis
	conn,err :=redis.Dial("tcp","127.0.0.1:6379")
	if err!=nil{
		this.Redirect("/index_sx",302)
		return
	}
	name := this.GetSession("name")
	result,err :=redis.Ints(conn.Do("hgetall","cart_"+name.(string)))
	if err!=nil{
		this.Redirect("/index_sx",302)
		return
	}

	//定义容器装载数据[]map[string]interface{}
	var goods []map[string]interface{}
	//获取用户数据
	o:=orm.NewOrm()
	//总价
	totalPrice:=0
	totalCount:=0
	//商品数量

	for i:=0;i<len(result);i+=2{
		//定义行容器
		temp:=make(map[string]interface{})
		//获取商品信息
		var goodsSku models.GoodsSKU
		goodsSku.Id=result[i]//商品属性值
		o.Read(&goodsSku)

		//对行容器赋值
		temp["goodsSku"]=goodsSku
		temp["count"]=result[i+1]
		//小节
		littlePrice:=result[i+1]*goodsSku.Price
		temp["littlePrice"]=littlePrice
		//总价
		totalPrice+=littlePrice
		//商品数量
		totalCount++
		//把行容器添加到大容器里面
		goods = append(goods,temp)
	}
	this.Data["goods"]=goods
	this.Data["totalPrice"]=totalPrice
	this.Data["totalCount"]=totalCount
	this.TplName="cart.html"
}


//处理添加购物车数量
func (this *CartController)HandleUpCart(){
	//获取数据
	count,err1 := this.GetInt("count")
	goodsId,err2 :=this.GetInt("goodsId")
	//校验数据
	//定义容器
	resp := make(map[string]interface{})
	defer respFunc(&this.Controller,resp)
	if err1!=nil || err2!=nil{
		resp["errnum"]=1
		resp["errmsg"]="传输数据不完整"
		return
	}
	//登陆检验
	name :=this.GetSession("name")
	if name==nil{
		resp["errnum"]=2
		resp["errmsg"]="当前用户未登录"
		return
	}

	//连接redis更新数据
	conn,err:=redis.Dial("tcp","127.0.0.1:6379")
	if err!=nil{
		resp["errnum"]=3
		resp["errmsg"]="redis链接错误"
		return
	}
	defer conn.Close()
	//更新数据
	_,err =conn.Do("hset","cart_"+name.(string),goodsId,count)
	if err!=nil{
		resp["errnum"]=4
		resp["errmsg"]="redis写入失败"
		return
	}
	resp["errnum"]=5
	resp["errmsg"]="OK"
}

//删除购物车
func(this *CartController)HandleDeleteCart(){
	//获取数据
	goodsId,err :=this.GetInt("goodsId")
	//定义容器储存错误
	resp:=make(map[string]interface{})
	defer respFunc(&this.Controller,resp)
	//校验数据
	if err!=nil{
		resp["errnum"]=1
		resp["errmsg"]="删除链接错误"
		return
	}
	name :=this.GetSession("name")
	if name==nil{
		resp["errnum"]=2
		resp["errmsg"]="当前用户不在登录状态"
		return
	}
	//连接redis
	conn,err :=redis.Dial("tcp","127.0.0.1:6379")
	if err!=nil{
		resp["errnum"]=3
		resp["errmsg"]="服务器异常"
		return
	}
	//删除redis中的数据
	_,err =conn.Do("hdel","cart_"+name.(string),goodsId)
	if err!=nil{
		resp["errnum"]=4
		resp["errmsg"]="数据库异常"
		return
	}
	//成功
	resp["errnum"]=5
	resp["errmsg"]="OK"
}