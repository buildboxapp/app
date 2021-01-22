package handlers

import (
	"context"
	"encoding/json"
	"github.com/buildboxapp/app/pkg/model"
	"github.com/gorilla/mux"
	"net/http"
)

// Page get user by login+pass pair
// @Summary get user by login+pass pair
// @Param login_input body model.Pong true "login data"
// @Success 200 {object} model.Pong [Result:model.Pong]
// @Failure 400 {object} model.Pong
// @Failure 500 {object} model.Pong
// @Router /api/v1/page [get]
func (h *handlers) Page(w http.ResponseWriter, r *http.Request) {
	block, err := PageDecodeRequest(r.Context(), r)
	if err != nil {
		h.logger.Error(err, "[PLogin] Error function execution (PLoginDecodeRequest).")
		return
	}
	serviceResult, err := h.service.Page(r.Context(), block)
	if err != nil {
		h.logger.Error(err, "[PLogin] Error service execution (Page).")
		return
	}
	response, _ := PageEncodeResponse(r.Context(), &serviceResult)
	if err != nil {
		h.logger.Error(err, "[PLogin] Error function execution (PLoginEncodeResponse).")
		return
	}
	err = PageTransportResponse(w, response)
	if err != nil {
		h.logger.Error(err, "[PLogin] Error function execution (PLoginTransportResponse).")
		return
	}

	return
}

func PageDecodeRequest(ctx context.Context, r *http.Request) (block string, err error)  {
	vars := mux.Vars(r)
	block = vars["block"]

	return block, err
}

func PageEncodeResponse(ctx context.Context, serviceResult *[]model.Pong) (response []model.Pong, err error)  {
	return *serviceResult, err
}

func PageTransportResponse(w http.ResponseWriter, response interface{}) (err error)  {
	d, err := json.Marshal(response)

	w.Write(d)
	return err
}