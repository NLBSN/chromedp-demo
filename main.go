package main

import (
	"gitee/zzf/chromedp-demo/baidu"
	"gitee/zzf/chromedp-demo/common"
	"gitee/zzf/chromedp-demo/douyin"
	"gitee/zzf/chromedp-demo/tengxun"
	"os"
)

func main() {
	switch os.Args[1] {
	case "1":
		common.ListDir(os.Args[2])
	case "2":
		baidu.LoginBD()
	case "3":
		baidu.UploadBD(os.Args[2])
	case "4":
		douyin.LoginDY()
	case "5":
		douyin.UploadDY(os.Args[2])
	case "6":
		tengxun.LoginTX()
	case "7":
		tengxun.UploadTX(os.Args[2])
	}
}
