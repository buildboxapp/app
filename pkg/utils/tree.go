// функции для вложение объектов Data в формат дерева DataTree

package utils

import (
	"encoding/json"
	"github.com/buildboxapp/app/pkg/model"
	"sort"
	"strconv"
)

type DataTree struct {
	Uid        		string               `json:"uid"`
	Id         		string               `json:"id"`
	Source     		string               `json:"source"`
	Parent     		string               `json:"parent"`
	Type       		string               `json:"type"`
	Title      		string               `json:"title"`
	Rev        		string               `json:"rev"`
	Сopies			string 				 `json:"copies"`
	Attributes 		map[string]model.Attribute `json:"attributes"`
	Sub    			[]string 			 `json:"sub"`
	Incl   			[]*DataTree 		 `json:"incl"`
}


////////////////////////////////////////////////////////////////////////////////////////
///////////////  /////////////////////
////////////////////////////////////////////////////////////////////////////////////////
// формируем вложенную структуру объектов
func (u *utils) DataToIncl(objData []model.Data) []*model.DataTree {

	// переводим slice в map, чтобы можно было удалять объект и обращаться по ключу при формировании подуровней навигатора
	mapLevel := map[string]*DataTree{}
	for _, v := range objData {
		item := DataTree{}

		item.Uid = v.Uid
		item.Source = v.Source
		item.Type = v.Type
		item.Attributes = v.Attributes
		item.Title = v.Title
		item.Type = v.Type
		item.Parent = v.Parent
		item.Rev = v.Rev
		item.Сopies = v.Сopies

		mapLevel[v.Uid] = &item
	}

	// делаю обратное наследование, добавляю в Sub значения всех потомков (для оптимальной функции вложения)
	for _, v := range mapLevel {
		if _, found := v.Attributes["leader"]; found {
			Leader := v.Attributes["leader"].Src
			if Leader != "" && v.Uid != "" {
				d, f := mapLevel[Leader]
				if f {
					d.Sub = append(d.Sub, v.Uid)
				}
			}
		}

	}

	// пробегаем карту полигонов и переносим вложенные внутрь
	for _, item := range mapLevel {
		item.scanSub(&mapLevel)
	}

	// преобразуем карту в слайс
	sliceNavigator := []*model.DataTree{}
	for _, m := range mapLevel {
		d := model.DataTree{}
		n, _ := json.Marshal(m)
		json.Unmarshal(n, d)
		sliceNavigator = append(sliceNavigator, &d)
	}

	// сортируем по order как число
	u.SortItems(sliceNavigator, "order", "int")


	return sliceNavigator
}

// метод типа Items (перемещаем структуры в карте, исходя из заявленной вложенности элементов)
// (переделать дубль фукнции)
func (p *DataTree) scanSub(maps *map[string]*DataTree) {
	if p.Sub != nil && len(p.Sub) != 0 {
		for _, c := range p.Sub {
			gg := *maps
			fromP := gg[c]
			if fromP != nil {
				copyPolygon := *fromP
				p.Incl = append(p.Incl, &copyPolygon)
				delete(*maps, c)
				copyPolygon.scanSub(maps)
			}
		}
	}
}

// сортируем в слейсе полигонов по полю sort
// typesort - тип сортировки (string/int) - если int то преобразуем в число перед сортировкой
// fieldsort - поле для сортировки
func (u *utils) SortItems(pd []*model.DataTree, fieldsort string, typesort string) {

	sort.Slice(u, func(i, j int) bool {

		value1 := "0"
		value2 := "0"
		if typesort == "int" {
			value1 = "0"
			value2 = "0"
		}


		if oi, found := pd[i].Attributes[fieldsort]; found {
			if oi.Value != "" {
				value1 = oi.Value
			}
		}
		if oj, found := pd[j].Attributes[fieldsort]; found {
			if oj.Value != "" {
				value2 = oj.Value
			}
		}

		vi1, err1 := strconv.Atoi(value1)
		vi2, err2 := strconv.Atoi(value2)

		// если передан int, но произошла ошибка, то не не меняем
		if typesort == "int" {
			if err1 == nil && err2 == nil {
				return vi1 < vi2
			} else {
				return false
			}
		} else {
			// если стринг, то всегда проверяем как-будто это сравнение строк
			return vi1 < vi2
		}


	})

	for i, _ := range pd {
		if pd[i].Incl != nil && len(pd[i].Incl) != 0 {
			f := pd[i].Incl
			u.SortItems(f, fieldsort, typesort)
		}
	}
}

// вспомогательная фукнция выбирает только часть дерево от заданного лидера
func (u *utils) TreeShowIncl(in []*model.DataTree, obj string) (out []*model.DataTree) {
	if obj == "" {
		return in
	}

	for _, v := range in {

		if v.Source == obj {
			out = append(out, v)
			return out
		} else {

			out = u.TreeShowIncl(v.Incl, obj)
			if len(out) != 0 {
				return out
			}
		}

	}
	return out
}

////////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////

