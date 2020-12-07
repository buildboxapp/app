package main

import (
	"fmt"
	"net/smtp"
	"os"
)

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
	// user we are authorizing as
	from := "loveckiy@gmail.com"

	// use we are sending email to
	to := "loveckiy@gmail.com"

	// server we are authorized to send email through
	host := "smtp.gmail.com:587"
	hostUser := "loveckiy@gmail.com"
	hostPass := "Nhbybnb7"

	// Create the authentication for the SendMail()
	// using PlainText, but other authentication methods are encouraged
	auth := smtp.PlainAuth("", hostUser, hostPass, host)

	// NOTE: Using the backtick here ` works like a heredoc, which is why all the
	// rest of the lines are forced to the beginning of the line, otherwise the
	// formatting is wrong for the RFC 822 style
	message := `To: "Some User" <someuser@example.com>
From: "Other User" <otheruser@example.com>
Subject: Testing Email From Go!!

This is the message we are sending. That's it!
`

	if err := smtp.SendMail(host, auth, from, []string{to}, []byte(message)); err != nil {
		fmt.Println("Error SendMail: ", err)
		os.Exit(1)
	}
	fmt.Println("Email Sent!")
}