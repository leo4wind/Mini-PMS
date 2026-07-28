package router

import (
	"minipms/internal/config"
	"minipms/internal/handler"
	"minipms/internal/middleware"
	"minipms/internal/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func Setup(cfg *config.Config, db *gorm.DB) *gin.Engine {
	gin.SetMode(cfg.Server.Mode)
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery(), cors())

	authSvc := service.NewAuthService(db)
	permSvc := service.NewPermService(db)
	productSvc := service.NewProductService(db)
	systemSvc := service.NewSystemService(db)
	projectSvc := service.NewProjectService(db)
	storySvc := service.NewStoryService(db)
	sprintSvc := service.NewSprintService(db)
	bugSvc := service.NewBugService(db)
	attachmentSvc := service.NewAttachmentService(db, cfg.Upload)
	dashboardSvc := service.NewDashboardService(db)

	authH := handler.NewAuthHandler(cfg, authSvc, permSvc)
	productH := handler.NewProductHandler(productSvc)
	systemH := handler.NewSystemHandler(systemSvc)
	projectH := handler.NewProjectHandler(projectSvc)
	storyH := handler.NewStoryHandler(storySvc)
	sprintH := handler.NewSprintHandler(sprintSvc)
	bugH := handler.NewBugHandler(bugSvc)
	attachmentH := handler.NewAttachmentHandler(attachmentSvc, permSvc)
	dashboardH := handler.NewDashboardHandler(dashboardSvc)

	api := r.Group("/api/v1")
	{
		api.POST("/auth/login", authH.Login)

		authed := api.Group("")
		authed.Use(middleware.JWTAuth(cfg.JWT.Secret))
		{
			authed.POST("/auth/logout", authH.Logout)
			authed.GET("/auth/me", authH.Me)
			authed.PUT("/auth/password", authH.Password)

			authed.GET("/dashboard/summary", middleware.RequirePerm(permSvc, "dashboard"), dashboardH.Summary)

			// 缺陷/需求建单需要选产品，故 list 对相关权限开放（不只 product.list）
			authed.GET("/products", middleware.RequireAnyPerm(permSvc, "product.list", "bug.create", "bug.list", "story.create", "story.list", "project.create", "project.list"), productH.List)
			authed.POST("/products", middleware.RequirePerm(permSvc, "product.create"), productH.Create)
			authed.GET("/products/:id", middleware.RequireAnyPerm(permSvc, "product.list", "bug.create", "bug.list", "story.list", "project.list"), productH.Get)
			authed.PUT("/products/:id", middleware.RequirePerm(permSvc, "product.edit"), productH.Update)
			authed.DELETE("/products/:id", middleware.RequirePerm(permSvc, "product.delete"), productH.Delete)
			authed.GET("/products/:id/projects", middleware.RequireAnyPerm(permSvc, "product.list", "project.list"), projectH.ListByProduct)

			// stories
			authed.GET("/stories", middleware.RequirePerm(permSvc, "story.list"), storyH.List)
			authed.POST("/stories", middleware.RequirePerm(permSvc, "story.create"), storyH.Create)
			authed.GET("/stories/:id", middleware.RequirePerm(permSvc, "story.list"), storyH.Get)
			authed.PUT("/stories/:id", middleware.RequirePerm(permSvc, "story.edit"), storyH.Update)
			authed.DELETE("/stories/:id", middleware.RequirePerm(permSvc, "story.delete"), storyH.Delete)
			authed.POST("/stories/:id/attachments", middleware.RequirePerm(permSvc, "story.attach"), attachmentH.UploadStory)

			// projects
			authed.GET("/projects", middleware.RequirePerm(permSvc, "project.list"), projectH.List)
			authed.POST("/projects", middleware.RequirePerm(permSvc, "project.create"), projectH.Create)
			authed.GET("/projects/:id", middleware.RequirePerm(permSvc, "project.list"), projectH.Get)
			authed.PUT("/projects/:id", middleware.RequirePerm(permSvc, "project.edit"), projectH.Update)
			authed.DELETE("/projects/:id", middleware.RequirePerm(permSvc, "project.delete"), projectH.Delete)
			authed.GET("/projects/:id/sprints", middleware.RequirePerm(permSvc, "sprint.list"), sprintH.ListByProject)
			authed.GET("/projects/:id/stories", middleware.RequirePerm(permSvc, "story.list"), storyH.ListByProject)

			// sprints
			authed.GET("/sprints", middleware.RequirePerm(permSvc, "sprint.list"), sprintH.List)
			authed.POST("/sprints", middleware.RequirePerm(permSvc, "sprint.create"), sprintH.Create)
			authed.GET("/sprints/:id", middleware.RequirePerm(permSvc, "sprint.list"), sprintH.Get)
			authed.PUT("/sprints/:id", middleware.RequirePerm(permSvc, "sprint.edit"), sprintH.Update)
			authed.DELETE("/sprints/:id", middleware.RequirePerm(permSvc, "sprint.delete"), sprintH.Delete)
			authed.GET("/sprints/:id/stories", middleware.RequirePerm(permSvc, "sprint.list"), sprintH.ListStories)
			authed.POST("/sprints/:id/stories", middleware.RequirePerm(permSvc, "sprint.linkStory"), sprintH.LinkStories)
			authed.DELETE("/sprints/:id/stories/:storyId", middleware.RequirePerm(permSvc, "sprint.linkStory"), sprintH.UnlinkStory)
			authed.GET("/sprints/:id/story-candidates", middleware.RequirePerm(permSvc, "sprint.linkStory"), sprintH.StoryCandidates)

			// bugs
			authed.GET("/bugs", middleware.RequirePerm(permSvc, "bug.list"), bugH.List)
			authed.POST("/bugs", middleware.RequirePerm(permSvc, "bug.create"), bugH.Create)
			authed.GET("/bugs/:id", middleware.RequirePerm(permSvc, "bug.list"), bugH.Get)
			authed.PUT("/bugs/:id", middleware.RequirePerm(permSvc, "bug.edit"), bugH.Update)
			authed.POST("/bugs/:id/resolve", middleware.RequirePerm(permSvc, "bug.resolve"), bugH.Resolve)
			authed.POST("/bugs/:id/close", middleware.RequirePerm(permSvc, "bug.close"), bugH.Close)
			authed.POST("/bugs/:id/activate", middleware.RequirePerm(permSvc, "bug.edit"), bugH.Activate)
			authed.DELETE("/bugs/:id", middleware.RequirePerm(permSvc, "bug.delete"), bugH.Delete)
			authed.POST("/bugs/:id/attachments", middleware.RequirePerm(permSvc, "bug.attach"), attachmentH.UploadBug)

			// attachments
			authed.GET("/attachments/:id/download", attachmentH.Download)
			authed.GET("/attachments/:id/preview", attachmentH.Preview)
			authed.DELETE("/attachments/:id", attachmentH.Delete)

			// users
			authed.GET("/users", middleware.RequireAnyPerm(permSvc, "user.list", "project.create", "project.edit", "story.create", "story.edit", "bug.create", "bug.edit"), systemH.ListUsers)
			authed.POST("/users", middleware.RequirePerm(permSvc, "user.create"), systemH.CreateUser)
			authed.GET("/users/:id", middleware.RequirePerm(permSvc, "user.list"), systemH.GetUser)
			authed.PUT("/users/:id", middleware.RequirePerm(permSvc, "user.edit"), systemH.UpdateUser)
			authed.POST("/users/:id/disable", middleware.RequirePerm(permSvc, "user.disable"), systemH.DisableUser)
			authed.POST("/users/:id/enable", middleware.RequirePerm(permSvc, "user.disable"), systemH.EnableUser)
			authed.PUT("/users/:id/roles", middleware.RequirePerm(permSvc, "user.assignRole"), systemH.AssignUserRoles)

			// roles
			authed.GET("/roles", middleware.RequireAnyPerm(permSvc, "role.list", "user.assignRole", "user.create"), systemH.ListRoles)
			authed.PUT("/roles/:id", middleware.RequirePerm(permSvc, "role.edit"), systemH.UpdateRole)
			authed.GET("/roles/:id/menus", middleware.RequirePerm(permSvc, "role.assignMenu"), systemH.GetRoleMenus)
			authed.PUT("/roles/:id/menus", middleware.RequirePerm(permSvc, "role.assignMenu"), systemH.AssignRoleMenus)

			// menus
			authed.GET("/menus/tree", middleware.RequirePerm(permSvc, "menu.list"), systemH.MenuTree)
			authed.POST("/menus", middleware.RequirePerm(permSvc, "menu.create"), systemH.CreateMenu)
			authed.PUT("/menus/:id", middleware.RequirePerm(permSvc, "menu.edit"), systemH.UpdateMenu)
			authed.DELETE("/menus/:id", middleware.RequirePerm(permSvc, "menu.delete"), systemH.DeleteMenu)
		}
	}

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "up"})
	})
	return r
}

func cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}
