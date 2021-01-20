package handlers

import (
	"context"
	"encoding/json"
	"github.com/buildboxapp/app/pkg/model"
	"net/http"
)

// Ping get user by login+pass pair
// @Summary get user by login+pass pair
// @Param login_input body models.LoginPassInput true "login data"
// @Success 200 {object} responses.Response [Result:models.ShortUser]
// @Failure 400 {object} responses.Response
// @Failure 500 {object} responses.Response
// @Router /api/v1/ping [get]
func (h *handlers) PLogin(w http.ResponseWriter, r *http.Request) {
	_, err := PLoginDecodeRequest(r.Context(), r)
	if err != nil {
		h.logger.Error(err, "[PLogin] Error function execution (PLoginDecodeRequest).")
		return
	}
	serviceResult, err := h.service.Ping(r.Context())
	if err != nil {
		h.logger.Error(err, "[PLogin] Error service execution (Ping).")
		return
	}
	response, _ := PLoginEncodeResponse(r.Context(), &serviceResult)
	if err != nil {
		h.logger.Error(err, "[PLogin] Error function execution (PLoginEncodeResponse).")
		return
	}
	err = PLoginTransportResponse(w, response)
	if err != nil {
		h.logger.Error(err, "[PLogin] Error function execution (PLoginTransportResponse).")
		return
	}

	return
}

func PLoginDecodeRequest(ctx context.Context, r *http.Request) (request *model.Pong, err error)  {
	return request, err
}

func PLoginEncodeResponse(ctx context.Context, serviceResult *model.Pong) (response model.Pong, err error)  {

	return *serviceResult, err
}

func PLoginTransportResponse(w http.ResponseWriter, response interface{}) (err error)  {
	d, err := json.Marshal(response)

	w.Write(d)
	return err
}