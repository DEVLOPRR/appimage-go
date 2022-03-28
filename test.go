package main

import (
	"appimagego/src"
	"fmt"
	"strings"
)

func main() {
	// myAppImage, err := appimagego.NewAppImage("/home/adityam/Documents/appimage-go/LiteXL.AppImage")
	myAppImage, err := appimagego.NewAppImage("/home/adityam/Applications/bread-0.5.0-x86_64.AppImage")
	if err != nil {
		fmt.Println(err)
		return;
	}

	fmt.Println("Name:", myAppImage.Name)
	fmt.Println("Description:", myAppImage.Description)
	fmt.Println("Version:", myAppImage.Version)
	fmt.Println("Categories:", strings.Join(myAppImage.Categories, ", "))
	fmt.Println("Mime Type:", strings.Join(myAppImage.MimeType, ", "))
	updateInfo, err := myAppImage.GetUpdateInformation();
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Println("Update Info:", updateInfo)
	fmt.Println("AppImage Type:", myAppImage.Type())
	fmt.Println("Should Be Integrated?", myAppImage.ShallBeIntegrated())
}