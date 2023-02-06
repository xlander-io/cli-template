package api

import (
	"net/http"

	"github.com/coreservice-io/cli-template/cmd_default/http/api/http_middleware"
	"github.com/coreservice-io/cli-template/src/common/captcha"
	"github.com/coreservice-io/cli-template/src/common/http/api"
	"github.com/labstack/echo/v4"
)

// @Msg_Resp_Captcha
type Msg_Resp_Captcha struct {
	api.API_META_STATUS
	Id      string `json:"id"`
	Content string `json:"content"`
}

func configCaptcha(httpServer *echo.Echo) {
	// user
	httpServer.GET("/api/user/captcha", getCaptchaHandler, http_middleware.MID_IP_Action_SL(1, 3))
}

// @Summary      get captcha
// @Tags         captcha
// @Produce      json
// @response 	 200 {object} Msg_Resp_Captcha "result"
// @Router       /api/captcha [get]
func getCaptchaHandler(ctx echo.Context) error {

	res := &Msg_Resp_Captcha{}

	id, base64Code, err := captcha.GenCaptcha()
	if err != nil {
		// error gen captcha
		res.MetaStatus(-1, "gen captcha err:"+err.Error())
		return ctx.JSON(http.StatusOK, res)
	}
	if id == "" || base64Code == "" {
		// error gen captcha
		res.MetaStatus(-1, "gen captcha err")
		return ctx.JSON(http.StatusOK, res)
	}

	res.MetaStatus(1, "success")
	res.Id = id
	res.Content = base64Code
	return ctx.JSON(http.StatusOK, res)
}
