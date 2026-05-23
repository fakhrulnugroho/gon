package formatter

import (
	"fmt"
	"gon/internal/color"
	"gon/internal/httpclient"
	"strconv"
)

func HttpCall(response *httpclient.Response, output string) {
	fmt.Println(color.Success(strconv.Itoa(response.StatusCode)))
	fmt.Print("\n")
	fmt.Println(PrettyJSON(response.BodyBytes))
}
