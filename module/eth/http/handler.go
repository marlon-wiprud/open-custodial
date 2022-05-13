package eth_http

import (
	"net/http"
	eth_svc "open_custodial/module/eth/service"
	"open_custodial/pkg/_err"
	"open_custodial/pkg/_http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service eth_svc.ETHService
}

type createAddressForm struct {
	Label string `json:"label"`
}

func NewHandler(s eth_svc.ETHService) *Handler {
	return &Handler{s}
}

func (h *Handler) Setup(r *gin.RouterGroup) {
	r.POST("/address", h.createAddress)
	r.GET("/address/:label", h.getAddress)
	r.GET("/slotaddress/:slotID", h.getSlotAddress)
}

func newCreateAddressForm(c *gin.Context) (f createAddressForm, err error) {
	err = c.BindJSON(&f)
	return f, err
}

func (h *Handler) createAddress(c *gin.Context) {

	f, err := newCreateAddressForm(c)
	if err != nil {
		_http.ErrorResponse(c, _err.NewBadFormErr(err), http.StatusBadRequest)
		return
	}

	addr, err := h.service.CreateAddress(f.Label)
	if err != nil {
		switch e := err.(type) {
		case _err.DuplicateLabel:
			_http.ErrorResponse(c, e, http.StatusBadRequest)
			return
		default:
			_http.UnknownError(c, e, http.StatusInternalServerError)
			return
		}
	}

	c.JSON(http.StatusOK, addr)
}

func (h *Handler) getAddress(c *gin.Context) {
	label := _http.GetParamLabel(c)
	addr, err := h.service.GetAddressByLabel(label)
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	c.JSON(http.StatusOK, addr)

}

func (h *Handler) getSlotAddress(c *gin.Context) {
	slotID, err := _http.GetParamSlotID(c)
	if err != nil {
		_http.ErrorResponse(c, _err.NewError(err, "invalid slotID parameter"), http.StatusBadRequest)
		return
	}

	addr, err := h.service.GetSlotAddress(uint(slotID))
	if err != nil {
		_http.ErrorResponse(c, _err.NewError(err, "unable to get address by slot"), http.StatusBadRequest)
		return
	}

	c.JSON(http.StatusOK, addr)

}
