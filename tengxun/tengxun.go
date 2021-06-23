package tengxun

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"gitee/zzf/chromedp-demo/common"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"github.com/unknwon/com"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func (s *TXSturct) WriteInfo(str string) {
	file, err := os.OpenFile(s.CsvFile, os.O_CREATE|os.O_RDWR|os.O_APPEND, fs.ModePerm)
	defer func() {
		if file != nil {
			_ = file.Close()
		}
	}()
	writer := csv.NewWriter(file)
	// https://bj.bcebos.com/v1/ai-studio-online/7cd09ae961a54678ad4b784642192b40ca657f0368724552934384571&ab3234? responseContentDisposition=attachment%3B%20filename%3D文件名.mp4
	err = writer.Write([]string{filepath.Base(s.File), "https://xft.gzc.svp.tencent-cloud.com/" + str})
	if err != nil {
		fmt.Println("error write csv douyin ...", []string{filepath.Base(s.File), str})
		return
	}
	writer.Flush()
}
func (s *TXSturct) Task() {
	//fmt.Println("开始监听...")
	//fmt.Println("加载cookie...")
	if err := chromedp.Run(s.Ctx, s.LoadCookies()); err != nil {
		fmt.Println("error", common.Enc.ConvertString(s.File), err)
		return
	}
	// 打开网页
	//fmt.Println("打开网页...")
	fmt.Println("info open url ...", common.Enc.ConvertString(s.Url))
	if err := chromedp.Run(s.Ctx, chromedp.Navigate(s.Url)); err != nil {
		fmt.Println("error", common.Enc.ConvertString(s.File), err)
		return
	}
	//fmt.Println("获取cookie...")
	//if err := chromedp.Run(s.Ctx, saveCookies()); err != nil {
	//	_ = fmt.Errorf("错误:%v\n", err)
	//}
	// 加载cookies

	downloadComplete := make(chan bool, 1)
	var requestId network.RequestID
	var requestUrl string
	//requestId := make([]network.RequestID, 0)
	chromedp.ListenTarget(s.Ctx, func(v interface{}) {
		//marshal, _ := json.Marshal(v)
		switch ev := v.(type) {
		case *network.EventRequestWillBeSent: // 开始发送
			if strings.Contains(ev.Request.URL, "https://videotranspond.3g.qq.com/v2/upload/uploadpart?filename=") {
				if ev.Request.HasPostData {
					//fmt.Println("开始发送...文件...", string(marshal))
					requestId = ev.RequestID
					requestUrl = ev.Request.URL
				}
				//requestId = append(requestId, ev.RequestID)
			}
		case *network.EventResponseReceived: // 返回的资源
			//resp := ev.Response
			//if len(resp.Headers) != 0 {
			//	log.Printf("received headers: %s", resp.Headers)
			//}
			//fmt.Println("发送中...文件...", string(marshal))
		case *network.EventLoadingFinished: // 资源加载完毕
			if ev.RequestID == requestId {
				//fmt.Println("结束发送...文件...", string(marshal))
				downloadComplete <- true
				//close(downloadComplete)
			}
		}
	})

	// 上传视频
	//fmt.Println("上传视频...")
	fmt.Println("info video upload:", common.Enc.ConvertString(s.File))
	now := time.Now()
	if err := chromedp.Run(s.Ctx, chromedp.Click(`document.querySelector("#uploadDefault > div")`, chromedp.ByJSPath),
		chromedp.SetUploadFiles(`document.querySelector("#videoUploadBtnSingle")`, []string{s.File}, chromedp.ByJSPath),
	); err != nil {
		fmt.Println("error", common.Enc.ConvertString(s.File), err)
		return
	}
	// 等待视频传输完成
	fmt.Println("info waiting for video upload ...")
	//if err := chromedp.Run(s.Ctx, chromedp.WaitVisible(`document.querySelector("#app > div > div.viewpage > div.listBody > div.videoEditor > div.videoInfo > div.itemControl > div > div.coverList > div.item > div > span.success")`,
	//	chromedp.ByJSPath,
	//)); err != nil {
	//	fmt.Println("error",s.File, err)
	//	return
	//}
	// 等待文件上传完成
	if <-downloadComplete {
		fmt.Println("info video sent successfully:", common.Enc.ConvertString(s.File))
	} else {
		fmt.Println("error video sent fail:", common.Enc.ConvertString(s.File))
		return
	}
	fmt.Println("info video upload time:", time.Now().Sub(now).Seconds(), "s")
	var downloadBytes []byte
	if err := chromedp.Run(s.Ctx, chromedp.ActionFunc(func(ctx context.Context) error {
		var err error
		//for _, v := range requestId {
		//	d, _ := network.GetResponseBody(v).Do(ctx)
		//	fmt.Println(v.String(), string(d))
		//downloadBytes = append(downloadBytes, d)
		//}
		downloadBytes, err = network.GetResponseBody(requestId).Do(ctx)
		return err
	})); err != nil {
		fmt.Println("error1", err)
		return
	}
	//fmt.Println("dow  ", string(downloadBytes))
	var t TengxunDataStruct
	err := json.Unmarshal(downloadBytes, &t)
	if err != nil {
		fmt.Println("error", common.Enc.ConvertString(s.File), err)
		return
	}
	if t.Msg != "ok" {
		fmt.Println("error tengxun result ...")
		return
	}
	splitNUrl := strings.SplitN(strings.SplitN(requestUrl, "filename=", 2)[1], "&", 2)
	fmt.Println("info results obtained:", common.Enc.ConvertString(splitNUrl[0]))

	// 写备注
	//fmt.Println("写备注...")
	fmt.Println("info fill in the video description ...")
	if err = chromedp.Run(s.Ctx, chromedp.SendKeys(
		`document.querySelector("#app > div > div.viewpage > div.listBody > div.videoEditor > div.videoInfo > div.wrap3 > div > div > div.titleEditor > div > div > div > div > div.ql-editor")`,
		filepath.Base(s.File),
		chromedp.ByJSPath,
	)); err != nil {
		fmt.Println("error", common.Enc.ConvertString(s.File), err)
		return
	}
	// 修改封页
	//time.Sleep(time.Second*2)
	getwd, _ := os.Getwd()
	if err = chromedp.Run(s.Ctx,
		chromedp.Click(`document.querySelector("#app > div > div.viewpage > div.listBody > div.videoEditor > div.videoInfo > div.itemControl > div > div.img.undefined.s916.videoCover > label")`, chromedp.ByJSPath),
		chromedp.SetUploadFiles(`document.querySelector("#imgUpload")`, []string{filepath.Join(getwd, "fengye.jpg")}, chromedp.ByJSPath),
	); err != nil {
		fmt.Println("error fengye", err)
		return
	}
	// 等待封页出来
	//fmt.Println("info wait for the video cover to come out ...")
	//if err := chromedp.Run(s.Ctx, chromedp.WaitVisible(`document.querySelector("#app > div > div.viewpage > div.listBody > div.videoEditor > div.videoInfo > div.itemControl > div > div.img.undefined.s916.videoCover.portrait > div > img")`,
	//	chromedp.ByJSPath,
	//)); err != nil {
	//	fmt.Println("error",s.File, err)
	//	return
	//}
	var cvalue string // button publishReady
	for {
		if err := chromedp.Run(s.Ctx, chromedp.AttributeValue(`document.querySelector("#pubtool > button")`, "class", &cvalue, nil, chromedp.ByJSPath)); err != nil {
			fmt.Println("error", common.Enc.ConvertString(s.File), err)
			return
		}
		if cvalue == "button publishReady" {
			//fmt.Println("info video sent successfully:", s.File)
			break
		} /*else{
			fmt.Println("------",time.Now().String())
		}*/
		time.Sleep(time.Millisecond * 100)
	}

	//fmt.Println("======",cvalue)
	// 发布
	fmt.Println("info start publishing ...")
	if err := chromedp.Run(s.Ctx, chromedp.Click(`document.querySelector("#pubtool > button")`,
		chromedp.ByJSPath,
	)); err != nil {
		fmt.Println("error", common.Enc.ConvertString(s.File), err)
		return
	}
	fmt.Println("info file uploaded successfully ...", common.Enc.ConvertString(s.File))
	if err := chromedp.Run(s.Ctx, chromedp.Click(`document.querySelector("#app > div > div.viewpage > div.listBody > div.publishSuccess > div > div")`,
		chromedp.ByJSPath,
	)); err != nil {
		fmt.Println("error", common.Enc.ConvertString(s.File), err)
		return
	}
	s.WriteInfo(splitNUrl[0])
}

func (s *TXSturct) UploadFiles(filestr string) {
	defer s.Cancel()
	files := strings.Split(filestr, ",")
	for _, file := range files {
		if file == "" {
			fmt.Println("error file name, next file ...local filename:", common.Enc.ConvertString(file))
			continue
		}
		if !com.IsExist(file) { // 文件不存在
			fmt.Println("error file not exist, next file ... local filename:", common.Enc.ConvertString(file))
			continue
		}
		s.File = file
		fmt.Println("info files ready to start uploading:", common.Enc.ConvertString(s.File))
		s.Task()
	}
}

func (s *TXSturct) SaveCookie() chromedp.ActionFunc {
	return func(ctx context.Context) (err error) {
		// 等待二维码登陆
		fmt.Println("info please log in to tengxun ...")
		time.Sleep(time.Second * 2)
		fmt.Println("info please move the mouse to the user's position in the upper right corner of the page ...")
		if err = chromedp.WaitVisible(`document.querySelector("#mod_head_notice_pop > div > div.quick_pop_user_hd > span")`, chromedp.ByJSPath).Do(ctx); err != nil {
			fmt.Println("err...", err)
			return err
		} // #main > header > div > div.studio-logged-wrapper > div > div > span
		fmt.Println("info login successful, save cookie ...")
		// cookies的获取对应是在devTools的network面板中
		// 1. 获取cookies
		cookies, err := network.GetAllCookies().Do(ctx)
		if err != nil {
			return err
		}

		// 2. 序列化
		cookiesData, err := network.GetAllCookiesReturns{Cookies: cookies}.MarshalJSON()
		if err != nil {
			return err
		}

		// 3. 存储到临时文件
		if err = ioutil.WriteFile(s.Cookie, cookiesData, 0755); err != nil {
			return err
		}
		fmt.Println("info cookie saved successfully ...")
		return
	}
}

func (s *TXSturct) Loginx() error {
	defer s.Cancel()
	fmt.Println("info login tengxun ...")
	fmt.Println("info cookie remove ...")
	// 加载cookie
	if err := s.LoadCookiesExistRemove(); err != nil {
		return err
	}
	// 打开网页
	fmt.Println("info open chrom ...")
	if err := chromedp.Run(s.Ctx, chromedp.Navigate(s.Url)); err != nil {
		fmt.Println("error", common.Enc.ConvertString(s.File), err)
		return err
	}
	s.WaitTimes(60)
	if err := chromedp.Run(s.Ctx, s.SaveCookie()); err != nil {
		fmt.Println("error", common.Enc.ConvertString(s.File), err)
		return err
	}
	return nil
}

func NewTX(boo bool) common.AutoTranI {
	s := common.AutoTranS{
		Url:     "https://cm.cc.v.qq.com/upload",
		Cookie:  "tengxun.tmp",
		CsvFile: "tengxun.csv",
	}
	s.Ctx, s.Cancel = s.LoginContext(boo)
	return &TXSturct{
		s,
	}
}

type TXSturct struct {
	common.AutoTranS
}

//腾讯
/**
{
    "code":0,
    "msg":"ok",
    "partSha":"11f6b68e8149b1285877a454fd7338dd29f25c63"
}
*/
type TengxunDataStruct struct {
	Code    int    `json:"code"`
	Msg     string `json:"msg"`
	PartSha string `json:"partSha"`
}
