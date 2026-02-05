package controllers

import (
	"net/http"
	"strconv"

	"stok-hadiah/config"
	"stok-hadiah/realtime"
	"stok-hadiah/repositories"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var queueWSUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func ViewQueueWS(c *gin.Context) {
	storeID, _ := strconv.Atoi(c.Param("store_id"))
	if storeID <= 0 {
		storeRepo := &repositories.StoreRepository{DB: config.DB}
		store, err := storeRepo.GetFirstActive()
		if err != nil {
			c.Status(http.StatusInternalServerError)
			return
		}
		storeID = store.StoreID
	}

	conn, err := queueWSUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	client := realtime.NewClient(conn, storeID)
	realtime.QueueHub.Register(client)

	if payload, err := buildQueueViewPayload(storeID, "sync"); err == nil {
		_ = client.WriteJSON(payload)
	}

	go func() {
		defer realtime.QueueHub.Unregister(client)
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				break
			}
		}
	}()
}
