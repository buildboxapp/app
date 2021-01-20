package handlers

import (
	"context"
	"encoding/json"
	"github.com/buildboxapp/app/pkg/model"
	"net/http"
)

// Ping get user by login+pass pair
// @Summary get user by login+pass pair
// @Param login_input body model.Pong true "login data"
// @Success 200 {object} model.Pong [Result:model.Pong]
// @Failure 400 {object} model.Pong
// @Failure 500 {object} model.Pong
// @Router /api/v1/ping [get]
func (h *handlers) Ping(w http.ResponseWriter, r *http.Request) {
	_, err := PingDecodeRequest(r.Context(), r)
	if err != nil {
		h.logger.Error(err, "[PLogin] Error function execution (PLoginDecodeRequest).")
		return
	}
	serviceResult, err := h.service.Ping(r.Context())
	if err != nil {
		h.logger.Error(err, "[PLogin] Error service execution (Ping).")
		return
	}
	response, _ := PingEncodeResponse(r.Context(), &serviceResult)
	if err != nil {
		h.logger.Error(err, "[PLogin] Error function execution (PLoginEncodeResponse).")
		return
	}
	err = PingTransportResponse(w, response)
	if err != nil {
		h.logger.Error(err, "[PLogin] Error function execution (PLoginTransportResponse).")
		return
	}

	return
}

func PingDecodeRequest(ctx context.Context, r *http.Request) (request *model.Pong, err error)  {
	return request, err
}

func PingEncodeResponse(ctx context.Context, serviceResult *model.Pong) (response model.Pong, err error)  {
	return *serviceResult, err
}

func PingTransportResponse(w http.ResponseWriter, response interface{}) (err error)  {
	d, err := json.Marshal(response)

	w.Write(d)
	return err
}