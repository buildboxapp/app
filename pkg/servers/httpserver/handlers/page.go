package handlers

import (
	"context"
	"encoding/json"
	"fmt"
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
	in, err := pageDecodeRequest(r.Context(), r)
	if err != nil {
		h.logger.Error(err, "[Page] Error function execution (PageDecodeRequest).")
		return
	}
	serviceResult, err := h.service.Page(r.Context(), in)
	if err != nil {
		h.logger.Error(err, "[Page] Error service execution (Page).")
		return
	}
	response, _ := pageEncodeResponse(r.Context(), &serviceResult)
	if err != nil {
		h.logger.Error(err, "[Page] Error function execution (PageEncodeResponse).")
		return
	}
	err = pageTransportResponse(w, response)
	if err != nil {
		h.logger.Error(err, "[Page] Error function execution (PageTransportResponse).")
		return
	}

	return
}

func pageDecodeRequest(ctx context.Context, r *http.Request) (in model.ServiceIn, err error)  {
	vars := mux.Vars(r)
	in.Page = vars["page"]

	in.Url = r.URL.Query().Encode()
	in.Referer = r.Referer()
	in.RequestURI = r.RequestURI
	in.Form = r.Form
	in.Host = r.Host
	in.Query = r.URL.Query()

	// указатель на профиль текущего пользователя
	var profile model.ProfileData
	profileRaw := r.Context().Value("UserRaw")
	json.Unmarshal([]byte(fmt.Sprint(profileRaw)), &profile)

	in.Profile = profile

	return in, err
}

func pageEncodeResponse(ctx context.Context, serviceResult *model.ServicePageOut) (response string, err error)  {
	response = serviceResult.Body
	return response, err
}

func pageTransportResponse(w http.ResponseWriter, response string) (err error)  {
	w.WriteHeader(200)

	if err != nil {
		w.WriteHeader(403)
	}
	w.Write([]byte(response))
	return err
}
