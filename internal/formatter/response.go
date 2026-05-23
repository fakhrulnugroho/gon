package formatter

import (
	"fmt"
	"gon/internal/color"
	"gon/internal/httpclient"
	"strconv"
)

func HttpCall(response *httpclient.Response, output string) {
	fmt.Print("\n")
	if response.StatusCode >= 500 {
		fmt.Println(color.Danger(strconv.Itoa(response.StatusCode)))
	} else if response.StatusCode >= 400 {
		fmt.Println(color.Warning(strconv.Itoa(response.StatusCode)))
	} else if response.StatusCode >= 300 {
		fmt.Println(color.Info(strconv.Itoa(response.StatusCode)))
	} else {
		fmt.Println(color.Success(strconv.Itoa(response.StatusCode)))
	}
	fmt.Print("\n")
	fmt.Println(PrettyJSON(response.BodyBytes))
	fmt.Print("\n\n")
}
