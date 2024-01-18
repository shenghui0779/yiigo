package util

import "time"

// GMT8 东八区时区
var GMT8 = time.FixedZone("CST", 8*3600)

// TimeToStr 时间戳格式化为时间字符串
// 若 timestamp < 0，则使用 `time.Now()`
func TimeToStr(timestamp int64, layout string) string {
	if timestamp < 0 {
		return time.Now().In(GMT8).Format(layout)
	}

	return time.Unix(timestamp, 0).In(GMT8).Format(layout)
}

// StrToTime 时间字符串解析为时间戳
func StrToTime(datetime, layout string) time.Time {
	t, _ := time.ParseInLocation(layout, datetime, GMT8)
	return t
}

// WeekAround 返回给定时间戳所在周的「周一」和「周日」时间字符串
func WeekAround(timestamp int64, layout string) (monday, sunday string) {
	t := time.Unix(timestamp, 0).In(GMT8)
	weekday := t.Weekday()

	// monday
	offset := int(time.Monday - weekday)
	if offset > 0 {
		offset = -6
	}
	today := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, GMT8)
	monday = today.AddDate(0, 0, offset).Format(layout)

	// sunday
	offset = int(time.Sunday - weekday)
	if offset < 0 {
		offset += 7
	}
	sunday = today.AddDate(0, 0, offset).Format(layout)

	return
}
