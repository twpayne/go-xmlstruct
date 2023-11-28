package xmlstruct

import (
	"encoding/xml"
	"strconv"
	"time"
)

// A value describes an observed simple value, either an attribute value or
// chardata.
type value struct {
	boolCount                 int
	float64Count              int
	intCount                  int
	name                      xml.Name
	observations              int
	optional                  bool
	repeated                  bool
	stringCount               int
	timeCount                 int
	unexpectedElementTypeName string
}

// goType returns the most specific Go type that can represent all of the values
// observed for v.
func (v *value) goType(options *generateOptions) string {
	distinctTypes := 0
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
	if v.unexpectedElementTypeName != "" {
		return v.unexpectedElementTypeName
	}
	switch {
	case distinctTypes == 0:
		return "struct{}"
	case distinctTypes == 1 && v.boolCount > 0:
		return prefix + "bool"
	case distinctTypes == 1 && v.intCount > 0:
		return prefix + options.intType
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

// observe records s as being observed for v.
func (v *value) observe(s string, options *observeOptions) {
	v.observations++
	if _, err := strconv.ParseInt(s, 10, 64); err == nil {
		v.intCount++
		return
	}
	if _, err := strconv.ParseBool(s); err == nil {
		v.boolCount++
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
