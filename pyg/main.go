package main

import (
	_ "pyg/pyg/routers"
	"github.com/astaxie/beego"
	_ "pyg/pyg/models"
	"strings"
)

func main() {
	beego.AddFuncMap("hide", hide)
	beego.Run()
}

//隐藏电话号码
func hide(tel string) string {
	//str := tel[0:4] + "****" + tel[8:]  //效率差
	selic  := []string {tel[0:4],"****",tel[8:]}
	str := strings.Join(selic,"")
	return str
}
