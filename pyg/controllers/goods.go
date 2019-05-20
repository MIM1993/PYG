package controllers

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"pyg/pyg/models"
	"fmt"
	"math"
	"github.com/gomodule/redigo/redis"
)

//货物控制器
type GoodController struct {
	beego.Controller
}

//展示首页
func (this *GoodController) ShowIndex() {
	//获取一级菜单数据
	o := orm.NewOrm()
	//创建容器
	var oneclass []models.TpshopCategory
	//获取sission，判断是否登陆
	name := this.GetSession("name")
	if name != nil {
		this.Data["name"] = name.(string)
	} else {
		this.Data["name"] = ""
	}

	o.QueryTable("TpshopCategory").Filter("Pid", 0).All(&oneclass)
	//获取二菜单级数据
	//定义总容器  map切片
	var types []map[string]interface{}
	//遍历查出的一级数据
	for _, v := range oneclass {
		//定义行容器
		t := make(map[string]interface{})
		var twoclass []models.TpshopCategory
		o.QueryTable("TpshopCategory").Filter("Pid", v.Id).All(&twoclass)
		t["t1"] = v
		t["t2"] = twoclass
		//追加进总容器
		types = append(types, t)
	}
	//获取三级菜单数据
	for _, v1 := range types {
		//定义储存二级菜单总容器
		var erji []map[string]interface{}
		for _, v2 := range v1["t2"].([]models.TpshopCategory) {
			//二级菜单行容器
			t := make(map[string]interface{})
			//定义存储三级菜单数据容器
			var threeclass []models.TpshopCategory
			o.QueryTable("TpshopCategory").Filter("Pid", v2.Id).All(&threeclass)

			t["t22"] = v2
			t["t23"] = threeclass
			erji = append(erji, t)
		}
		v1["t3"] = erji
	}

	//传递数据给前端
	this.Data["types"] = types
	this.TplName = "index.html"
}

//海鲜
func (this *GoodController) ShowIndex_sx() {
	//获取海鲜首页内容
	//获取商品类型
	o := orm.NewOrm()
	var goodtypes []models.GoodsType
	o.QueryTable("GoodsType").All(&goodtypes)
	this.Data["GoodsTypes"] = goodtypes

	//获取论波图
	var goodsBanners []models.IndexGoodsBanner
	o.QueryTable("IndexGoodsBanner").OrderBy("Index").All(&goodsBanners)
	this.Data["goodsBanners"] = goodsBanners

	//促销商品
	var promotionBanners []models.IndexPromotionBanner
	o.QueryTable("IndexPromotionBanner").OrderBy("Index").All(&promotionBanners)
	this.Data["promotions"] = promotionBanners

	//二级联动
	//定义总容器
	var goods []map[string]interface{}
	for _, v := range goodtypes {
		var textGoods []models.IndexTypeGoodsBanner
		var imageGoods []models.IndexTypeGoodsBanner
		qs := o.QueryTable("IndexTypeGoodsBanner").RelatedSel("GoodsType", "GoodsSKU").Filter("GoodsType__Id", v.Id).OrderBy("Index")
		//文字商品
		qs.Filter("DisplayType", 0).All(&textGoods)
		//图片商品
		qs.Filter("DisplayType", 1).All(&imageGoods)
		//定义行容器
		temp := make(map[string]interface{})
		temp["goodsType"] = v
		temp["textGoods"] = textGoods
		temp["imageGoods"] = imageGoods
		//追加
		goods = append(goods, temp)
	}
	this.Data["goods"] = goods
	this.TplName = "index_sx.html"
}

//展示详情
func (this *GoodController) ShowDetail() {
	//获取数据
	Id, err := this.GetInt("Id")
	//校验数据
	if err != nil {
		fmt.Println("商品连接错误")
		this.Redirect("/index_sx", 302)
		return
	}
	//处理数据
	o := orm.NewOrm()
	var goodsku models.GoodsSKU
	//goodsku.Id=Id
	//o.Read(&goodsku)
	//获取商品详情
	o.QueryTable("GoodsSKU").RelatedSel("Goods", "GoodsType").Filter("Id", Id).One(&goodsku)

	//获取同一类型的新品推荐
	//定义容器
	var newgoods []models.GoodsSKU
	qs := o.QueryTable("GoodsSKU").RelatedSel("GoodsType").Filter("GoodsType__Name", goodsku.GoodsType.Name)
	qs.OrderBy("-Time").Limit(2, 0).All(&newgoods)

	//储存浏览数据
	name := this.GetSession("name")
	if name !=nil{
		conn,err := redis.Dial("tcp","127.0.0.1:6379")
		if err!=nil{
			defer conn.Close()
			fmt.Println("redis连接错误")
			this.Redirect("/index_sx", 302)
			return
		}else{
			defer conn.Close()
			conn.Do("lrem","history_"+name.(string),0,Id) //删除重复数据
			conn.Do("lpush","history_"+name.(string),Id)  //左插入
		}
	}

	//返回数据
	this.Data["Id"]=Id
	this.Data["newgoods"] = newgoods
	this.Data["goodsku"] = goodsku
	this.TplName = "detail.html"
}

//获取显示页码
func PageEdit(pageCount int, pageIndex int) []int {
	//定义容器
	var pages []int
	//不足5页
	if pageCount < 5 {
		for i := 1; i <= pageCount; i++ {
			pages = append(pages, i)
		}
	} else if pageIndex <= 3 {
		for i := 1; i <= 5; i++ {
			pages = append(pages, i)
		}
	} else if pageIndex >= pageCount-2 {
		for i := pageCount - 4; i <= pageCount; i++ {
			pages = append(pages, i)
		}
	} else {
		for i := pageIndex - 2; i <= pageIndex+2; i++ {
			pages = append(pages, i)
		}
	}
	return pages
}

//展示list
func (this *GoodController) ShowList() {
	//获取数据
	Id, err := this.GetInt("Id")
	//校验数据
	if err != nil {
		fmt.Println("")
		this.Redirect("/index_sx", 302)
		return
	}
	//处理数据
	//展示列表页面，并且实现排序
	o := orm.NewOrm()
	var goods []models.GoodsSKU
	qs := o.QueryTable("GoodsSKU").RelatedSel("GoodsType").Filter("GoodsType__Id", Id)
	//新品推荐
	var newgoods []models.GoodsSKU
	qs.OrderBy("-Time").Limit(2, 0).All(&newgoods)

	//处理页码问题【重点】
	conut, err := qs.Count()                                        //获取总数据条数
	pageSize := 1                                                   //定义每页数据条数
	pageCount := int(math.Ceil(float64(conut) / float64(pageSize))) //计算一共有几页
	pageIndex, _ := this.GetInt("pageIndex", 1)                     //获取当前页页码
	pages := PageEdit(pageCount, pageIndex)                         //获取分页页数展示，调用页码生成函数
	this.Data["pages"] = pages                                      //将页码数据传向前台

	//上一页下一页问题
	var prePage, nextPage int //设置上一页、下一页的值
	//范围设置
	if pageIndex-1 <= 0 { //上一页范围设置
		prePage = 1
	} else {
		prePage = pageIndex - 1
	}
	if pageIndex+1 >= pageCount { //下一页范围设置
		nextPage = pageCount
	} else {
		nextPage = pageIndex + 1
	}
	//传递数据
	this.Data["prePage"] = prePage
	this.Data["nextPage"] = nextPage

	//为查询每页显示数据的查询语句加上限制，在指定位置查询指定条数【重点】
	qs = qs.Limit(pageSize, (pageIndex-1)*pageSize)
	//处理排序问题  【重点】
	sort := this.GetString("sort") //接收排序条件   【价格，人气】
	if sort == "" {
		qs.All(&goods) //默认查询
	} else if sort == "Price" {
		qs.OrderBy("Price").All(&goods) //按价格排序
	} else {
		qs.OrderBy("-Sales").All(&goods) //按销量排序
	}

	//传递数据
	this.Data["sort"] = sort
	this.Data["Id"] = Id
	this.Data["newgoods"] = newgoods
	this.Data["goods"] = goods
	this.TplName = "list.html"
}

//搜索商品
func (this *GoodController) Seach() {
	//获取数据
	goodsName := this.GetString("goodsName")
	//校验数据
	if goodsName == "" {
		this.Redirect("/seach", 302)
		return
	}
	//处理数据
	o := orm.NewOrm()
	var goods []models.GoodsSKU                                               //定义容器
	o.QueryTable("GoodsSKU").Filter("Name__icontains", goodsName).All(&goods) //模糊查询
	//返回数据
	this.Data["goods"] = goods
	this.TplName = "search.html"
}
