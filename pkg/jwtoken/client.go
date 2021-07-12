package jwtoken

import (
	"fmt"
	"github.com/buildboxapp/app/pkg/model"
)

// отправляем старый X-Auth-Access-токен
// получаем X-Auth-Access токен (два токена + текущая авторизационная сессия)
// этот ключ добавляется в куки или сохраняется в сервисе как ключ доступа
func (o *jwtoken) Refresh(token string) (result string, err error) {
	var res model.Response
	_, err = o.utils.Curl("POST", o.cfg.UrlIam + "/token/refresh", token, &res, map[string]string{})

	result = fmt.Sprint(res.Data)

	return result, err
}