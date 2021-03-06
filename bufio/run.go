/**
 * @Author: realpeanut
 * @Date: 2020/11/3 10:50 上午
 */
package main

import (
	"bufio"
	"fmt"
	"net"
)

/**
	顾名思义，bufio是io的缓冲，重新实现了reader和writer
 */

/**
	eg. NewReaderSize(rd io.Reader, size int)->ReadByte() 分析
	目的：从buf中读取1个字节
	步骤：
		1.初始化buf
		2.fill()填充，（buf可读为空时填充）
		3.读取字节数据
	code :读取redis response
			conn,_ := net.Dial("tcp4","localhost:6379")
			var b = []byte("$2\r\n$3\r\nget\r\n$1a\r\n")
			_,_ = conn.Write(b)
			r := bufio.NewReader(conn)
			_,_ = r.ReadByte()
	分析：
		bufio的一个最关键的结构体是Reader
        Reader 实现了 io.Reader 接口.
		type Reader struct {
			buf          []byte    // 缓冲区
			rd           io.Reader // io.Reader
			r, w         int       // 已读buf和已写入buf的位置
			err          error     // 错误
			lastByte     int       // 最后一个字节是否已读
			lastRuneSize int       // 最后四个字节是否已读
		}
		//最小缓存大小
		const minReadBufferSize = 16
		//连续写入最多100次，直到有数据写入直接break,每次读取buf剩余大小
        const maxConsecutiveEmptyReads = 100

		//按照用户指定的size 初始化一个Reader buf
		//Reader是io.Reader的实现,rd.(*Reader)判断rd是否是*Reader类型
		func NewReaderSize(rd io.Reader, size int) *Reader {
			b, ok := rd.(*Reader)
			//如果是*Reader 而且缓存大小>=size
			if ok && len(b.buf) >= size {
				return b
			}
			if size < minReadBufferSize {
				size = minReadBufferSize
			}
			r := new(Reader)
			//或者 r := &Reader{}
			r.reset(make([]byte, size), rd)
			return r
		}

		//从buf 中读取一个字节
		func (b *Reader) ReadByte() (byte, error) {
			b.lastRuneSize = -1
			//r==w 代表所以写入都已经读取 所以执行 fill()向缓存区写入数据
			for b.r == b.w {
				if b.err != nil {
					return 0, b.readErr()
				}
				b.fill() // buffer is empty
			}
			c := b.buf[b.r]
			//已读位置加1
			b.r++
			//最后读取的字节赋值
			b.lastByte = int(c)
			//返回一个字节数据
			return c, nil
		}
		// 读取数据填充到buf
		func (b *Reader) fill() {


			if b.r > 0 {
				//有未读时才会copy ，没有未读时,r == w，所以是不会copy的
				copy(b.buf, b.buf[b.r:b.w])
				b.w -= b.r
                //已读初始化为0
				b.r = 0
			}
			//已写入数据大于缓存大小时，直接报错
			if b.w >= len(b.buf) {
				panic("bufio: tried to fill full buffer")
			}

			//连续写入最多100次，直到有数据写入直接break,每次读取buf剩余大小，
			//w位置要更新
			for i := maxConsecutiveEmptyReads; i > 0; i-- {
				n, err := b.rd.Read(b.buf[b.w:])
				if n < 0 {
					panic(errNegativeRead)
				}
				b.w += n
				if err != nil {
					b.err = err
					return
				}
				if n > 0 {
					return
				}
			}
			b.err = io.ErrNoProgress
		}
 */

func main()  {
	conn,_ := net.Dial("tcp4","localhost:6379")
	var b = []byte("$2\r\n$3\r\nget\r\n$1a\r\n")
	_,_ = conn.Write(b)
	r := bufio.NewReader(conn)
	res,_:=r.ReadSlice('\n')
	fmt.Println(res)
}

/**
	读取一rune
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
 */

/**
	//读取指定字节，但是不更新b.r，下次仍然能读取的到，前提是读取小于buf len

	func (b *Reader) Peek(n int) ([]byte, error) {
		if n < 0 {
			return nil, ErrNegativeCount
		}

		b.lastByte = -1
		b.lastRuneSize = -1

		//可读小于n & 缓存未满  此时填充数据
		for b.w-b.r < n && b.w-b.r < len(b.buf) && b.err == nil {
			b.fill() // b.w-b.r < len(b.buf) => buffer is not full
		}


		if n > len(b.buf) {
			return b.buf[b.r:b.w], ErrBufferFull
		}

		// 0 <= n <= len(b.buf)
		var err error
		if avail := b.w - b.r; avail < n {
			// not enough data in buffer
			n = avail
			err = b.readErr()
			if err == nil {
				err = ErrBufferFull
			}
		}
		return b.buf[b.r : b.r+n], err
	}
 */

/**
	读取指定分割符之前的所有字节
	例如读取\n之前的所有数据

	func (b *Reader) ReadSlice(delim byte) (line []byte, err error) {
		s := 0
		//循环读取数据遍历
		for {
			//搜索指定byte
			if i := bytes.IndexByte(b.buf[b.r+s:b.w], delim); i >= 0 {
                //返回未读buf到i+s的字节数据
				i += s
				line = b.buf[b.r : b.r+i+1]
				b.r += i + 1
				break
			}

			// 未找到返回buf所有未读数据，返回error
			if b.err != nil {
				line = b.buf[b.r:b.w]
				b.r = b.w
				err = b.readErr()
				break
			}

			// 未读大于等于缓存大小，说明没有找到，将返回所有buf缓存，所有判断查找失败应该看err,而不是line有没有数据
			if b.Buffered() >= len(b.buf) {
				b.r = b.w
				line = b.buf
				err = ErrBufferFull
				break
			}

			s = b.w - b.r // do not rescan area we scanned before

			b.fill() // buffer is not full
		}

		// Handle last byte, if any.
		if i := len(line) - 1; i >= 0 {
			b.lastByte = int(line[i])
			b.lastRuneSize = -1
		}

	return
}


// IndexByte returns the index of the first instance of c in b, or -1 if c is not present in b.
func IndexByte(b []byte, c byte) int {
	return bytealg.IndexByte(b, c)
}
 */




/**
	读取缓存区所有数据 缓存区最大为4096byte，直接读取4096byte即可
	conn,_ := net.Dial("tcp4","localhost:6379")
	var b = []byte("$2\r\n$3\r\nget\r\n$1a\r\n")
	_,_ = conn.Write(b)
	r := bufio.NewReader(conn)
	_,_=r.Read(make([]byte,4096))
	//
	func (b *Reader) Read(p []byte) (n int, err error) {
		n = len(p)
		//读取0字节时，直接返回0
		if n == 0 {
			//判断可读字节
			if b.Buffered() > 0 {
				return 0, nil
			}
			//可读字节 <= 0 返回error
			return 0, b.readErr()
		}
		// 如果缓存数据已被读取或者缓存buf中为空时
		if b.r == b.w {
			if b.err != nil {
				return 0, b.readErr()
			}
			//如果想读的字节数大于、等于缓存大小时
			if len(p) >= len(b.buf) {

				//直接从io中读取想读的字节，因为读出来的数据大于缓存，则不需要往缓存中添加数据
				n, b.err = b.rd.Read(p)
				if n < 0 {
					panic(errNegativeRead)
				}
				if n > 0 {
					b.lastByte = int(p[n-1])
					b.lastRuneSize = -1
				}

				return n, b.readErr()
			}
			//读取的数据小于等于缓存大小
			b.r = 0
			b.w = 0
			//buf读满，这里可能会阻塞
			n, b.err = b.rd.Read(b.buf)
			if n < 0 {
				panic(errNegativeRead)
			}
			if n == 0 {
				return 0, b.readErr()
			}
			b.w += n
	}

	//copy函数copy 字节较小的数，len(p) < b.buf[b.r:b.w]  最多读取len(p)
	n = copy(p, b.buf[b.r:b.w])
	b.r += n
	b.lastByte = int(b.buf[b.r-1])
	b.lastRuneSize = -1
	return n, nil
}

func (b *Reader) Buffered() int {
	return b.w - b.r
}

 */