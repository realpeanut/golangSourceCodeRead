/**
 * @Author: realpeanut
 * @Date: 2020/11/3 10:50 上午
 */
package main

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