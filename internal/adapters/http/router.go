package http

import (
	"github.com/gin-gonic/gin"
	"github.com/maneeshsagar/tps/internal/application"
)

func NewRouter(svc application.TransferServiceIntf) *gin.Engine {
	r := gin.Default()
	h := NewHandler(svc)

	r.GET("/health", h.HealthCheck)
	r.POST("/accounts", h.CreateAccount)
	r.GET("/accounts/:account_id", h.GetAccount)

	// this endpoint will perform a synchronous transfer and return the result immediately
	r.POST("/transactions", h.CreateTransaction)

	// this is a new endpoint for creating async transactions and checking their status
	r.POST("/async-transactions", h.CreateAsyncTransaction)
	r.GET("/async-transactions/:id/status", h.GetAsyncTransactionStatus)

	return r
}
