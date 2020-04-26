package yiigo

import (
	"bytes"
	"encoding/xml"
	"errors"
	"math"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/go-playground/locales/zh"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	zhcn "github.com/go-playground/validator/v10/translations/zh"
	"github.com/hashicorp/go-version"
	"go.uber.org/zap"
)

// AsDefault alias for "default"
const AsDefault = "default"

// X is a convenient alias for a map[string]interface{}.
type X map[string]interface{}

// CDATA XML CDATA section which is defined as blocks of text that are not parsed by the parser, but are otherwise recognized as markup.
type CDATA string

// MarshalXML encodes the receiver as zero or more XML elements.
func (c CDATA) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	return e.EncodeElement(struct {
		string `xml:",cdata"`
	}{string(c)}, start)
}

// Date format a local time/date and
// returns a string formatted according to the given format string using the given timestamp of int64.
// The default layout is: 2006-01-02 15:04:05.
func Date(timestamp int64, layout ...string) string {
	l := "2006-01-02 15:04:05"

	if len(layout) > 0 {
		l = layout[0]
	}

	date := time.Unix(timestamp, 0).Format(l)

	return date
}

// WeekAround returns the date of monday and sunday for current week
func WeekAround() (monday, sunday string) {
	now := time.Now()
	offset := int(time.Monday - now.Weekday())

	if offset > 0 {
		offset = -6
	}

	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)

	monday = today.AddDate(0, 0, offset).Format("20060102")

	offset = int(time.Sunday - now.Weekday())

	if offset < 0 {
		offset += 7
	}

	sunday = today.AddDate(0, 0, offset).Format("20060102")

	return
}

// IP2Long converts a string containing an (IPv4) Internet Protocol dotted address into a long integer.
func IP2Long(ip string) uint32 {
	ipv4 := net.ParseIP(ip).To4()

	if ipv4 == nil {
		return 0
	}

	return uint32(ipv4[0])<<24 | uint32(ipv4[1])<<16 | uint32(ipv4[2])<<8 | uint32(ipv4[3])
}

// Long2IP converts an long integer address into a string in (IPv4) Internet standard dotted format.
func Long2IP(ip uint32) string {
	if ip > math.MaxUint32 {
		return ""
	}

	return net.IPv4(byte(ip>>24), byte(ip>>16), byte(ip>>8), byte(ip)).String()
}

// BufferPool type of buffer pool
type BufferPool struct {
	pool sync.Pool
}

// Get return a buffer
func (b *BufferPool) Get() *bytes.Buffer {
	buf := b.pool.Get().(*bytes.Buffer)
	buf.Reset()

	return buf
}

// Put put a buffer to pool
func (b *BufferPool) Put(buf *bytes.Buffer) {
	if buf == nil {
		return
	}

	b.pool.Put(buf)
}

// NewBufferPool returns a new buffer pool
func NewBufferPool(cap int64) *BufferPool {
	return &BufferPool{pool: sync.Pool{
		New: func() interface{} {
			return bytes.NewBuffer(make([]byte, 0, cap))
		},
	}}
}

// BufPool buffer pool
var BufPool = NewBufferPool(4 << 10) // 4KB

type ginValidator struct {
	validator  *validator.Validate
	translator ut.Translator
}

// ValidateStruct receives any kind of type, but only performed struct or pointer to struct type.
func (v *ginValidator) ValidateStruct(obj interface{}) error {
	if err := v.validator.Struct(obj); err != nil {
		e, ok := err.(validator.ValidationErrors)

		if !ok {
			return e
		}

		errM := e.Translate(v.translator)
		msgs := make([]string, 0, len(errM))

		for _, v := range errM {
			msgs = append(msgs, v)
		}

		return errors.New(strings.Join(msgs, ";"))
	}

	return nil
}

// Engine returns the underlying validator engine which powers the default
// Validator instance. This is useful if you want to register custom validations
// or struct level validations. See validator GoDoc for more info -
// https://godoc.org/gopkg.in/go-playground/validator.v10
func (v *ginValidator) Engine() interface{} {
	return v.validator
}

// NewGinValidator returns a validator for gin
func NewGinValidator() *ginValidator {
	zhCn := zh.New()
	uniTrans := ut.New(zhCn)

	validate := validator.New()
	validate.SetTagName("valid")

	translator, _ := uniTrans.GetTranslator("zh")

	zhcn.RegisterDefaultTranslations(validate, translator)

	return &ginValidator{
		validator:  validate,
		translator: translator,
	}
}

// VersionCompare compares semantic versions range, support: >, >=, =, !=, <, <=, | (or), & (and)
// eg: 1.0.0, =1.0.0, >2.0.0, >=1.0.0&<2.0.0, <2.0.0|>3.0.0, !=4.0.4
func VersionCompare(rangeVer, curVer string) bool {
	if rangeVer == "" || curVer == "" {
		return true
	}

	semVer, err := version.NewVersion(curVer)

	if err != nil {
		logger.Error("invalid semantic version", zap.Error(err), zap.String("range_version", rangeVer), zap.String("cur_version", curVer))

		return true
	}

	orVers := strings.Split(rangeVer, "|")

	for _, ver := range orVers {
		andVers := strings.Split(ver, "&")

		constraints, err := version.NewConstraint(strings.Join(andVers, ","))

		if err != nil {
			logger.Error("version compared error", zap.Error(err), zap.String("range_version", rangeVer), zap.String("cur_version", curVer))

			return true
		}

		if constraints.Check(semVer) {
			return true
		}
	}

	return false
}
