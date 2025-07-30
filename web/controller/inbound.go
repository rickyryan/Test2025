package controller

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"strconv"
	"x-ui/database"
	"x-ui/database/model"
	"x-ui/logger"
	"x-ui/web/global"
	"x-ui/web/service"
	"x-ui/web/session"
)

type InboundController struct {
	inboundService service.InboundService
	xrayService    service.XrayService
}

func NewInboundController(g *gin.RouterGroup) *InboundController {
	a := &InboundController{}
	a.initRouter(g)
	a.startTask()
	return a
}

func (a *InboundController) initRouter(g *gin.RouterGroup) {
	g = g.Group("/inbound")

	g.POST("/list", a.getInbounds)
	g.POST("/add", a.addInbound)
	g.POST("/del/:id", a.delInbound)
	g.POST("/update/:id", a.updateInbound)
}

func (a *InboundController) startTask() {
	webServer := global.GetWebServer()
	c := webServer.GetCron()
	c.AddFunc("@every 10s", func() {
		if a.xrayService.IsNeedRestartAndSetFalse() {
			err := a.xrayService.RestartXray(false)
			if err != nil {
				logger.Error("restart xray failed:", err)
			}
		}
	})
}

func (a *InboundController) getInbounds(c *gin.Context) {
	user := session.GetLoginUser(c)

	var (
		inbounds []*model.Inbound
		err      error
	)

	if session.IsAdmin(user) {
		inbounds, err = a.inboundService.GetAllInbounds()
	} else {
		inbounds, err = a.inboundService.GetInbounds(user.Id)
	}

	if err != nil {
		jsonMsg(c, "获取入站", err)
		return
	}

	jsonObj(c, inbounds, nil)
}

func (a *InboundController) addInbound(c *gin.Context) {
	user := session.GetLoginUser(c)
	if user.Role != "admin" {
		jsonMsg(c, "权限不足，仅管理员可添加入站", errors.New("forbidden"))
		return
	}

	inbound := &model.Inbound{}
	err := c.ShouldBind(inbound)
	if err != nil {
		jsonMsg(c, "参数绑定失败", err)
		return
	}

	inbound.Enable = true
	inbound.Tag = fmt.Sprintf("inbound-%v", inbound.Port)

	db := database.GetDB()
	userService := service.UserService{}

	// 🔍 检查是否已存在同名用户
	var count int64
	db.Model(&model.User{}).Where("username = ?", inbound.Remark).Count(&count)

	if count == 0 {
		// ✅ 创建 viewer 用户
		newUser := &model.User{
			Username: inbound.Remark,
			Password: "admin",
			Role:     "viewer",
		}
		err = db.Create(newUser).Error
		if err == nil {
			inbound.UserId = newUser.Id
		}
	} else {
		// ❗ 用户已存在，则不创建，默认绑定当前操作用户（admin）
		inbound.UserId = user.Id
	}

	// ✅ 添加入站
	err = a.inboundService.AddInbound(inbound)
	jsonMsg(c, "添加入站", err)

	if err == nil {
		a.xrayService.SetToNeedRestart()
	}
}

func (a *InboundController) delInbound(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		jsonMsg(c, "删除", err)
		return
	}
	err = a.inboundService.DelInbound(id)
	jsonMsg(c, "删除", err)
	if err == nil {
		a.xrayService.SetToNeedRestart()
	}
}

func (a *InboundController) updateInbound(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		jsonMsg(c, "修改", err)
		return
	}
	inbound := &model.Inbound{
		Id: id,
	}
	err = c.ShouldBind(inbound)
	if err != nil {
		jsonMsg(c, "修改", err)
		return
	}
	err = a.inboundService.UpdateInbound(inbound)
	jsonMsg(c, "修改", err)
	if err == nil {
		a.xrayService.SetToNeedRestart()
	}
}
