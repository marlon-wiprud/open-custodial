package main

import (
	eth_http "open_custodial/module/eth/http"
	eth_svc "open_custodial/module/eth/service"
	"open_custodial/pkg/config"
	"open_custodial/pkg/hsm"

	"github.com/gin-gonic/gin"
)

func main() {
	c := config.NewConfig()
	h, err := hsm.NewHSM(c.HSMLibPath, c.CU_USERNAME, c.CU_PASSWORD, c.SO_PASSWORD)
	if err != nil {
		panic(err)
	}

	svc := eth_svc.NewETHService(h)
	handler := eth_http.NewHandler(svc)

	g := gin.Default()
	v1 := g.Group("/v1")

	handler.Setup(v1)

	g.Run()
}
