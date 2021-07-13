package iam

import (
	"encoding/json"
	"fmt"
	"github.com/buildboxapp/app/pkg/model"
)

// отправляем старый X-Auth-Access-токен
// получаем X-Auth-Access токен (два токена + текущая авторизационная сессия)
// этот ключ добавляется в куки или сохраняется в сервисе как ключ доступа
func (o *iam) Refresh(token string) (result string, err error) {
	var res model.Response

	_, err = o.utils.Curl("POST", o.cfg.UrlIam + "/token/refresh", token, &res, map[string]string{})
	result = fmt.Sprint(res.Data)

	return result, err
}

func (o *iam) ProfileGet(sessionID string) (result model.ProfileData, err error) {
	var res model.Response

	_, err = o.utils.Curl("GET", o.cfg.UrlIam + "/profile/"+sessionID, "", &res, map[string]string{})
	if err != nil {
		return result, err
	}

	b := fmt.Sprint(res.Data)
	err = json.Unmarshal([]byte(b), &result)

	return result, err
}

func (o *iam) ProfileList(sessionID string) (result string, err error) {
	var res model.Response

	_, err = o.utils.Curl("GET", o.cfg.UrlIam + "/profile/list", sessionID, &res, map[string]string{})
	result = fmt.Sprint(res.Data)

	return result, err
}