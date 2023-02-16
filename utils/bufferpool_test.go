package utils

import (
	"bytes"
	"testing"
)

var count = 5
var str = `{}`
var strlong = `{}`

func Benchmark_TestStringAppend(b *testing.B) {
	for i := 0; i < b.N; i++ {
		s := ""
		for j := 0; j < count; j++ {
			s = s + str
		}
	}
}

func Benchmark_TestBufferPool(b *testing.B) {
	b.StopTimer() //调用该函数停止压力测试的时间计数
	var buffers = NewBufferPool(516)
	b.StartTimer() //重新开始时间
	for i := 0; i < b.N; i++ {
		buf := buffers.Get()
		for j := 0; j < count; j++ {
			buf.WriteString(str)
		}
		buffers.Put(buf)
	}
}

func Benchmark_TestBuffer(b *testing.B) {
	for i := 0; i < b.N; i++ {
		buf := bytes.NewBufferString("")
		for j := 0; j < count; j++ {
			buf.WriteString(str)
		}
	}
}

func BenchmarkJI_TestBufferPool_Parallel(b *testing.B) {
	b.StopTimer() //调用该函数停止压力测试的时间计数
	var buffers = NewBufferPool(516)
	b.StartTimer() //重新开始时间
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			buf := buffers.Get()
			for j := 0; j < count; j++ {
				buf.WriteString(str)
			}
			buffers.Put(buf)
		}
	})
}

func BenchmarkJI_TestBuffer_Parallel(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			buf := bytes.NewBufferString("")
			for j := 0; j < count; j++ {
				buf.WriteString(str)
			}
		}
	})
}

func BenchmarkJI_TestStringAppend_Parallel(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			s := ""
			for j := 0; j < count; j++ {
				s = s + str
			}
		}
	})
}
