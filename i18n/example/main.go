// 运行文件夹和文件的情况需要进入到 example 文件夹
package main

import (
	"fmt"
	"github.com/LiangNing7/onex/pkg/i18n"
	"github.com/LiangNing7/onex/pkg/i18n/example/locales"
	"golang.org/x/text/language"
)

func main() {
	i := i18n.New()
	// 1. add dir
	//i.Add("./locales")

	// 2. add files
	//i.Add("./locales/en.yaml")
	//i.Add("./locales/zh.yaml")

	// 3. add embed fs
	i.AddFS(locales.Locales)

	fmt.Println(i.T("no.permission"))
	fmt.Println(i.Select(language.Chinese).T("no.permission"))
}
