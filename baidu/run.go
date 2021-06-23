package baidu

// 百度登录
func LoginBD() {
	ats := NewBD(false)
	_ = ats.Loginx()
}

// 百度视频上传
func UploadBD(files string) {
	bd := NewBD(false)
	bd.UploadFiles(files)
}
