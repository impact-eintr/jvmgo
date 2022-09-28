package heap

import "unicode/utf16"

var internedStrings = map[string]*Object{}

// go string -> java.lang.String
func JString(loader *ClassLoader, goStr string) *Object {
	if internedStr, ok := internedStrings[goStr];ok {
		return internedStr
	}

	chars := stringToUtf16(goStr)
	jChars := &Object{
		loader.LoadClass("[C"),
		chars,
		nil,
	}

	jStr := loader.LoadClass("java/lang/String").NewObject()
	jStr.SetRefVar("value", "[C", jChars)

	internedStrings[goStr] = jStr
	return jStr
}

func GoString(jStr *Object) string {
	charArr := jStr.GetRefVar("value", "[C")
	return utf16ToString(charArr.Chars())
}

// utf8  -> utf16
func stringToUtf16(s string) []uint16 {
	runes := []rune(s)
	return utf16.Encode(runes)
}

// utf16 -> utf8
func utf16ToString(s []uint16) string {
	runes := utf16.Decode(s)
	return string(runes)
}
