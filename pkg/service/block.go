package service

import (
	"context"
	"encoding/json"
	"github.com/buildboxapp/app/pkg/model"
	"github.com/gorilla/mux"
	"os"
	"strconv"
	"strings"
)


// Ping ...
func (s *service) Block(ctx context.Context, uid string) (result []model.Pong, err error) {
	var objBlock model.ResponseData
	dataPage 		:= model.Data{} // пустое значение, используется в блоке для кеширования если он вызывается из страницы

	s.utils.Curl("GET", "_objs/"+uid, "", &objBlock)

	moduleResult := c.ModuleBuild(objBlock.Data[0], r, dataPage, nil, false)

	w.Write([]byte(moduleResult.result))

	return r, err
}