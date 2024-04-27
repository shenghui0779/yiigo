package curl

import (
	"errors"
	"io"
	"mime/multipart"
	"net/http"

	"github.com/shenghui0779/yiigo/value"
)

type options struct {
	header http.Header
	cookie []*http.Cookie
	close  bool
}

// Option HTTP请求选项
type Option func(o *options)

// WithHeader 设置HTTP请求头
func WithHeader(key string, vals ...string) Option {
	return func(o *options) {
		if len(vals) == 1 {
			o.header.Set(key, vals[0])
			return
		}
		for _, v := range vals {
			o.header.Add(key, v)
		}
	}
}

// WithCookies 设置HTTP请求Cookie
func WithCookies(cookies ...*http.Cookie) Option {
	return func(o *options) {
		o.cookie = cookies
	}
}

// WithClose 请求结束后关闭请求
func WithClose() Option {
	return func(o *options) {
		o.close = true
	}
}

// UploadForm HTTP文件上传表单
type UploadForm interface {
	// Field 返回表单普通字段
	Field(name string) string
	// Write 将表单文件写入流
	Write(w *multipart.Writer) error
}

// FormFileFunc 将文件写入表单流
type FormFileFunc func(w io.Writer) error

type formfile struct {
	fieldname string
	filename  string
	filefunc  FormFileFunc
}

type uploadform struct {
	files  []*formfile
	fields value.V
}

func (form *uploadform) Field(name string) string {
	return form.fields.Get(name)
}

func (form *uploadform) Write(w *multipart.Writer) error {
	if len(form.files) == 0 {
		return errors.New("empty file field")
	}

	for _, v := range form.files {
		part, err := w.CreateFormFile(v.fieldname, v.filename)
		if err != nil {
			return err
		}
		if err = v.filefunc(part); err != nil {
			return err
		}
	}
	for name, value := range form.fields {
		if err := w.WriteField(name, value); err != nil {
			return err
		}
	}
	return nil
}

// UploadField 文件上传表单字段
type UploadField func(form *uploadform)

// WithFormFile 设置表单文件字段
func WithFormFile(fieldname, filename string, fn FormFileFunc) UploadField {
	return func(form *uploadform) {
		form.files = append(form.files, &formfile{
			fieldname: fieldname,
			filename:  filename,
			filefunc:  fn,
		})
	}
}

// WithFormField 设置表单普通字段
func WithFormField(name, value string) UploadField {
	return func(form *uploadform) {
		form.fields.Set(name, value)
	}
}

// NewUploadForm 生成一个文件上传表单
func NewUploadForm(fields ...UploadField) UploadForm {
	form := &uploadform{
		files:  make([]*formfile, 0),
		fields: make(value.V),
	}

	for _, fn := range fields {
		fn(form)
	}

	return form
}
