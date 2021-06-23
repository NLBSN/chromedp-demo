package tengxun

// 腾讯登录
func LoginTX() {
	tx := NewTX(false)
	_ = tx.Loginx()
}

// 腾讯视频上传
func UploadTX(files string) {
	tx := NewTX(false)
	tx.UploadFiles(files)
}
