package routes

import (
	"stok-hadiah/controllers"
	"stok-hadiah/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterWebRoutes(r *gin.Engine) {
	r.Use(middleware.UserMiddleware())

	r.GET("/", controllers.LoginPage)
	r.GET("/login", controllers.LoginPage)
	r.POST("/login", controllers.LoginPost)
	r.POST("/register", controllers.CreateUser)
	r.GET("/logout", controllers.Logout)
	r.GET("/guest/:store_id", controllers.GuestQueuePage)
	r.POST("/guest/ticket", controllers.GuestQueuePrint)
	r.GET("/guest/ticket/:id", controllers.GuestTicketShow)

	r.GET("/view-queue/:store_id", controllers.ViewQueuePage)
	r.GET("/ws/view-queue/:store_id", controllers.ViewQueueWS)

	auth := r.Group("/")
	auth.Use(middleware.AuthRequired(), middleware.PermissionContext())
	{
		auth.GET("/dashboard", controllers.DashboardIndex)
		auth.GET("/dashboard/queue/state", controllers.DashboardQueueState)
		auth.POST("/dashboard/counter/status", controllers.DashboardCounterStatus)
		auth.POST("/dashboard/queue/next", controllers.QueueCallNext)
		auth.POST("/dashboard/queue/recall", controllers.QueueRecall)
		auth.POST("/dashboard/queue/done", controllers.QueueDone)
		auth.POST("/dashboard/queue/skip", controllers.QueueSkip)
		auth.POST("/dashboard/queue/recall-skipped", controllers.QueueRecallSkipped)
		auth.GET("/reports", controllers.ReportsIndex)

		auth.GET("/users", middleware.RequirePermission("user_management_access"), controllers.UserIndex)
		auth.POST("/users", middleware.RequirePermission("user_create"), controllers.UserStore)
		auth.POST("/users/update", middleware.RequirePermission("user_edit"), controllers.UserUpdate)
		auth.GET("/users/delete/:id", middleware.RequirePermission("user_delete"), controllers.UserDelete)
		auth.GET("/role", controllers.RoleIndex)
		auth.GET("/roleForm", controllers.RoleFormIndex)
		auth.GET("/role/:id/edit", middleware.RequirePermission("role_edit"), controllers.RoleEdit)
		auth.POST("/role", middleware.RequirePermission("role_create"), controllers.RoleStore)
		auth.POST("/role/update", middleware.RequirePermission("role_edit"), controllers.RoleUpdate)
		auth.GET("/role/delete/:id", middleware.RequirePermission("role_delete"), controllers.RoleDelete)

		auth.GET("/counters", controllers.CounterIndex)
		auth.POST("/counters", controllers.CounterStore)
		auth.POST("/counters/update", controllers.CounterUpdate)
		auth.GET("/counters/delete/:id", controllers.CounterDelete)

		auth.GET("/counter_staff", controllers.CounterStaffIndex)
		auth.POST("/counter_staff", controllers.CounterStaffStore)
		auth.POST("/counter_staff/update", controllers.CounterStaffUpdate)
		auth.GET("/counter_staff/delete/:id", controllers.CounterStaffDelete)
	}
}
