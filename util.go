package gwf

import (
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
)

func fileExist(filename string) bool {
	if filename == "" {
		return false
	}
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return false
	}
	return true
}

func getDeployRootPath(evalSymlinks bool) (string, error) {
	p, err := os.Executable()
	if err != nil {
		return "", err
	}
	if evalSymlinks {
		p, err = filepath.EvalSymlinks(ex)
		if err != nil {
			return "", err
		}
	}
	return filepath.Dir(p), nil
}

// SaveUploadedFile用于存储上传文件到文件路径dst
func SaveUploadedFile(file *multipart.FileHeader, dst string) error {
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	io.Copy(out, src)
	return nil
}
