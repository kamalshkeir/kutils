package kfile

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/kamalshkeir/klog"
	"github.com/kamalshkeir/kutils/kslice"
	"github.com/kamalshkeir/kutils/kstring"
)

func DeleteFile(path string) error {
	err := os.Remove("." + path)
	if err != nil {
		return err
	} else {
		return nil
	}
}

// Parse body multipart
func ParseMultipartForm(r *http.Request, size ...int64) (formData url.Values, formFiles map[string][]*multipart.FileHeader) {
	s := int64(32 << 20)
	if len(size) > 0 {
		s = size[0]
	}
	parseErr := r.ParseMultipartForm(s)
	if parseErr != nil {
		klog.Printf("rdParse error = %v\n",parseErr)
	}
	defer func() {
		err := r.MultipartForm.RemoveAll()
		klog.CheckError(err)
	}()
	formData = r.Form
	formFiles = r.MultipartForm.File
	return formData, formFiles
}

// UPLOAD Multipart FILE
func UploadMultipart(file multipart.File, filename string, outPath string, acceptedFormats ...string) (string, error) {
	//create destination file making sure the path is writeable.
	if outPath != "" {
		if !strings.HasSuffix(outPath, "/") {
			outPath += "/"
		}
	} else {
		outPath="./uploads/"
	}
	err := os.MkdirAll(outPath, 0770)
	if err != nil {
		return "", err
	}

	l := []string{"jpg", "jpeg", "png", "json"}
	if len(acceptedFormats) > 0 {
		l = acceptedFormats
	}

	if _,ok := kstring.Contains(filename, l...);ok {
		dst, err := os.Create(outPath + filename)
		if err != nil {
			return "", err
		}
		defer dst.Close()

		//copy the uploaded file to the destination file
		if _, err := io.Copy(dst, file); err != nil {
			return "", err
		} else {
			url := "/" + outPath + filename
			return url, nil
		}
	} else {
		return "", fmt.Errorf("not in allowed extensions 'jpg','jpeg','png','json' : %v", l)
	}
}

// UPLOAD FILE
func UploadBytes(fileData []byte, filename string, outPath string, acceptedFormats ...string) (string, error) {
	//create destination file making sure the path is writeable.
	if outPath == "" {
		outPath = "./uploads/"
	} else {
		if !strings.HasSuffix(outPath, "/") {
			outPath += "/"
		}
	}
	err := os.MkdirAll(outPath, 0770)
	if err != nil {
		return "", err
	}

	l := []string{"jpg", "jpeg", "png", "json"}
	if len(acceptedFormats) > 0 {
		l = acceptedFormats
	}

	if _,ok := kstring.Contains(filename, l...);ok {
		dst, err := os.Create(outPath + filename)
		if err != nil {
			return "", err
		}
		defer dst.Close()

		//copy the uploaded file to the destination file
		if _, err := io.Copy(dst, bytes.NewReader(fileData)); err != nil {
			return "", err
		} else {
			url := "/" + outPath + filename
			return url, nil
		}
	} else {
		return "", fmt.Errorf("not in allowed extensions 'jpg','jpeg','png','json' : %v", l)
	}
}

func Upload(received_filename, folder_out string, r *http.Request, acceptedFormats ...string) (string, []byte, error) {
	r.ParseMultipartForm(10 << 20) //10Mb
	defer func() {
		err := r.MultipartForm.RemoveAll()
		klog.CheckError(err)
	}()
	var buff bytes.Buffer
	file, header, err := r.FormFile(received_filename)
	if klog.CheckError(err) {
		return "", nil, err
	}
	defer file.Close()
	// copy the uploaded file to the buffer
	if _, err := io.Copy(&buff, file); err != nil {
		return "", nil, err
	}

	data_string := buff.String()

	// make DIRS if not exist
	err = os.MkdirAll(folder_out+"/", 0664)
	if err != nil {
		return "", nil, err
	}
	// make file
	if len(acceptedFormats) == 0 {
		acceptedFormats = []string{"jpg", "jpeg", "png", "json"}
	}
	if _,ok := kstring.Contains(header.Filename, acceptedFormats...);ok {
		dst, err := os.Create(folder_out + "/" + header.Filename)
		if err != nil {
			return "", nil, err
		}
		defer dst.Close()
		dst.Write([]byte(data_string))

		url := folder_out + "/" + header.Filename
		return url, []byte(data_string), nil
	} else {
		return "", nil, fmt.Errorf("expecting filename to finish to be %v", acceptedFormats)
	}
}

func UploadMany(received_filenames []string, folder_out string, r *http.Request, acceptedFormats ...string) ([]string, [][]byte, error) {
	_, formFiles := ParseMultipartForm(r)
	urls := []string{}
	datas := [][]byte{}
	for inputName, files := range formFiles {
		var buff bytes.Buffer
		_,okContainSlice := kslice.Contains(received_filenames, inputName)
		if len(files) > 0 && okContainSlice {
			for _, f := range files {
				file, err := f.Open()
				if klog.CheckError(err) {
					return nil, nil, err
				}
				defer file.Close()
				// copy the uploaded file to the buffer
				if _, err := io.Copy(&buff, file); err != nil {
					return nil, nil, err
				}

				data_string := buff.String()

				// make DIRS if not exist
				err = os.MkdirAll(folder_out+"/", 0664)
				if err != nil {
					return nil, nil, err
				}
				// make file
				if len(acceptedFormats) == 0 {
					acceptedFormats = []string{"jpg", "jpeg", "png", "json"}
				}
				if _,ok := kstring.Contains(f.Filename, acceptedFormats...);ok {
					dst, err := os.Create(folder_out + "/" + f.Filename)
					if err != nil {
						return nil, nil, err
					}
					defer dst.Close()
					dst.Write([]byte(data_string))

					url := folder_out + "/" + f.Filename
					urls = append(urls, url)
					datas = append(datas, []byte(data_string))
				} else {
					klog.Printf("%s not handled\n",f.Filename)
					return nil, nil, fmt.Errorf("expecting filename to finish to be %v", acceptedFormats)
				}
			}
		}
	}
	return urls, datas, nil
}

func CopyDir(source, destination string) error {
	var err error = filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		var relPath string = strings.Replace(path, source, "", 1)
		if relPath == "" {
			return nil
		}
		if info.IsDir() {
			return os.Mkdir(filepath.Join(destination, relPath), 0755)
		} else {
			var data, err1 = os.ReadFile(filepath.Join(source, relPath))
			if err1 != nil {
				return err1
			}
			return os.WriteFile(filepath.Join(destination, relPath), data, 0777)
		}
	})
	return err
}


func PathExists(path string) bool {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}