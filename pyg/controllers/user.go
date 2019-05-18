package controllers

import (
	"github.com/astaxie/beego"
	"fmt"
	"encoding/json"
	"regexp"
	"time"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"math/rand"
	"github.com/astaxie/beego/orm"
	"pyg/pyg/models"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk"
	"github.com/astaxie/beego/utils"
	"encoding/base64"
)

//用户控制器
type UserController struct {
	beego.Controller
}

//展示注册页面
func (this *UserController) ShowRegister() {
	this.TplName = "register.html"
}

//发送错误信息函数
func respFunc(this *beego.Controller, resp map[string]interface{}) {
	//发送数据
	this.Data["json"] = resp
	//定义发送方式,json方式
	this.ServeJSON()
}

//定义接收数据结构体
type Message struct {
	Message   string
	RequestId string
	BizId     string
	Code      string
}

//发送短信
func (this *UserController) HandleSendMsg() {
	//获取数据
	phone := this.GetString("phone")
	//定义容器
	resp := make(map[string]interface{})

	//向前端发送信息
	defer respFunc(&this.Controller, resp)
	//校验数据
	if phone == "" {
		fmt.Println("获取电话号码失败")
		//给容器赋值
		resp["errnum"] = 1
		resp["errmsg"] = "获取电话号码失败"
		return
	}

	//检查电话号码是否正确
	reg, _ := regexp.Compile(`^1[3-9][0-9]{9}$`)
	result := reg.FindString(phone)
	if result == "" {
		fmt.Println("电话号码格式错误")
		resp["errno"] = 2
		resp["errmsg"] = "电话号码格式错误"
		return
	}

	//发短信
	//发送短信   SDK调用
	client, err := sdk.NewClientWithAccessKey("cn-hangzhou", "LTAIu4sh9mfgqjjr", "sTPSi0Ybj0oFyqDTjQyQNqdq9I9akE")
	if err != nil {
		fmt.Println("电话号码格式错误")
		resp["errno"] = 3
		resp["errmsg"] = "初始化短信错误"
		return
	}
	//生成6位数随机数
	//方法二
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	msgnum := fmt.Sprintf("%06d", rnd.Int31n(1000000))

	request := requests.NewCommonRequest()
	request.Method = "POST"
	request.Scheme = "https" // https | http
	request.Domain = "dysmsapi.aliyuncs.com"
	request.Version = "2017-05-25"
	request.ApiName = "SendSms"
	request.QueryParams["RegionId"] = "cn-hangzhou"
	request.QueryParams["PhoneNumbers"] = phone
	request.QueryParams["SignName"] = "品优购"
	request.QueryParams["TemplateCode"] = "SMS_164275022"
	request.QueryParams["TemplateParam"] = "{\"code\":" + msgnum + "}"

	response, err := client.ProcessCommonRequest(request)
	if err != nil {
		fmt.Println("短信发送失败")
		resp["errno"] = 4
		resp["errmsg"] = "短信发送失败"
		return
	}

	//json数据解析
	//创建一个接受返回数据的结构体
	var message Message
	//将json数据解析为字符切片，存储在容器message中【注意】
	json.Unmarshal(response.GetHttpContentBytes(), &message)
	if message.Message != "OK" {
		fmt.Println("电话号码格式错误")
		resp["errno"] = 6
		resp["errmsg"] = message.Message
		return
	}
	//数据发送成功
	resp["errno"] = 5
	resp["errmsg"] = "发送成功"

	//传递生成的验证码给前端，做数据校验
	resp["code"] = msgnum
}

//注册操作具体实现
func (this *UserController) HandleRegister() {
	//获取数据
	phone := this.GetString("phone")
	password := this.GetString("password")
	repassword := this.GetString("repassword")
	//校验数据
	//数据完整性校验
	if phone == "" || password == "" || repassword == "" {
		fmt.Println("输入信息不完整,请重新输入！")
		this.Data["errmsg"] = "输入信息不完整,请重新输入！"
		this.TplName = "register.html"
		return
	}
	//两次输入密码是否一致校验
	if password != repassword {
		fmt.Println("两次输入密码不一致")
		this.Data["errmsg"] = "两次输入密码不一致"
		this.TplName = "register.html"
		return
	}

	//处理数据
	o := orm.NewOrm()
	var user models.User
	user.Name = phone
	user.Pwd = password
	user.Phone = phone
	o.Insert(&user)

	//返回数据 //激活页面
	//储存cookie
	this.Ctx.SetCookie("userName", user.Name, 60*10)
	//跳转到激活邮箱页面
	this.Redirect("/register-email", 302)
}

//展示激活邮箱页面
func (this *UserController) ShowEmail() {
	this.TplName = "register-email.html"
}

//激活邮箱操作
func (this *UserController) HandleEmail() {
	//获取数据
	email := this.GetString("email")
	password := this.GetString("password")
	repassword := this.GetString("repassword")
	//校验数据
	//完整性校验
	if email == "" || password == "" || repassword == "" {
		fmt.Println("数据不完整")
		this.Data["errmsg"] = "数据不完整"
		this.TplName = "register-email.html"
		return
	}
	//两次密码是否一致校验
	if password != repassword {
		fmt.Println("两次输入密码不一致")
		this.Data["errmsg"] = "两次输入密码不一致"
		this.TplName = "register-email.html"
		return
	}
	//校验邮箱格式
	reg, _ := regexp.Compile(`^\w[\w\.-]*@[0-9a-z][0-9a-z-]*(\.[a-z]+)*\.[a-z]{2,6}$`)
	result := reg.FindString(email)
	if result == "" {
		fmt.Println("邮箱格式不合格")
		this.Data["errmsg"] = "邮箱格式不合格"
		this.TplName = "register-email.html"
		return
	}

	//处理数据
	//发送邮件
	//utils     全局通用接口  工具类  邮箱配置
	config := `{"username":"13511085358@163.com","password":"mim616293","host":"smtp.163.com","port":25}`
	emailReg := utils.NewEMail(config)
	//内容配置
	emailReg.Subject = "品优购用户激活"
	emailReg.From = "13511085358@163.com"
	emailReg.To = []string{email}
	//从cookie获取用户名
	userName := this.Ctx.GetCookie("userName")
	//
	emailReg.HTML = `<a href="http://127.0.0.1:8080/active?userName=` + userName + `"> 点击激活该用户</a>`
	//发送邮件
	emailReg.Send()

	//向用户数据插入email数据    更新邮箱字段
	o := orm.NewOrm()
	var user models.User
	user.Name = userName
	err := o.Read(&user, "Name")
	if err != nil {
		fmt.Println("用户名错误")
		this.TplName = "register-email.html"
		return
	}
	user.Email = email
	o.Update(&user, "Email")

	//返回数据
	this.Ctx.WriteString("邮件已发送，请去目标邮箱激活用户！")
}

//用户激活
func (this *UserController) Active() {
	//获取数据
	userName := this.GetString("userName")
	//校验数据
	if userName == "" {
		fmt.Println("获取数据不完整")
		this.Redirect("/register-email", 302)
		return
	}
	//处理数据
	o := orm.NewOrm()
	var user models.User
	user.Name = userName
	err := o.Read(&user, "Name")
	if err != nil {
		fmt.Println("用户名错误")
		this.Redirect("/register-email", 302)
		return
	}
	user.Active = true
	o.Update(&user, "Active")

	//返回数据
	this.Redirect("/login", 302)
}

//展示用户登陆
func (this *UserController) ShowLogin() {
	//获取cookie数据，如果获取查到了，说明上一次记住用户名，不然的话，不记住用户名
	loginUserName := this.Ctx.GetCookie("loginUserName")
	//解密
	dec, _ := base64.StdEncoding.DecodeString(loginUserName)
	if loginUserName != "" {
		this.Data["checked"] = "checked"
	} else {
		this.Data["checked"] = ""
	}
	//传递数据
	this.Data["loginUserName"] = string(dec)

	//展示页面
	this.TplName = "login.html"
}

//用户登录操作实现
func (this *UserController) HandleLogin() {
	//获取数据
	name := this.GetString("name")
	pwd := this.GetString("pwd")
	//校验数据
	if name == "" || pwd == "" {
		fmt.Println("数据获取不完整")
		this.TplName = "login.html"
		return
	}
	//处理数据
	o := orm.NewOrm()
	var user models.User
	//赋值操作
	reg, _ := regexp.Compile(`^\w[\w\.-]*@[0-9a-z][0-9a-z-]*(\.[a-z]+)*\.[a-z]{2,6}$`)
	result := reg.FindString(name)
	if result != "" {
		user.Email = name
		err := o.Read(&user, "Email")
		if err != nil {
			fmt.Println("用户名不正确")
			this.Data["errmsg"] = "邮箱未注册"
			this.TplName = "login.html"
			return
		}
		if user.Pwd != pwd {
			fmt.Println("密码不正确")
			this.Data["errmsg"] = "密码不正确"
			this.TplName = "login.html"
			return
		}
	} else {
		//不是邮箱就是手机号
		user.Phone = name
		err := o.Read(&user, "Phone")
		if err != nil {
			fmt.Println("用户名不正确")
			this.Data["errmsg"] = "手机号未注册"
			this.TplName = "login.html"
			return
		}
		if user.Pwd != pwd {
			fmt.Println("密码不正确")
			this.Data["errmsg"] = "密码不正确"
			this.TplName = "login.html"
			return
		}
	}

	//校验邮箱是否激活
	if user.Active == false {
		fmt.Println("邮箱未激活")
		this.Data["errmsg"] = "邮箱未激活"
		this.TplName = "login.html"
		return
	}

	//实现自动登陆
	//实现记住用户名功能  上一次登陆成功以后，点击了记住用户名，下一次登陆的时候默认显示用户名
	remember := this.GetString("remember")
	//用户名加密
	enc := base64.StdEncoding.EncodeToString([]byte(user.Name))
	if remember == "2" {
		this.Ctx.SetCookie("loginUserName", enc, 60*60)
	} else {
		this.Ctx.SetCookie("loginUserName", enc, -1)
	}

	//设置sission
	this.SetSession("name", user.Name)
	//返回数据
	this.Redirect("/user/index", 302)
}

//退出登陆
func (this *UserController) Logout() {
	this.DelSession("name")
	this.Redirect("/user/index", 302)
}

//展示用户中心页面
func (this *UserController) ShowUserCenterInfo() {
	name := this.GetSession("name")
	o := orm.NewOrm()
	var addr models.Address
	o.QueryTable("Address").RelatedSel("User").Filter("User__Name", name.(string)).Filter("IsDefault", true).One(&addr)
	this.Data["addr"] = addr

	this.Data["index"]=1
	this.Layout = "layout.html"
	this.TplName = "user_center_info.html"
}

//展示用户中心收货地址页面
func (this *UserController) ShowSite() {
	//获取sission中的name
	username := this.GetSession("name")

	name := username.(string)
	//从表中查询数据，多表查询
	o := orm.NewOrm()
	//定义容器
	var address models.Address
	o.QueryTable("Address").RelatedSel("User").Filter("User__Name", name).Filter("IsDefault", true).One(&address)
	//传递数据
	this.Data["address"] = address
	this.Data["index"]=3
	//this.Data["username"] = username
	//返回数据
	this.Layout = "layout.html"
	this.TplName = "user_center_site.html"
}

//添加地址
func (this *UserController) HandleSite() {
	//获取数据
	receiver := this.GetString("receiver")
	addr := this.GetString("addr")
	postCode := this.GetString("postCode")
	phone := this.GetString("phone")
	//校验数据
	if receiver == "" || addr == "" || postCode == "" || phone == "" {
		fmt.Println("")
		this.Redirect("user/site", 302)
		return
	}
	//处理数据
	o := orm.NewOrm()
	var useraddr models.Address
	useraddr.Addr = addr
	useraddr.PostCode = postCode
	useraddr.Phone = phone
	useraddr.Receiver = receiver
	//获取关联表数据
	var user models.User
	name := this.GetSession("name")
	user.Name = name.(string)
	o.Read(&user, "Name")
	useraddr.User = &user

	var oldAddress models.Address
	err := o.QueryTable("Address").RelatedSel("User").Filter("User__Name", user.Name).Filter("IsDefault", true).One(&oldAddress)
	if err != nil {
		oldAddress.IsDefault = false
		o.Update(&oldAddress)
	}
	useraddr.IsDefault = true
	_, err = o.Insert(&useraddr)
	if err != nil {
		fmt.Println("插入失败")
		this.TplName = "user_center_site.html"
	}
	//返回
	this.Redirect("/user/site", 302)
}

//展示个人信息页面
func (this *UserController) ShowPerinfo() {
	name := this.GetSession("name")
	o := orm.NewOrm()
	var addr models.Address
	o.QueryTable("Address").RelatedSel("User").Filter("User__Name", name.(string)).Filter("IsDefault", true).One(&addr)
	this.Data["addr"] = addr
	this.Data["index"]=1
	this.Layout = "layout.html"
	this.TplName = "user_center_info.html"
}

//展示订单信息
func (this *UserController)ShowOrder(){
	//username := this.GetSession("name")

	this.Data["index"]=2
	this.Layout="layout.html"
	this.TplName="user_center_order.html"
}