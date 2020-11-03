
### 读取一rune

````
func (b *Reader) ReadRune() (r rune, size int, err error) {
    
    for b.r+utf8.UTFMax > b.w && !utf8.FullRune(b.buf[b.r:b.w]) && b.err == nil && b.w-b.r < len(b.buf) {
        b.fill() // b.w-b.r < len(buf) => buffer is not full
    }
    b.lastRuneSize = -1
    if b.r == b.w {
        return 0, 0, b.readErr()
    }
    r, size = rune(b.buf[b.r]), 1
    //判断是否是ascii
    if r >= utf8.RuneSelf {
        //解码
        r, size = utf8.DecodeRune(b.buf[b.r:b.w])
    }
    b.r += size
    b.lastByte = int(b.buf[b.r-1])
    b.lastRuneSize = size
    return r, size, nil
}
````

