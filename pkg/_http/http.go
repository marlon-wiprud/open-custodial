package _http

import (
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"

	_err "open_custodial/pkg/_err"
)

type HttpParam string

const (
	ParamLabel  HttpParam = "label"
	ParamSlotID HttpParam = "slotID"
)

func GetParamLabel(c *gin.Context) string {
	return c.Param(string(ParamLabel))
}

func GetParamSlotID(c *gin.Context) (uint64, error) {
	return strconv.ParseUint(c.Param(string(ParamSlotID)), 10, 32)
}

func ErrorResponse(c *gin.Context, err error, statusCode int) {
	c.JSON(statusCode, err)
}

func UnknownError(c *gin.Context, err error, statusCode int) {
	c.JSON(statusCode, _err.ErrorRender{
		Message: fmt.Sprintf("Unknown error: %s", err.Error()),
		Details: err.Error(),
	})
}
