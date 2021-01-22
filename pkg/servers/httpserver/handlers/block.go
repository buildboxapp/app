package handlers

import (
	"context"
	"encoding/json"
	"github.com/buildboxapp/app/pkg/model"
	"net/http"
)


// Block get user by login+pass pair
// @Summary get user by login+pass pair
// @Param login_input body model.Pong true "login data"
// @Success 200 {object} model.Pong [Result:model.Pong]
// @Failure 400 {object} model.Pong
// @Failure 500 {object} model.Pong
// @Router /api/v1/block [get]
func (h *handlers) Block(w http.ResponseWriter, r *http.Request) {
	request, err := BlockDecodeRequest(r.Context(), r)
	if err != nil {
		h.logger.Error(err, "[Block] Error function execution (BlockDecodeRequest).")
		return
	}
	serviceResult, err := h.service.Block(r.Context(), request)
	if err != nil {
		h.logger.Error(err, "[Block] Error service execution (Block).")
		return
	}
	response, _ := BlockEncodeResponse(r.Context(), &serviceResult)
	if err != nil {
		h.logger.Error(err, "[Block] Error function execution (BlockEncodeResponse).")
		return
	}
	err = BlockTransportResponse(w, response)
	if err != nil {
		h.logger.Error(err, "[Block] Error function execution (BlockTransportResponse).")
		return
	}

	return
}

func BlockDecodeRequest(ctx context.Context, r *http.Request) (request model.BlockRequest, err error)  {
	return request, err
}

func BlockEncodeResponse(ctx context.Context, serviceResult *[]model.Pong) (response []model.Pong, err error)  {
	return *serviceResult, err
}

func BlockTransportResponse(w http.ResponseWriter, response interface{}) (err error)  {
	d, err := json.Marshal(response)

	w.Write(d)
	return err
}