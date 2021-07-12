package service

import (
	"context"
	"fmt"
	"github.com/buildboxapp/app/pkg/model"
)

func (s *service) Block(ctx context.Context, in model.ServiceIn) (out model.ServiceBlockOut, err error) {
	var objBlock model.ResponseData
	dataPage := model.Data{} // пустое значение, используется в блоке для кеширования если он вызывается из страницы

	objBlock, err = s.api.ObjGet(in.Block)
	//s.utils.Curl("GET", "_objs/"+in.Block, "", &objBlock, map[string]string{})

	if len(objBlock.Data) == 0 {
		return out, fmt.Errorf("%s", "Error. Lenght data from objBlock is 0.")
	}
	moduleResult, err := s.block.Generate(in, objBlock.Data[0], dataPage, nil)
	out.Result = moduleResult.Result

	return
}
