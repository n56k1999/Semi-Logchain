package main

import (
	"../core"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"strings"
	"sync"
	"time"
)

var lock = new(sync.RWMutex)

func SHandler(c net.Conn, node nodeInfo) {
	if c == nil {
		log.Panic("Invalid socket")
	}
	buf := make([]byte, 4096)
	cnt, err := c.Read(buf)
	if cnt == 0 || err != nil {
		_ = c.Close()
		return
	}
	cInputs := strings.Split(string(buf[0:cnt-64]), "$")
	fCommand := cInputs[0]
	fmt.Println("method->" + fCommand)
	rmAddr := cInputs[1]
	//on init,no CA to use
	if fCommand == "/join0" && buf[cnt-65] == '*' && buf[cnt-66] == '*' {
		if config.IsMain {
			fmt.Println("SYNC...")
			_, _ = c.Write(lockedP(cInputs, node))
			node.broadcast(cInputs[2]+"$"+cInputs[1]+"$"+cInputs[3], "/join1")
			_ = c.Close()
		}
		return
	} else {
		var PK []byte
		if rmAddr == config.LocalAddress {
			PK = core.GetPubKey()
		} else {
			PK = core.GetPubKey(rmAddr)
			if PK == nil {
				PK = updatePub(rmAddr)
			}
		}
		//fmt.Println("very:",rmAddr,string(PK))
		if !core.RsaVerySignWithSha256(buf[:cnt-64], buf[cnt-64:cnt], PK) {
			fmt.Printf("Authtication Fail\n")
			return
		}
	}
	//Authenticated
	//inStr := strings.TrimSpace(string(buf[0:cnt-64]))
	switch fCommand {
	case "/upload":
		fmt.Println("upload start", time.Now().UnixNano()/100000)
		node.Upload(cInputs[2])
		_, _ = c.Write([]byte("re:upload\n"))
	case "/entry":
		//node.Entry(cInputs[1])
		_, _ = c.Write([]byte("re: entry\n"))
	case "/prePrepare":
		node.PrePrepare(cInputs[2])
		_, _ = c.Write([]byte("re: preprepare\n"))
	case "/prepare":
		node.Prepare(cInputs[2], cInputs[1])
		_, _ = c.Write([]byte("re: prepare\n"))
	case "/commit":
		node.Commit(cInputs[2])
		_, _ = c.Write([]byte("re: commit\n"))
	case "/join1":
		//change the nodes in nodeTable
		//synchronization
		if !config.IsMain {
			cid, _ := hex.DecodeString(cInputs[2][0:2])
			node.NodeListChange(cid[0], cInputs[3], cInputs[4])
			_, _ = c.Write([]byte("re: positive\n"))
			//fmt.Println(config.NodeTables)
		}
	case "/indexPub":
		_, _ = c.Write(append(core.GetPubKey(cInputs[2]), core.RsaSignWithSha256(core.GetPubKey(cInputs[2]), core.GetPrivKey())...))
	case "/rollback":
		cid, _ := hex.DecodeString(cInputs[2][0:2])
		if config.RollbackInfo[cid[0]] {
			ledger, _ := json.Marshal(*NChain[cid[0]])
			table, _ := json.Marshal(config.NodeTables[cid[0]])
			str := append([]byte(cInputs[2][0:2]), uint8(len(table)>>8))
			str = append(str, uint8(len(table)))
			str = append(str, table...)
			str = append(str, ledger...)
			_, _ = c.Write(append(str, core.RsaSignWithSha256(str, core.GetPrivKey())...))
		}
	case "/Print":
		cid, _ := hex.DecodeString(cInputs[2])
		NChain[cid[0]].Print()
	case "/query":
		cid, _ := hex.DecodeString(cInputs[2])
		result := NChain[cid[0]].ExcuteQuery(cInputs[3])
		_, _ = c.Write(result)
	default:
		_, _ = c.Write([]byte("done"))
	}
	_ = c.Close()

}

//Locked code area
func lockedP(cInputs []string, node nodeInfo) []byte {
	lock.Lock()
	cid, _ := hex.DecodeString(cInputs[2][0:2])
	node.NodeListChange(cid[0], cInputs[1], cInputs[3])
	ledger, _ := json.Marshal(*NChain[cid[0]])
	table, _ := json.Marshal(config.NodeTables[cid[0]])
	str := append([]byte(cInputs[2][0:2]), uint8(len(table)>>8))
	str = append(str, uint8(len(table)))
	str = append(str, table...)
	str = append(str, ledger...)
	str = append(str, core.RsaSignWithSha256(str, core.GetPrivKey())...)
	lock.Unlock()
	return str
}
func CHandler(c net.Conn, data string) {
	//缓存 conn 中的数据
	buf := make([]byte, 1024)
	input := data
	//input := strings.TrimSpace(data)
	if len(input) != 0 {
		_, _ = c.Write(append([]byte(input), core.RsaSignWithSha256([]byte(input), core.GetPrivKey())...))
		//服务器端返回的数据写入空buf
		cnt, err := c.Read(buf)
		if err != nil {
			fmt.Printf("read0 fail %s\n", err)
			return
		}
		if cnt-64 > 25 { //cid+lenT+table+ledger
			//rmAddr:=strings.Split(c.RemoteAddr().String(), ":")
			rAddr := c.RemoteAddr().String()
			cPub := core.GetPubKey(rAddr)
			if cPub == nil {
				cPub = updatePub(rAddr)
			}
			if !core.RsaVerySignWithSha256(buf[:cnt-64], buf[cnt-64:cnt], cPub) {
				fmt.Printf("Authtication Fail\n")
				return
			}
			cid, _ := hex.DecodeString(string(buf[0:2]))
			NChain[cid[0]] = &core.Blockchain{}
			var lenT = int(buf[2])
			lenT = lenT<<8 + int(buf[3])
			var parT []string
			err = json.Unmarshal(buf[4:4+lenT], &parT)
			//fmt.Println(err)
			config.NodeTables[cid[0]] = parT
			err = json.Unmarshal(buf[4+lenT:cnt-64], &NChain[cid[0]])
			NChain[cid[0]].KeyUpdate()
			fmt.Println("Chain", cid[0], " is ready")
		}
		//fmt.Print(string(buf[0:cnt]))
	}
}
func updatePub(rAddr string) []byte {
	cInput := "/indexPub$" + config.LocalAddress + "$" + rAddr
	cm, _ := net.Dial("tcp", config.MaAddr)
	cbuf := make([]byte, 1024)
	_, _ = cm.Write(append([]byte(cInput), core.RsaSignWithSha256([]byte(cInput), core.GetPrivKey())...))
	ccnt, _ := cm.Read(cbuf)
	if !core.RsaVerySignWithSha256(cbuf[:ccnt-64], cbuf[ccnt-64:ccnt], core.GetPubKey(config.MaAddr)) {
		fmt.Printf("Authtication Fail\n")
		return nil
	}
	writePub(rAddr, string(cbuf[:ccnt-64]))
	return cbuf[:ccnt-64]
}

//开启serverSocket
func ServerSocket(node nodeInfo) {
	//1.监听端口
	server, err := net.Listen("tcp", node.url)
	if err != nil {
		fmt.Println("unable to establish the connect")
	}
	fmt.Println("Initializing Service...,", node.url)
	for {
		//2.接收来自 client 的连接,会阻塞
		conn, err := server.Accept()
		//fmt.Println("outLink")
		if err != nil {
			fmt.Println("Connect Error")
		}
		//并发模式 接收来自客户端的连接请求
		go SHandler(conn, node)
	}
}
func ClientSocket(svAddr string, data string) {
	conn, err := net.Dial("tcp", svAddr)
	if err != nil {
		fmt.Println("ucTo ", svAddr)
		return
	}
	CHandler(conn, data)
}
func Upload(svAddr string, info string) {
	conn, err := net.Dial("tcp", svAddr)
	if err != nil {
		fmt.Println("Connect Fail")
		return
	}
	buf := make([]byte, 1024)
	input := "/upload$" + info
	//客户端请求数据写入 conn，并添加校验
	if len(input) != 0 {
		_, _ = conn.Write(append([]byte(input), core.RsaSignWithSha256([]byte(input), core.GetPrivKey())...))
		//服务器端返回的数据写入空buf
		cnt, err := conn.Read(buf)
		if err != nil {
			fmt.Printf("READ FAIL%s\n", err)
		}
		//回显服务器端回传的信息
		fmt.Print(string(buf[0:cnt]))
	}
}
func _(svAddr string, info string) {
	conn, err := net.Dial("tcp", svAddr)
	if err != nil {
		fmt.Println("Connect Fail")
		return
	}
	buf := make([]byte, 1024)
	input := "/Print$" + info
	_, _ = conn.Write([]byte(input))
	_, _ = conn.Read(buf)
}
func QueryInfoByIP(svAddr string, key string) []byte {
	c, err := net.Dial("tcp", svAddr)
	if err != nil {
		fmt.Println("Connect Fail")
		return nil
	}
	buf := make([]byte, 1024)
	input := "/query$" + key
	//客户端请求数据写入 conn，并传输
	if len(input) != 0 {
		_, _ = c.Write(append([]byte(input), core.RsaSignWithSha256([]byte(input), core.GetPrivKey())...))
		//服务器端返回的数据写入空buf
		cnt, err := c.Read(buf)
		if err != nil {
			fmt.Printf("Client cTo Fail%s\n", err)
		}
		//回显服务器端回传的信息
		//fmt.Print(string(buf[0:cnt]))
		return buf[0:cnt]
	}
	return nil
}
