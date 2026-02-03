package controllers

import (
	"github.com/gin-gonic/gin"
)

func ReportsIndex(c *gin.Context) {
	Render(c, "reports.html", gin.H{
		"Title": "Reports",
		"Page":  "reports",
	})
}
