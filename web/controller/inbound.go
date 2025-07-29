package controller

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"strconv"
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
		// 管理员：获取全部入站
		inbounds, err = a.inboundService.GetAllInbounds()
	} else {
		// 普通用户：仅获取自己的入站
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
		jsonMsg(c, "添加", err)
		return
	}

	// 获取当前操作用户（一般为管理员）
	currentUser := session.GetLoginUser(c)
	inbound.UserId = currentUser.Id
	inbound.Enable = true
	inbound.Tag = fmt.Sprintf("inbound-%v", inbound.Port)

	// 添加入站信息
	err = a.inboundService.AddInbound(inbound)
	if err == nil {
		a.xrayService.SetToNeedRestart()

		// 🔽 自动创建新用户：用户名为入站备注，密码为 admin
		userService := service.UserService{}
		db := database.GetDB()

		// 检查是否已存在该用户名
		existing := userService.CheckUser(inbound.Remark, "admin")
		if existing == nil {
			newUser := &model.User{
				Username: inbound.Remark,
				Password: "admin", // 明文存储，不安全（建议改为加密）
				Role:     "viewer", // 默认入站自动创建的用户为只读角色
			}
			errUser := db.Create(newUser).Error
			if errUser == nil {
				// ✅ 将新建入站的 UserId 设置为新用户 ID
				inbound.UserId = newUser.Id
				_ = db.Save(inbound).Error
			}
		}
	}

	jsonMsg(c, "添加", err)
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
