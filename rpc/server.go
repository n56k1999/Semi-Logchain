package main

import (
	"../core"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

//Num of chains
const CN = 2

type MapConfig struct {
	LocalAddress string             `json:"localAddress"`
	MaAddr       string             `json:"rmAddress"`
	NodeTables   map[uint8][]string `json:"nodeTables"`
	INodeTables  map[uint8][]string `json:"inodeTables"`
	RollbackInfo []bool             `json:"rollBackInfo"`
	IsMain       bool               `json:"isMain"`
	IsInit       bool               `json:"isInit"`
	OnChain      []bool             `json:"onChain"`
}

var config MapConfig

var MaxTimeout = 400 //max time before Gen new block(ms)
var MaxBaSize = 3    //max batchSize per block
var Plock = new(sync.RWMutex)
var NChain = [CN]*core.Blockchain{core.NewBlockchain(0), core.NewBlockchain(1)}

type nodeInfo struct {
	//节点名称
	id string
	//路径
	url string
	//http响应
}

var tmpInfs [CN][]*core.Inf
var timer [CN]bool

//用于记录正常响应的好的节点
var authenticationNodeMap = [CN]map[string]bool{make(map[string]bool), make(map[string]bool)}
var commitMap = [CN]map[string]int{make(map[string]int), make(map[string]int)}
var authenticationSuccess = [CN]map[string]bool{make(map[string]bool), make(map[string]bool)}
var PreIndex = [CN]int64{0, 0}

//args
func main() {
	//PHASE config
	toConfig()
	//Select Operation
	if len(os.Args) >= 2 {
		fmt.Println(os.Args)
		switch os.Args[1] {
		case "upload":
			fileBytes, err := ioutil.ReadFile("info")
			if err != nil {
				panic(err)
			}
			lines := strings.Split(string(fileBytes), "\n")
			for _, inf := range lines {
				Upload(config.MaAddr, config.LocalAddress+"$"+inf)
			}
			println("upload complete", time.Now().UnixNano()/100000)
		case "query":
			var toC, key0 string
			print("Input the channel to query on:")
			_, _ = fmt.Scanln(&toC)
			qcid, _ := hex.DecodeString(toC[0:2])
			if qcid[0] >= CN {
				fmt.Println("no such channel")
				return
			}
			print("Input the key to query:")
			_, _ = fmt.Scanln(&key0)
			time.Sleep(100 * time.Millisecond)
			result := QueryInfoByIP(config.LocalAddress, config.LocalAddress+"$"+toC+"$"+key0)
			fmt.Println("---", key0, " on chain No.", toC, "---")
			if len(result) == 0 {
				fmt.Println("no result")
				return
			}
			inStr := strings.TrimSpace(string(result))
			cInputs := strings.Split(inStr, "@")
			for i := range cInputs {
				infTime := binary.BigEndian.Uint64([]byte(cInputs[i][0:8]))
				println(i, " time:", infTime, " func:", cInputs[i][8])
			}
		case "keyGen":
			if len(os.Args) < 3 {
				fmt.Println("usage0: keyGen [seedInfo]")
			}
			core.GenRsaKeys()
		default:

		}

	} else { //init service
		//var file = "/home/n56k1998/ServiceLog0"
		//var LogFile, _ = os.OpenFile(file, os.O_RDWR|os.O_TRUNC, 777)
		//_, _ = LogFile.Write([]byte("0say"))

		//INIT
		node := nodeInfo{id: "-1", url: ""}
		node.url = config.LocalAddress
		if config.IsMain {
			for cid := 0; cid < CN; cid++ {
				config.NodeTables[uint8(cid)] = append(config.NodeTables[uint8(cid)], node.url)
			}
			if !config.IsInit {
				fmt.Println("Rollback...")
				config.NodeTables = config.INodeTables
				node.broadcast("00", "/rollback")
				node.broadcast("01", "/rollback")
			}
		} else {
			for cid, on := range config.OnChain {
				if on {
					fmt.Println("Connect to SCMS...")
					if !node.SECtoMain(cid) {
						fmt.Println("Fail,exiting")
						return
					}
				}
			}
		}
		//Service Start
		ServerSocket(node)
		defer core.Db.Close()
	}

	//defer LogFile.Close()

}

//&*&*&*&*&*+链操作
func (node *nodeInfo) Upload(data string) {
	if len(data) > 0 && config.IsMain {
		//有值
		Inf := core.Inf{}
		if !core.DataValidCheck(data, &Inf) {
			return
		}
		//timer on(for every chain)
		if !timer[Inf.Channel] {
			timer[Inf.Channel] = true
			go func() {
				cid := Inf.Channel
				cids := data[4:6]
				select {
				case <-time.After(time.Millisecond * time.Duration(MaxTimeout)):
					Plock.Lock()
					if timer[cid] {
						fmt.Println("SCMS START CONSENSUS(PBFT,TO)")
						timer[cid] = false
						var bcData []byte
						PreIndex[cid], bcData = NChain[cid].BlockGen(tmpInfs[cid])
						tmpInfs[cid] = []*core.Inf{}
						block := cids + "#" + string(bcData)
						node.broadcast(block, "/prePrepare")
					}
					Plock.Unlock()
					return
				}
			}()
		}
		Plock.Lock()
		tmpInfs[Inf.Channel] = append(tmpInfs[Inf.Channel], &Inf)
		if len(tmpInfs[Inf.Channel]) >= MaxBaSize {
			fmt.Println("SCMS START CONSENSUS(PBFT,TO)")
			timer[Inf.Channel] = false
			var bcData []byte
			PreIndex[Inf.Channel], bcData = NChain[Inf.Channel].BlockGen(tmpInfs[Inf.Channel])
			tmpInfs[Inf.Channel] = []*core.Inf{}
			block := data[4:6] + "#" + string(bcData)
			node.broadcast(block, "/prePrepare")

		}
		Plock.Unlock()
	}
	return
}

//入口
/*func (node *nodeInfo) Entry(data string){
	fmt.Printf("Config Received %s\n", data)
	//主节点
	//if (time.Now().Unix()%(int64(len(config.NodeTables[0]))*60))/60==int64(node.id[0]-'0'){
	if node.id=="0"{
		fmt.Println("主节点发布区块","开始PBFT共识")
		cid,_:=hex.DecodeString(data[4:6])
		block:=data[4:6]+"$"+string(NChain[cid[0]].BlockGen(data))
		node.broadcast(block, "/prePrepare")
	}
	return
}*/

//共识确认-Preprepare、prepare、commit
//Preprepare
func (node *nodeInfo) PrePrepare(data string) {
	//request.ParseForm()
	fmt.Println("Block Received") //,commitN)//, request.Form["data"][0])
	//若数据有值，则进行广播
	if len(data) > 0 {
		//有值
		node.broadcast(data, "/prepare")
	}
}

//prepare
func (node *nodeInfo) Prepare(data string, url string) {
	cid, _ := hex.DecodeString(data[0:2])
	//fmt.Println("接收到子节点的广播", url,cid)
	if (len(data) > 0) && node.authentication(url, data[15:25], cid[0]) {
		//进行拜占庭校验
		node.broadcast(data, "/commit")
	}
}
func (node *nodeInfo) authentication(nul string, dig string, cid uint8) bool {
	//第一次进去
	if !authenticationSuccess[cid][dig] {
		//运行的正常节点的判断
		if (len(nul)) > 0 {
			//证明节点是OK的
			authenticationNodeMap[cid][nul] = true
			//如果有两个节点正确返回了结果，可以发送
			if len(authenticationNodeMap[cid]) > len(config.NodeTables[cid])/3 {
				//(n-1)/3的容错性
				//进入commit阶段
				authenticationSuccess[cid][dig] = true
				return true
			}
		}
	}
	return false
}

//commit
func (node *nodeInfo) Commit(data string) {
	cid, _ := hex.DecodeString(data[0:2])
	commitMap[cid[0]][data[15:25]] += 1
	//fmt.Println( "Cn:",commitN)
	if commitMap[cid[0]][data[15:25]] == len(config.NodeTables[cid[0]])-1 {
		if authenticationSuccess[cid[0]][data[15:25]] {
			delete(commitMap[cid[0]], data[15:25])
			authenticationNodeMap[cid[0]] = make(map[string]bool)
		}
		fmt.Println("upload completed", time.Now().UnixNano()/100000)
		fmt.Println("CONSENSUS COMPLETED,Chain", cid[0], "COMMIT#", data[9:11])
		//拓扑更新&*&*&*&*+信息更新
		//VERSION check
		//ver,_:=dec.DecodeString(data[4:6])
		//println(data)
		ver := NChain[cid[0]].InfoUpdate(data[3:])
		if PreIndex[cid[0]] != ver-1 {
			go func() {
				time.Sleep(30 * time.Millisecond)
				node.SECtoMain(int(cid[0]))
			}()
		} else {
			PreIndex[cid[0]] = ver
		}
		//PRINT the information in BlockChain
		/*if(*Anew==1){
			NChain[0].Print()
		}*/
	}

}

//广播
func (node *nodeInfo) broadcast(msg string, path string) {
	if path == "/prepare" || path == "/prePrepare" || path == "/commit" {
		fmt.Println("Broadcast", path)
	}
	//cid-Channel
	//遍历所有节点进行广播
	cid, _ := hex.DecodeString(msg[0:2])
	for _, url := range config.NodeTables[cid[0]] { //nodeTable[cid[0]]
		//判断是否是自己，若为自己，则跳出循环
		if url == node.url {
			continue
		}
		ClientSocket(url, path+"$"+node.url+"$"+msg)
	}
}
func (node *nodeInfo) NodeListChange(idS byte, url string, Pkey string) {
	//cid,_:=hex.DecodeString(idS[0:2])
	wFlag := false
	if len(config.NodeTables[idS]) == 0 {
		wFlag = true
	} else {
		for i, addr := range config.NodeTables[idS] {
			if addr == url {
				break
			} else if i == (len(config.NodeTables[idS]) - 1) { //no info exists, update
				wFlag = true
			}
		}
	}
	//SYNC
	if wFlag {
		config.NodeTables[idS] = append(config.NodeTables[idS], url)
		wFlag = false
		//if config.IsMain {
		writePub(url, Pkey)
		//}
	}
}
func writePub(url string, Pkey string) {
	PFile, err := os.OpenFile("../ki/"+url+"PUB", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		log.Panic(err)
	}
	_, _ = PFile.Write([]byte(Pkey))
	PFile.Close()
}

func toConfig() {
	in, e := ioutil.ReadFile("./conf")
	if e != nil {
		fmt.Println(e)
	}

	err := json.Unmarshal(in, &config)
	if err != nil {
		fmt.Println(err)
	}
	if config.NodeTables == nil {
		config.NodeTables = make(map[uint8][]string)
	}
}

func (node *nodeInfo) SECtoMain(cid int) bool {
	i := 0
	Pkey, err := ioutil.ReadFile("./RSA_PUB")
	if err != nil {
		log.Panic(err)
	}
	for {
		c, _ := net.Dial("tcp", config.MaAddr)
		if c != nil || i > 20 {
			if c != nil {
				CHandler(c, "/join0$"+node.url+"$"+"0"+strconv.Itoa(cid)+"$"+string(Pkey)[:len(Pkey)-1]+"$**")
				_ = c.Close()
				PreIndex[cid] = NChain[cid].Blocks[len(NChain[cid].Blocks)-1].Index
				return true
			}
			return false
		}
		i++
		time.Sleep(500 * time.Millisecond)
	}
}
