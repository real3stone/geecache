package geecache

type ByteView struct {
	b []byte
}


// 重写Len()方法，这样相当于实现了Value接口，所以ByteView可以作为Value的实现类传入函数
func (v ByteView) Len() int {
	return len(v.b)
}

func (v ByteView) ByteSlice() []byte {
	return cloneBytes(v.b)
}

// byte[]转换为string
func (v ByteView) String() string {
	return string(v.b)
}

func cloneBytes(b []byte) []byte {
	c := make([]byte, len(b))
	copy(c, b)
	return c
}

