package yiigo

import (
	"encoding/xml"
	"errors"
	"net"
	"reflect"
	"strings"
	"time"

	"github.com/go-playground/locales/zh"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	zhcn "github.com/go-playground/validator/v10/translations/zh"
	"github.com/hashicorp/go-version"
	"go.uber.org/zap"
)

// Default defines for `default` name
const Default = "default"

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

	if len(layout) != 0 {
		l = layout[0]
	}

	date := time.Unix(timestamp, 0).Local().Format(l)

	return date
}

// StrToTime Parse English textual datetime description into a Unix timestamp.
// The default layout is: 2006-01-02 15:04:05.
func StrToTime(datetime string, layout ...string) int64 {
	l := "2006-01-02 15:04:05"

	if len(layout) != 0 {
		l = layout[0]
	}

	t, err := time.ParseInLocation(l, datetime, time.Local)

	// mismatch layout
	if err != nil {
		logger.Error("yiigo: parse layout mismatch", zap.Error(err))

		return 0
	}

	return t.Unix()
}

// WeekAround returns the date of monday and sunday for current week.
func WeekAround(t time.Time) (monday, sunday string) {
	weekday := t.Local().Weekday()

	// monday
	offset := int(time.Monday - weekday)

	if offset > 0 {
		offset = -6
	}

	today := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.Local)

	monday = today.AddDate(0, 0, offset).Format("20060102")

	// sunday
	offset = int(time.Sunday - weekday)

	if offset < 0 {
		offset += 7
	}

	sunday = today.AddDate(0, 0, offset).Format("20060102")

	return
}

// IP2Long converts a string containing an (IPv4) Internet Protocol dotted address into an uint32 integer.
func IP2Long(ip string) uint32 {
	ipv4 := net.ParseIP(ip).To4()

	if ipv4 == nil {
		return 0
	}

	return uint32(ipv4[0])<<24 | uint32(ipv4[1])<<16 | uint32(ipv4[2])<<8 | uint32(ipv4[3])
}

// Long2IP converts an uint32 integer address into a string in (IPv4) Internet standard dotted format.
func Long2IP(ip uint32) string {
	return net.IPv4(byte(ip>>24), byte(ip>>16), byte(ip>>8), byte(ip)).String()
}

// Validator a validator which can be used for Gin.
type Validator struct {
	validator  *validator.Validate
	translator ut.Translator
}

// ValidateStruct receives any kind of type, but only performed struct or pointer to struct type.
func (v *Validator) ValidateStruct(obj interface{}) error {
	if reflect.Indirect(reflect.ValueOf(obj)).Kind() != reflect.Struct {
		return nil
	}

	if err := v.validator.Struct(obj); err != nil {
		e, ok := err.(validator.ValidationErrors)

		if !ok {
			return err
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
func (v *Validator) Engine() interface{} {
	return v.validator
}

// NewValidator returns a new validator.
// Used for Gin: binding.Validator = yiigo.NewValidator()
func NewValidator() *Validator {
	locale := zh.New()
	uniTrans := ut.New(locale)

	validate := validator.New()
	validate.SetTagName("valid")

	translator, _ := uniTrans.GetTranslator("zh")

	zhcn.RegisterDefaultTranslations(validate, translator)

	return &Validator{
		validator:  validate,
		translator: translator,
	}
}

// VersionCompare compares semantic versions range, support: >, >=, =, !=, <, <=, | (or), & (and).
// Param `rangeVer` eg: 1.0.0, =1.0.0, >2.0.0, >=1.0.0&<2.0.0, <2.0.0|>3.0.0, !=4.0.4
func VersionCompare(rangeVer, curVer string) (bool, error) {
	semVer, err := version.NewVersion(curVer)

	// invalid semantic version
	if err != nil {
		return false, err
	}

	orVers := strings.Split(rangeVer, "|")

	for _, ver := range orVers {
		andVers := strings.Split(ver, "&")

		constraints, err := version.NewConstraint(strings.Join(andVers, ","))

		if err != nil {
			return false, err
		}

		if constraints.Check(semVer) {
			return true, nil
		}
	}

	return false, nil
}
