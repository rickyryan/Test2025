package controller

import (
	"errors"
	"github.com/gin-gonic/gin"
	"x-ui/web/service"
	"x-ui/web/session"
)

type XUIController struct {
	BaseController

	inboundController *InboundController
	settingController *SettingController
}

func NewXUIController(g *gin.RouterGroup) *XUIController {
	a := &XUIController{}
	a.initRouter(g)
	return a
}

func (a *XUIController) initRouter(g *gin.RouterGroup) {
	g = g.Group("/xui")
	g.Use(a.checkLogin)

	g.GET("/", a.index)
	g.GET("/inbounds", a.inbounds)
	g.GET("/setting", a.setting)

	// ✅ 注册用户信息更新接口
	g.POST("/user/update", a.updateUser)

	a.inboundController = NewInboundController(g)
	a.settingController = NewSettingController(g)
}

func (a *XUIController) index(c *gin.Context) {
	html(c, "index.html", "系统状态", nil)
}

func (a *XUIController) inbounds(c *gin.Context) {
	html(c, "inbounds.html", "入站列表", nil)
}

func (a *XUIController) setting(c *gin.Context) {
	html(c, "setting.html", "设置", nil)
}

func (a *XUIController) updateUser(c *gin.Context) {
	user := session.GetLoginUser(c)

	var param struct {
		Id       int    `json:"id"`
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := c.ShouldBindJSON(&param); err != nil {
		jsonMsg(c, "参数错误", err)
		return
	}

	// ✅ 权限控制：仅本人或管理员可修改
	if !session.IsSelfOrAdmin(user, param.Id) {
		jsonMsg(c, "权限不足，仅允许修改自己的信息", errors.New("forbidden"))
		return
	}

	err := (&service.UserService{}).UpdateUser(param.Id, param.Username, param.Password)
	jsonMsg(c, "更新用户信息", err)
}
