package routers

import (
	"pyg/pyg/controllers"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/context"
)

//过滤器函数
func filFunc(ctx *context.Context){
	name := ctx.Input.Session("name")
	if name == nil{
		ctx.Redirect(302,"/login")
		return
	}
}

func init() {
//----------------陆游过滤器-----------------------------------
	beego.InsertFilter("/user/*",beego.BeforeExec,filFunc)

//----------------------------------用户模块-----------------------------------------------------------------------
    //用户注册
    beego.Router("/register",&controllers.UserController{},"get:ShowRegister;post:HandleRegister")

    //发送短信
    beego.Router("/sendmsg",&controllers.UserController{},"post:HandleSendMsg")

    //邮箱激活
    beego.Router("/register-email",&controllers.UserController{},"get:ShowEmail;post:HandleEmail")

    //用户激活
    beego.Router("/active",&controllers.UserController{},"get:Active")

    //用户登陆
    beego.Router("/login",&controllers.UserController{},"get:ShowLogin;post:HandleLogin")

	//退出登陆
	beego.Router("/user/logout",&controllers.UserController{},"get:Logout")

	//展示用户中心页
	beego.Router("/user/userCenterInfo",&controllers.UserController{},"get:ShowUserCenterInfo")

	//展示与操作用户中心地址页面
	beego.Router("/user/site",&controllers.UserController{},"get:ShowSite;post:HandleSite")

	//展示个人信息页面
	beego.Router("/user/perinfo",&controllers.UserController{},"get:ShowPerinfo")

	//展示订单页面
	beego.Router("/user/order",&controllers.UserController{},"get:ShowOrder")


	//----------------------------------------货物模块--------------------------------------------------------------
    //首页展示
    beego.Router("/user/index",&controllers.GoodController{},"get:ShowIndex")

    //海鲜
    beego.Router("/index_sx",&controllers.GoodController{},"get:ShowIndex_sx")

    //商品详情
    beego.Router("/goodsDetail",&controllers.GoodController{},"get:ShowDetail")

    //商品list展示
    beego.Router("/goodsType",&controllers.GoodController{},"get:ShowList")

    //商品搜索
    beego.Router("/search",&controllers.GoodController{},"post:Seach")

    //------------------购物车模块-------------------------------------------------------------
    //添加购物车操作
    beego.Router("/addCart",&controllers.CartController{},"post:HandleAddCar")
    
    //展示购物车
    beego.Router("/user/showCart",&controllers.CartController{},"get:ShowCart")

	//处理添加购物车数量
	beego.Router("/upCart",&controllers.CartController{},"post:HandleUpCart")

	//删除购物车
	beego.Router("/deleteCart",&controllers.CartController{},"post:HandleDeleteCart")

	//--------------------------订单操作----------------------------------------------------------------------
	//展示订单操作
	beego.Router("/user/addOrder",&controllers.OrderController{},"post:ShowOrder")

	//提交订单操作
	beego.Router("/pushOrder",&controllers.OrderController{},"post:HandlePushOrder")
}

