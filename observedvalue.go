package xmlstruct

import (
	"strconv"
	"time"
)

type ObservedValue struct {
	repeated     bool
	optional     bool
	observations int
	emptyCount   int
	boolCount    int
	intCount     int
	float64Count int
	timeCount    int
	stringCount  int
}

func (v *ObservedValue) goType(options *sourceOptions) string {
	distinctTypes := 0
	if v.emptyCount > 0 {
		distinctTypes++
	}
	if v.boolCount > 0 {
		distinctTypes++
	}
	if v.intCount > 0 {
		distinctTypes++
	}
	if v.float64Count > 0 {
		distinctTypes++
	}
	if v.timeCount > 0 {
		distinctTypes++
	}
	if v.stringCount > 0 {
		distinctTypes++
	}
	prefix := ""
	if v.repeated {
		prefix += "[]"
	}
	if options.usePointersForOptionalFields && v.optional {
		prefix += "*"
	}
	switch {
	case distinctTypes == 1 && v.boolCount > 0:
		return prefix + "bool"
	case distinctTypes == 1 && v.intCount > 0:
		return prefix + "int"
	case distinctTypes == 1 && v.float64Count > 0:
		return prefix + "float64"
	case distinctTypes == 1 && v.timeCount > 0:
		options.importPackageNames["time"] = struct{}{}
		return prefix + "time.Time"
	case distinctTypes == 2 && v.intCount > 0 && v.float64Count > 0:
		return prefix + "float64"
	default:
		return prefix + "string"
	}
}

func (v *ObservedValue) observe(s string, options *observeOptions) {
	v.observations++
	if s == "" {
		v.emptyCount++
		return
	}
	if _, err := strconv.ParseBool(s); err == nil {
		v.boolCount++
		return
	}
	if _, err := strconv.ParseInt(s, 10, 64); err == nil {
		v.intCount++
		return
	}
	if _, err := strconv.ParseFloat(s, 64); err == nil {
		v.float64Count++
		return
	}
	if options.timeLayout != "" {
		if _, err := time.Parse(options.timeLayout, s); err == nil {
			v.timeCount++
			return
		}
	}
	v.stringCount++
}
