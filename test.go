package main

import (
	"fmt"
	buildboxapp "github.com/buildboxapp/app/lib"
)

var app1 = buildboxapp.App{}

func main()  {
	params := []string{
		"2020-01-02 15:04:05",
		" ",
		"-",
		"-1",
	}

	//fmt.Println(app1.Time([]string{"THIS",""}))
	result := app1.DReplace(params)

	//arg := "@Time(now), '+10h' "
	//
	//args := strings.Split(arg, ",")
	//// очищаем каждый параметр от ' если есть
	//argsClear := []string{}
	//for _, v := range args{
	//	v = strings.Trim(v, " ")
	//	v = strings.Trim(v, "'")
	//	argsClear = append(argsClear, v)
	//}
	//
	//for _, v := range argsClear {
	//	fmt.Println("-"+v+"-")
	//}

	fmt.Println(result)
}