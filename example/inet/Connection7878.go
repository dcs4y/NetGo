package inet

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/dcs4y/NetGo/ginterface"
	"github.com/dcs4y/NetGo/gnet"
	"net"
	"strings"
)

type Connection7878 struct {
	gnet.Connection
}

// NewConnection  创建连接的方法
func NewConnection(server ginterface.IServer, conn *net.TCPConn) (ginterface.IConnection, error) {
	// 起始位
	buf := make([]byte, 1)
	bufLength, err := conn.Read(buf)
	if err != nil {
		fmt.Println(err)
		conn.Close()
		return nil, err
	}
	if bufLength == 0 {
		conn.Close()
		return nil, errors.New("未读取到起始位信息，连接建立失败！")
	}

	var protocolNo string
	// 7878|7979 (0x78=120,0x79=121)
	if buf[0] == 120 || buf[0] == 121 {
		// 读取第二位，此处的值与第一位是一样的。
		bufLength, err := conn.Read(buf)
		if err != nil {
			fmt.Println(err)
			conn.Close()
			return nil, err
		}
		if bufLength == 0 {
			conn.Close()
			return nil, errors.New("未读取到完整协议头，连接建立失败！")
		}
		if buf[0] == 120 {
			protocolNo = "7878"
		} else if buf[0] == 121 {
			protocolNo = "7979"
		}
	}

	oldMessage, err := parseMessage(conn, protocolNo)
	if err != nil {
		return nil, err
	}
	body := oldMessage.GetBody()

	if oldMessage.GetAction() != "01" {
		conn.Close()
		return nil, errors.New("非登录包，停止连接！" + fmt.Sprintf("%#v", oldMessage))
	}
	deviceNo := strings.TrimLeft(body[:16], "0")

	//初始化Conn属性
	connection := &Connection7878{}
	c := gnet.Connection{
		Server:          server,
		Conn:            conn,
		ConnID:          server.GetConnectionIndex(),
		ConnName:        deviceNo,
		IsClosed:        false,
		MessageHandler:  server.GetMessageHandle(),
		MessageChan:     make(chan []byte),
		MessageBuffChan: make(chan []byte),
	}
	connection.Connection = c

	connection.StartReader = connection.iStartReader
	connection.Pack = connection.iPack

	// 建立连接后回复消息
	connection.MessageHandler.HandleRequest(connection, oldMessage)

	return connection, nil
}

//iStartReader 读消息Goroutine，用于从客户端中读取数据
func (c *Connection7878) iStartReader() {
	fmt.Println("[Reader Goroutine is running]")
	defer fmt.Println(c.RemoteAddr().String(), "[conn Reader exit!]")
	defer c.Stop()

	for {
		select {
		case <-c.Ctx.Done():
			return
		default:
			//fmt.Println("从socket中读取数据。。。")

			// 起始位
			buf := make([]byte, 2)
			bufLength, err := c.Conn.Read(buf)
			if err != nil {
				fmt.Println(err)
				return
			}
			if bufLength == 0 {
				continue
			}

			start := hex.EncodeToString(buf)
			message, err := parseMessage(c.Conn, start)
			if err != nil {
				fmt.Println(err)
				return
			}

			// 响应客户端消息
			c.MessageHandler.HandleRequest(c, message)
		}
	}
}

//iPack 封包方法(压缩数据)
func (c *Connection7878) iPack(msg ginterface.IMessage) ([]byte, error) {
	if msg.GetProtocolNo() == "" {
		msg.SetProtocolNo(c.GetProtocolNo())
	}
	// 长度=协议号+信息内容+信息序列号+错误校验
	//index := strconv.FormatUint(msg.GetIndex(), 16)
	//fmt.Printf("%#v\n", msg)
	mData, err := hex.DecodeString(msg.GetDataLen() + msg.GetAction() + msg.GetBody() + msg.GetIndex())
	if err != nil {
		fmt.Println(err)
	}
	checkCode := fmt.Sprintf("%04X", crcCheckCode(mData))
	result := strings.ToUpper(msg.GetProtocolNo() + msg.GetDataLen() + msg.GetAction() + msg.GetBody() + msg.GetIndex() + checkCode + c.getMessageEnd())
	//fmt.Println("7878封包结果：" + result)
	return hex.DecodeString(result)
}

// GetProtocolNo 获取连接的开始标识
func (c *Connection7878) GetProtocolNo() string {
	// 深圳市几米物联有限公司
	return "7878"
}

// GetMessageEnd 获取连接的结束标识
func (c *Connection7878) getMessageEnd() string {
	return "0d0a"
}

//unpack 拆包方法(解压数据) 包长度，协议号+信息内容+信息序列号+错误校验
func unpack(protocolNo string, lengthByte []byte, contentByte []byte) (ginterface.IMessage, error) {
	//fmt.Println("去头尾数据：" + hex.EncodeToString(bytesCombine(lengthByte, contentByte)))

	crcCode := hex.EncodeToString(contentByte[len(contentByte)-2:])
	temp := bytesCombine(lengthByte, contentByte[:len(contentByte)-2])
	//fmt.Println("拆包数据(去头尾数据、错误校验)：" + hex.EncodeToString(temp))
	ourCrcCode := fmt.Sprintf("%04x", crcCheckCode(temp))
	if ourCrcCode != crcCode {
		return nil, errors.New("校验码验证失败！" + ourCrcCode + "=" + crcCode)
	}

	//只解压head的信息，得到dataLen和msgID
	msg := &gnet.Message{
		ProtocolNo: protocolNo,
		BodyData:   contentByte[1 : len(contentByte)-4],
		DataLen:    hex.EncodeToString(lengthByte),
	}
	//fmt.Println("BodyData=" + hex.EncodeToString(msg.GetBodyData()))
	content16 := hex.EncodeToString(contentByte)
	msg.Action = content16[0:2]
	msg.Body = content16[2 : len(content16)-8]
	//fmt.Println("Body=" + msg.GetBody())
	/*index, err := strconv.ParseUint(content16[len(content16)-8:len(content16)-4], 16, 32)
	if err != nil {
		fmt.Println(err)
	}*/
	msg.Index = content16[len(content16)-8 : len(content16)-4]

	//这里只需要把head的数据拆包出来就可以了，然后再通过head的长度，再从conn读取一次数据
	return msg, nil
}

func parseMessage(conn *net.TCPConn, protocolNo string) (message ginterface.IMessage, err error) {
	var buf []byte
	// 包长度
	var length uint16
	var bufLength int
	if protocolNo == "7878" {
		buf = make([]byte, 1)
		bufLength, err = conn.Read(buf)
		if err != nil {
			fmt.Println(err)
			return
		}
		if bufLength == 0 {
			return
		}
		bytesBuffer := bytes.NewBuffer(buf)
		var tmp uint8
		err = binary.Read(bytesBuffer, binary.BigEndian, &tmp)
		if err != nil {
			fmt.Println(err)
			return
		}
		length = uint16(tmp)
	} else if protocolNo == "7979" {
		buf = make([]byte, 2)
		bufLength, err = conn.Read(buf)
		if err != nil {
			fmt.Println(err)
			return
		}
		if bufLength == 0 {
			return
		}
		bytesBuffer := bytes.NewBuffer(buf)
		var tmp uint16
		err = binary.Read(bytesBuffer, binary.BigEndian, &tmp)
		if err != nil {
			fmt.Println(err)
			return
		}
		length = tmp
	} else {
		return
	}

	// 包内容=协议号+信息内容+信息序列号+错误校验
	contentBuf := make([]byte, length)
	bufLength, err = conn.Read(contentBuf)
	if err != nil {
		fmt.Println(err)
		return
	}
	if bufLength == 0 {
		return
	}

	// 停止位
	end := make([]byte, 2)
	bufLength, err = conn.Read(end)
	if err != nil {
		fmt.Println(err)
		return
	}
	if bufLength == 0 {
		return
	}
	//fmt.Println("停止位数据：" + hex.EncodeToString(end))

	//拆包，得到msgID 和 datalen 放在msg中
	message, err = unpack(protocolNo, buf, contentBuf)
	if err != nil {
		//fmt.Println("unpack error ", err)
		return
	}
	return
}

var mbTable = []uint16{
	0x0000, 0x1189, 0x2312, 0x329B, 0x4624, 0x57AD, 0x6536, 0x74BF,
	0x8C48, 0x9DC1, 0xAF5A, 0xBED3, 0xCA6C, 0xDBE5, 0xE97E, 0xF8F7,
	0x1081, 0x0108, 0x3393, 0x221A, 0x56A5, 0x472C, 0x75B7, 0x643E,
	0x9CC9, 0x8D40, 0xBFDB, 0xAE52, 0xDAED, 0xCB64, 0xF9FF, 0xE876,
	0x2102, 0x308B, 0x0210, 0x1399, 0x6726, 0x76AF, 0x4434, 0x55BD,
	0xAD4A, 0xBCC3, 0x8E58, 0x9FD1, 0xEB6E, 0xFAE7, 0xC87C, 0xD9F5,
	0x3183, 0x200A, 0x1291, 0x0318, 0x77A7, 0x662E, 0x54B5, 0x453C,
	0xBDCB, 0xAC42, 0x9ED9, 0x8F50, 0xFBEF, 0xEA66, 0xD8FD, 0xC974,
	0x4204, 0x538D, 0x6116, 0x709F, 0x0420, 0x15A9, 0x2732, 0x36BB,
	0xCE4C, 0xDFC5, 0xED5E, 0xFCD7, 0x8868, 0x99E1, 0xAB7A, 0xBAF3,
	0x5285, 0x430C, 0x7197, 0x601E, 0x14A1, 0x0528, 0x37B3, 0x263A,
	0xDECD, 0xCF44, 0xFDDF, 0xEC56, 0x98E9, 0x8960, 0xBBFB, 0xAA72,
	0x6306, 0x728F, 0x4014, 0x519D, 0x2522, 0x34AB, 0x0630, 0x17B9,
	0xEF4E, 0xFEC7, 0xCC5C, 0xDDD5, 0xA96A, 0xB8E3, 0x8A78, 0x9BF1,
	0x7387, 0x620E, 0x5095, 0x411C, 0x35A3, 0x242A, 0x16B1, 0x0738,
	0xFFCF, 0xEE46, 0xDCDD, 0xCD54, 0xB9EB, 0xA862, 0x9AF9, 0x8B70,
	0x8408, 0x9581, 0xA71A, 0xB693, 0xC22C, 0xD3A5, 0xE13E, 0xF0B7,
	0x0840, 0x19C9, 0x2B52, 0x3ADB, 0x4E64, 0x5FED, 0x6D76, 0x7CFF,
	0x9489, 0x8500, 0xB79B, 0xA612, 0xD2AD, 0xC324, 0xF1BF, 0xE036,
	0x18C1, 0x0948, 0x3BD3, 0x2A5A, 0x5EE5, 0x4F6C, 0x7DF7, 0x6C7E,
	0xA50A, 0xB483, 0x8618, 0x9791, 0xE32E, 0xF2A7, 0xC03C, 0xD1B5,
	0x2942, 0x38CB, 0x0A50, 0x1BD9, 0x6F66, 0x7EEF, 0x4C74, 0x5DFD,
	0xB58B, 0xA402, 0x9699, 0x8710, 0xF3AF, 0xE226, 0xD0BD, 0xC134,
	0x39C3, 0x284A, 0x1AD1, 0x0B58, 0x7FE7, 0x6E6E, 0x5CF5, 0x4D7C,
	0xC60C, 0xD785, 0xE51E, 0xF497, 0x8028, 0x91A1, 0xA33A, 0xB2B3,
	0x4A44, 0x5BCD, 0x6956, 0x78DF, 0x0C60, 0x1DE9, 0x2F72, 0x3EFB,
	0xD68D, 0xC704, 0xF59F, 0xE416, 0x90A9, 0x8120, 0xB3BB, 0xA232,
	0x5AC5, 0x4B4C, 0x79D7, 0x685E, 0x1CE1, 0x0D68, 0x3FF3, 0x2E7A,
	0xE70E, 0xF687, 0xC41C, 0xD595, 0xA12A, 0xB0A3, 0x8238, 0x93B1,
	0x6B46, 0x7ACF, 0x4854, 0x59DD, 0x2D62, 0x3CEB, 0x0E70, 0x1FF9,
	0xF78F, 0xE606, 0xD49D, 0xC514, 0xB1AB, 0xA022, 0x92B9, 0x8330,
	0x7BC7, 0x6A4E, 0x58D5, 0x495C, 0x3DE3, 0x2C6A, 0x1EF1, 0x0F78,
}

func crcCheckCode(data []byte) uint16 {
	var crc16 uint16
	crc16 = 0xffff
	for _, v := range data {
		crc16 = (crc16 >> 8) ^ mbTable[(crc16^uint16(v))&0xff]
	}
	return ^crc16 & 0xffff
}

//bytesCombine 多个[]byte数组合并成一个[]byte
func bytesCombine(pBytes ...[]byte) []byte {
	return bytes.Join(pBytes, []byte(""))
}
