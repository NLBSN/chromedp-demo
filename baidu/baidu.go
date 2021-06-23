package baidu

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"gitee/zzf/chromedp-demo/common"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"github.com/chromedp/chromedp/kb"
	"github.com/unknwon/com"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func (s *BDStruct) WriteInfo(str string) {
	file, err := os.OpenFile(s.CsvFile, os.O_CREATE|os.O_RDWR|os.O_APPEND, fs.ModePerm)
	defer func() {
		if file != nil {
			_ = file.Close()
		}
	}()
	writer := csv.NewWriter(file)
	// https://bj.bcebos.com/v1/ai-studio-online/7cd09ae961a54678ad4b784642192b40ca657f0368724552934384571&ab3234? responseContentDisposition=attachment%3B%20filename%3D文件名.mp4
	err = writer.Write([]string{filepath.Base(s.File), str + "?" + "responseContentDisposition=attachment%3B%20filename%3D" + filepath.Base(s.File)})
	if err != nil {
		fmt.Println("error write csv baidu ...", common.Enc.ConvertString(filepath.Base(s.File)), common.Enc.ConvertString(s.Url))
		return
	}
	writer.Flush()
}

func (s *BDStruct) Task() {
	b := make(chan bool, 1)
	//fmt.Println("开始监听...")
	downloadComplete := make(chan bool, 1)
	var requestId network.RequestID
	chromedp.ListenTarget(s.Ctx, func(v interface{}) {
		switch ev := v.(type) {
		case *network.EventRequestWillBeSent: // 开始发送
			if ev.Request.URL == "https://aistudio.baidu.com/studio/file/addfile" {
				//if strings.Contains(ev.Request.URL, "uploads=") {
				//fmt.Println("开始发送...文件...")
				requestId = ev.RequestID
			}
		case *network.EventResponseReceived: // 返回的资源
			//resp := ev.Response
			//if len(resp.Headers) != 0 {
			//	log.Printf("received headers: %s", resp.Headers)
			//}
		case *network.EventLoadingFinished: // 资源加载完毕
			if ev.RequestID == requestId {
				//fmt.Println("结束发送...文件...")
				//fmt.Println("video sent successfully:", s.File)
				downloadComplete <- true
				//close(downloadComplete)
			}
		}
	})
	var now time.Time
	go func() {
		// 加载cookie
		//fmt.Println("加载cookie...")
		// 先做定时3秒
		// 如果成功，进行成功的值传递
		// 没有看到成功的值传递，就是失败

		fmt.Println("info cookie load ...")
		_ = chromedp.Run(s.Ctx, s.LoadCookies())
		// 打开网页
		//fmt.Println("模拟打开网页...")
		fmt.Println("info open url ...", common.Enc.ConvertString(s.Url))
		_ = chromedp.Run(s.Ctx, chromedp.Navigate(s.Url))
		// 点击创建数据集按钮
		_ = chromedp.Run(s.Ctx, chromedp.Click(
			`document.querySelector("#main > div.content-main > div > div.section__3Uic0 > div > div.ant-tabs-nav > div.ant-tabs-extra-content > button > span")`,
			chromedp.ByJSPath))
		now = time.Now()
		// 上传文件
		_ = chromedp.Run(s.Ctx, chromedp.SetUploadFiles(
			`document.querySelector("body > div:nth-child(6) > div > div.ant-modal-wrap.ai-common-modal.vertical-center-modal.dataset-modal > div > div.ant-modal-content > div.ant-modal-body > div > div:nth-child(2) > div.create-dataset-item-value > div.create-dataset-item-value-height > input[type=file]")`,
			//[]string{"D:\\迅雷下载\\WeChat_20210526184543.mp4"},
			[]string{s.File},
			chromedp.ByJSPath,
		))
		b <- true
	}()
	select {
	case <-time.After(time.Second * 15):
		s.PrintErr("open chrom timeout...")
		return
	case <-b:
		fmt.Println("info video upload:", common.Enc.ConvertString(s.File))
	}

	fmt.Println("info waiting for video upload ...")
	fsize, _ := com.FileSize(s.File)
	//ft := (fsize / 1024.0) / (1024 * 1)
	ft := (fsize / 1024.0) / (1024 * 0.5)                                                          // 大概的传送时间，必须要大于这个  按照0.5m的带宽计算
	fmt.Println(common.Enc.ConvertString("按照512kb的速度，预计上传时间:"), ft, common.Enc.ConvertString("秒")) // 大概的传送时间，必须要大于这个  默认是按照1m的带宽计算

	select {
	case <-time.After(time.Second*30 + time.Second*time.Duration(ft)):
		s.PrintErr("error video sent fail...")
		return
	case <-downloadComplete:
		fmt.Println("info video sent successfully:", common.Enc.ConvertString(s.File))
	}
	fmt.Println("info video upload time:", time.Now().Sub(now).Seconds(), "s")
	//fmt.Println("文件上传结果:", <-downloadComplete)
	var downloadBytes []byte
	_ = chromedp.Run(s.Ctx, chromedp.ActionFunc(func(ctx context.Context) error {
		var err error
		downloadBytes, err = network.GetResponseBody(requestId).Do(ctx)
		return err
	}))
	var t AistudioDataStruct
	err := json.Unmarshal(downloadBytes, &t)
	if err != nil {
		s.PrintErr(err.Error())
		return
	}
	//fmt.Println("返回的结果状态:", t.ErrorMsg)
	if t.ErrorMsg != "成功" {
		s.PrintErr("baidu result ...")
		return
	}
	splitN := strings.SplitN(t.Result.BosURL, "?", 2)
	fmt.Println("info results obtained:", common.Enc.ConvertString(splitN[0]))
	// 数据集名称
	//fmt.Println("修改数据集名称...")
	fmt.Println("info fill in the video description ...")
	baseFile := filepath.Base(s.File)
	if len(baseFile) > 38 {
		baseFile = baseFile[:38]
	}

	go func() {
		_ = chromedp.Run(s.Ctx, chromedp.SendKeys(`document.querySelector("body > div:nth-child(6) > div > div.ant-modal-wrap.ai-common-modal.vertical-center-modal.dataset-modal > div > div.ant-modal-content > div.ant-modal-body > div > div:nth-child(1) > div.create-dataset-item-value > input")`,
			baseFile, chromedp.ByJSPath,
		))
		// 标签
		//fmt.Println("填写标签...")
		_ = chromedp.Run(s.Ctx, chromedp.SendKeys(`document.querySelector("body > div:nth-child(6) > div > div.ant-modal-wrap.ai-common-modal.vertical-center-modal.dataset-modal > div > div.ant-modal-content > div.ant-modal-body > div > div:nth-child(3) > div.create-dataset-item-value > div > div > div > input")`,
			time.Now().String(),
			chromedp.ByJSPath,
		))
		// 确认标签
		//fmt.Println("确认标签信息...")
		_ = chromedp.Run(s.Ctx, chromedp.SendKeys(`document.querySelector("body > div:nth-child(6) > div > div.ant-modal-wrap.ai-common-modal.vertical-center-modal.dataset-modal > div > div.ant-modal-content > div.ant-modal-body > div > div:nth-child(3) > div.create-dataset-item-value > div > div > div > input")`,
			kb.Enter,
			chromedp.ByJSPath,
		))
		// 简介摘要
		//fmt.Println("填写简介...")
		_ = chromedp.Run(s.Ctx, chromedp.SendKeys(
			`document.querySelector("body > div:nth-child(6) > div > div.ant-modal-wrap.ai-common-modal.vertical-center-modal.dataset-modal > div > div.ant-modal-content > div.ant-modal-body > div > div:nth-child(5) > div.create-dataset-item-value > textarea")`,
			time.Now().String(),
			chromedp.ByJSPath,
		))

		// 点击确定
		//fmt.Println("点击确定...")
		fmt.Println("info start publishing ...")
		_ = chromedp.Run(s.Ctx, chromedp.Click(
			`document.querySelector("body > div:nth-child(6) > div > div.ant-modal-wrap.ai-common-modal.vertical-center-modal.dataset-modal > div > div.ant-modal-content > div.ant-modal-footer > button.ant-btn.ant-btn-primary.ai-btn")`,
			chromedp.ByJSPath))
		// 等待出现某个页面
		//fmt.Println("等待是否查看的页面显示...")
		fmt.Println("info deal with post release issues ...")
		//time.Sleep(time.Minute)
		_ = chromedp.Run(s.Ctx, chromedp.WaitVisible(
			`document.querySelector("body > div:nth-child(7) > div > div.ant-modal-wrap.ai-confirm.ai-confirm-show-close.ant-modal-centered > div > div.ant-modal-content > div.ant-modal-body")`,
			chromedp.ByJSPath),
		)
		//fmt.Println("不进行查看确定...")
		// 点击取消按钮  --- 数据集已创建成功, 是否查看数据集详情?
		_ = chromedp.Run(s.Ctx, chromedp.Click(
			`document.querySelector("body > div:nth-child(7) > div > div.ant-modal-wrap.ai-confirm.ai-confirm-show-close.ant-modal-centered > div > div.ant-modal-content > div.ant-modal-footer > button:nth-child(2) > span")`,
			chromedp.ByJSPath))
		b <- true
	}()
	select {
	case <-b:
		fmt.Println("info file uploaded successfully ...", common.Enc.ConvertString(s.File))
		s.WriteInfo(splitN[0])
	case <-time.After(time.Second * 15):
		s.PrintErr("timeout")
		//s.Cancel()
		//os.Exit(1)
	}
}

func (s *BDStruct) UploadFiles(filestr string) {
	defer func() {
		fmt.Println("chrom exit")
		s.Cancel()
	}()
	files := strings.Split(filestr, ",")
	for _, file := range files {
		if file == "" {
			s.PrintErr("error file name, next file ...local filename")
			continue
		}
		if !com.IsExist(file) { // 文件不存在
			s.PrintErr("error file not exist, next file ... local filename")
			continue
		}
		s.File = file
		fmt.Println("info files ready to start uploading:", common.Enc.ConvertString(s.File))
		s.Task()
	}
}

func (s *BDStruct) SaveCookie() chromedp.ActionFunc {
	return func(ctx context.Context) (err error) {
		// 等待二维码登陆
		fmt.Println("info please log in to baidu ...")
		time.Sleep(time.Second * 2)
		fmt.Println("info please move the mouse to the user's position in the upper right corner of the page ...")
		if err = chromedp.WaitVisible(`#studio-logged-pop-wrap > a:nth-child(1)`, chromedp.ByID).Do(ctx); err != nil {
			fmt.Println("error", common.Enc.ConvertString(s.File), err)
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

func (s *BDStruct) Loginx() error {
	defer s.Cancel()
	fmt.Println("info login baidu ...")
	fmt.Println("info cookie remove ...")
	// 加载cookie
	if err := s.LoadCookiesExistRemove(); err != nil {
		return err
	}
	// 打开网页
	fmt.Println("info open chrom ...")
	if err := chromedp.Run(s.Ctx, chromedp.Navigate(s.Url)); err != nil {
		s.PrintErr("open chrom ...")
		return err
	}
	s.WaitTimes(60)
	if err := chromedp.Run(s.Ctx, s.SaveCookie()); err != nil {
		s.PrintErr("save cookie ...")
		return err
	}
	return nil
}

type BDStruct struct {
	common.AutoTranS
}

func NewBD(boo bool) common.AutoTranI {
	s := common.AutoTranS{
		Url:     "https://aistudio.baidu.com/aistudio/datasetoverview",
		Cookie:  "baidu.tmp",
		CsvFile: "baidu.csv",
	}
	s.Ctx, s.Cancel = s.LoginContext(boo)
	return &BDStruct{
		s,
	}
}

// 百度
/**
{
    "logId":"024bf846-8434-48ab-be06-6cd1507b50c3",
    "errorCode":0,
    "errorMsg":"成功",
    "timestamp":0,
    "result":{
        "fileId":461949,
        "fileName":"WeChat_20210526184543.mp4",
        "createUser":807446,
        "isPublic":1,
        "fileContentType":"video/mp4",
        "bosUrl":"http://bj.bcebos.com/v1/ai-studio-online/6ab1b4b6c75f4b02ad71b0189291ad461d90070159954eb19bc8213b5a630863?responseContentDisposition=attachment%3B%20filename%3DWeChat_20210526184543.mp4&authorization=bce-auth-v1%2F0ef6765c1e494918bc0d4c3ca3e5c6d1%2F2021-05-27T08%3A10%3A10Z%2F-1%2F%2Fb672dd256b74c1c36cf450f3aef94b9497bef42fe2b59496ab9f7ce1e643209c",
        "fileMd5":"",
        "fileKey":"6ab1b4b6c75f4b02ad71b0189291ad461d90070159954eb19bc8213b5a630863",
        "fileSize":1482545,
        "fileReview":"",
        "fileAbs":"",
        "datasetFilePath":"",
        "fileOriginName":"WeChat_20210526184543.mp4",
        "fileDownloadUrl":"",
        "clusterFilePath":""
    }
}
*/
type AistudioDataStruct struct {
	LogID     string             `json:"logId"`
	ErrorCode int                `json:"errorCode"`
	ErrorMsg  string             `json:"errorMsg"`
	Timestamp int                `json:"timestamp"`
	Result    AistudioDataResult `json:"result"`
}

type AistudioDataResult struct {
	FileID          int    `json:"fileId"`
	FileName        string `json:"fileName"`
	CreateUser      int    `json:"createUser"`
	IsPublic        int    `json:"isPublic"`
	FileContentType string `json:"fileContentType"`
	BosURL          string `json:"bosUrl"`
	FileMd5         string `json:"fileMd5"`
	FileKey         string `json:"fileKey"`
	FileSize        int    `json:"fileSize"`
	FileReview      string `json:"fileReview"`
	FileAbs         string `json:"fileAbs"`
	DatasetFilePath string `json:"datasetFilePath"`
	FileOriginName  string `json:"fileOriginName"`
	FileDownloadURL string `json:"fileDownloadUrl"`
	ClusterFilePath string `json:"clusterFilePath"`
}
