package controllers

import (
	"github.com/astaxie/beego"
	"fmt"
	"github.com/astaxie/beego/orm"
	"pyg/pyg/models"
)

type CartController struct {
	beego.Controller
}

//添加购物车
func (this *CartController)HandleAddCar(){
	//获取数据
	skuid,err1:=this.GetInt("skuid")
	goodsCount,err2 :=this.GetInt("goodsCount")
	//校验数据
	resp:=make(map[string]interface{})  //定义容器返回错误从信息
	defer this.ServeJSON()  //最后将错误信息返回
	if err1 != nil || err2 != nil{ //校验是否有错误
		fmt.Println("获取数据错误")
		resp["errnum"]=1				//定义错误码
		resp["errmsg"]="获取数据信息错误" //定义错误信息
		this.Data["json"]=resp      //以json数据格式返回信息
		return
	}

	//处理数据
	o:=orm.NewOrm()
	var goods models.GoodsSKU
	goods.Id=skuid
	err :=o.Read(&goods)
	if err!=nil{
		fmt.Println("商品id不存在")
		resp["errnum"]=2				//定义错误码
		resp["errmsg"]="商品id不存在" //定义错误信息
		this.Data["json"]=resp      //以json数据格式返回信息
		return
	}
	//校验请求数量是否超出库存
	if goods.Stock<goodsCount{
		fmt.Println("库存不足")
		resp["errnum"]=3				//定义错误码
		resp["errmsg"]="库存不足" //定义错误信息
		this.Data["json"]=resp      //以json数据格式返回信息
		return
	}

	//校验登陆状态
	name :=this.GetSession("name")
	if name == nil{
		fmt.Println("用户未登陆，请先登陆")
		resp["errnum"]=4				//定义错误码
		resp["errmsg"]="用户未登陆，请先登陆" //定义错误信息
		this.Data["json"]=resp      //以json数据格式返回信息
		return
	}
}