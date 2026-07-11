package main

func Byte() {
	var a byte = 'a'
	// byte 是 uint8 的别名，所以这里输出字符 a 的 ASCII 编码 97。
	println(a)

	var str string = "this is string"
	var bs []byte = []byte(str)
	println(bs)
}
