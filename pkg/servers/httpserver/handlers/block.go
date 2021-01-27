package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/buildboxapp/app/pkg/model"
	"github.com/gorilla/mux"
	"html/template"
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
	in, err := blockDecodeRequest(r.Context(), r)
	if err != nil {
		h.logger.Error(err, "[Block] Error function execution (BlockDecodeRequest).")
		return
	}
	serviceResult, err := h.service.Block(r.Context(), in)
	if err != nil {
		h.logger.Error(err, "[Block] Error service execution (Block).")
		return
	}
	response, _ := blockEncodeResponse(r.Context(), &serviceResult)
	if err != nil {
		h.logger.Error(err, "[Block] Error function execution (BlockEncodeResponse).")
		return
	}
	err = blockTransportResponse(w, response)
	if err != nil {
		h.logger.Error(err, "[Block] Error function execution (BlockTransportResponse).")
		return
	}

	return
}

func blockDecodeRequest(ctx context.Context, r *http.Request) (in model.ServiceBlockIn, err error)  {
	vars := mux.Vars(r)
	in.Block = vars["block"]
	in.Url = r.URL.Query().Encode()
	in.Referer = r.Referer()
	in.RequestURI = r.RequestURI

	// указатель на профиль текущего пользователя
	var profile model.ProfileData
	profileRaw := r.Context().Value("UserRaw")
	json.Unmarshal([]byte(fmt.Sprint(profileRaw)), &profile)

	in.Profile = profile

	return in, err
}

func blockEncodeResponse(ctx context.Context, serviceResult *model.ServiceBlockOut) (response template.HTML, err error)  {
	response = serviceResult.Result
	return response, err
}

func blockTransportResponse(w http.ResponseWriter, response interface{}) (err error)  {
	d, err := json.Marshal(response)
	w.WriteHeader(200)

	if err != nil {
		w.WriteHeader(403)
	}
	w.Write(d)
	return err
}