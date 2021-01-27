package service

import (
	"context"
	"github.com/buildboxapp/app/pkg/model"
)

func (s *service) Block(ctx context.Context, in model.ServiceBlockIn) (out model.ServiceBlockOut, err error) {
	var objBlock model.ResponseData
	dataPage := model.Data{} // пустое значение, используется в блоке для кеширования если он вызывается из страницы

	s.utils.Curl("GET", "_objs/"+in.Block, "", &objBlock)

	moduleResult := s.block.ModuleBuild(in, objBlock.Data[0], dataPage, nil, false)
	out.Result = moduleResult.Result

	return
}
