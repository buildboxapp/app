package main

//import (
//	"fmt"
//	buildboxapp "github.com/buildboxapp/app/lib"
//)
//
//var app1 = buildboxapp.App{}
//
//func main()  {
//	params := []string{
//		"2020-01-02 15:04:05",
//		" ",
//		"-",
//		"-1",
//	}
//
//	//fmt.Println(app1.Time([]string{"THIS",""}))
//	result := app1.DReplace(params)
//
//	//arg := "@Time(now), '+10h' "
//	//
//	//args := strings.Split(arg, ",")
//	//// очищаем каждый параметр от ' если есть
//	//argsClear := []string{}
//	//for _, v := range args{
//	//	v = strings.Trim(v, " ")
//	//	v = strings.Trim(v, "'")
//	//	argsClear = append(argsClear, v)
//	//}
//	//
//	//for _, v := range argsClear {
//	//	fmt.Println("-"+v+"-")
//	//}
//
//	fmt.Println(result)
//}

//func main() {
//	go gor1()
//	time.Sleep(2 * time.Second)
//}
//
//func gor1()  {
//	go gor2()
//	return
//}
//
//func gor2()  {
//	for {
//		fmt.Print(".")
//		time.Sleep(10 * time.Millisecond)
//	}
//	return
//}


func main1() {
	//// user we are authorizing as
	//from := "loveckiy@gmail.com"
	//
	//// use we are sending email to
	//to := "loveckiy@gmail.com"
	//
	//// server we are authorized to send email through
	////host := "smtp.gmail.com"
	////hostUser := "loveckiy@gmail.com"
	////hostPass := "Nhbybnb7"
	//
	////auth := smtp.PlainAuth("", hostUser, hostPass, host)
	//
	//mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
	//fromTitle := "From: \"Ловецкий Иван\" <"+from+">\n"
	//toTitle := "To: \"Петрову Ильдору\" <"+to+">\n"
	//subject := "Subject: Test email from Go!\n"
	//mes := "http://labs.lovetsky.ru/buildbox/gui/view/page/welcome"
	//
	//var result interface{}
	//if len(mes) > 5 {
	//	if mes[:4] == "http" {
	//		result, _ = bblib.Curl("GET", mes, "", result, map[string]string{})
	//	}
	//}
	//
	//fmt.Println(result)
	//
	//message := []byte(subject + fromTitle + toTitle + mime + mes)
	//fmt.Println(message)
	////if err := smtp.SendMail(host+":587", auth, from, []string{to}, []byte(message)); err != nil {
	////	fmt.Println("Error SendMail: ", err)
	////	os.Exit(1)
	////}
	//fmt.Println("Email Sent!")
}