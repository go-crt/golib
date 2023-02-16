package exstrings

import (
	"reflect"
	"unsafe"
)

/*
UnsafeToBytes 把 string 转换为 []byte 没有多余的内存开销。
这个函数可以提升 string 转 []byte 的性能，并极大的降低内存开销，但是却相当的危险，对于不明确这个函数的人来说不建议使用该函数。
这个函数是 exbytes.ToString 反向操作版，但是不像 exbytes.ToString 那么的稳定安全，该函数使用不当会导致程序直接崩溃，且无法恢复。
	s := `{"k":"v"}`
	b := exstrings.UnsafeToBytes(s)
	// b[3] = 'k' // unexpected fault address 0x1118180
	data := map[string]string{}
	err := json.Unmarshal(b, &data)
	fmt.Println(data, err)
这是一个使用的例子，如果我们需要转换一个字符串很方便，且开销非常的低。
但是一定要注意，b[3] = 'k' 如果尝试修改获得的 []byte 将直接导致程序崩溃，并且不可能通过 recover() 恢复。
实际上我们可以突破这个限制，这就要了解字符串的一些规则，下面的例子可以完美运行，并修改字符串：
	s := strings.Repeat("A", 3)
	b := exstrings.UnsafeToBytes(s)
	b[1] = 'B'
	b[2] = 'C'
	fmt.Println(s, string(b))
非常完美，s和b变量的值都是 ABC， 为什么会这样呢？
这个就是 string 的内存分配方法， 字面量使用这种方式是没有办法修改的，因为这是在编译时就决定的，编译时会设定字符串的内存数据是只读数据。
如果程序运行时生成的数据，这种数据是可以安全使用该函数的，但是要当心你的字符串可能会被修改，
比如我们调用 json.Unmarshal(exstrings.UnsafeToBytes(s), &data)， 如果 json 包里面出现修改输入参数，那么原来的字符串就可能不是你想想的那样。
使用该函数要明确两个事情：
	- 确定字符串是否是字面量，或者内存被分配在只读空间上。
	- 确保访问该函数结果的函数是否会按照你的意愿访问或修改数据。
我公开该函数经过很多思考，虽然它很危险，但是有时间却很有用，如果我们需要大批量转换字符串的大小写，而且不再需要原字符串，我们可以原地安全的修改字符串。
当然还有更多的使用方法，可以极大的提升我们程序的性能。
*/
func UnsafeToBytes(s string) []byte {
	strHeader := (*reflect.StringHeader)(unsafe.Pointer(&s))
	return *(*[]byte)(unsafe.Pointer(&reflect.SliceHeader{
		Data: strHeader.Data,
		Len:  strHeader.Len,
		Cap:  strHeader.Len,
	}))
}
