package core

import (
	"encoding/json"
	"fmt"
	"github.com/syndtr/goleveldb/leveldb"
	"strings"
	"time"
	//"github.com/fjl/go-couchdb"
	"log"
	"os"
)

type Blockchain struct {
	Channel uint8    `json:"C"`
	Blocks  []*Block `json:"B"`
	TmpInfs []*Inf   `json:"tI"`
}

var Db, _ = leveldb.OpenFile("./db", nil)

//创建一个新的区块链
func NewBlockchain(chainId uint8) *Blockchain {
	genesisBlock := GenerateGenesisBlock()
	blockchain := Blockchain{}
	blockchain.Channel = chainId
	blockchain.AppendBlock(&genesisBlock)
	return &blockchain
}

// 添加新区块&*&*&*&*&*+多链支持
func (bc *Blockchain) AppendBlock(newBlock *Block) {
	//fmt.Println("append",bc.Channel," ",len(bc.Blocks)," ",newBlock.Index)
	if len(bc.Blocks) == 0 {
		openFile, e := os.OpenFile("./Ledger"+string('0'+bc.Channel)+"", os.O_RDWR|os.O_TRUNC, 777)
		if e != nil {
			fmt.Println(e)
		}
		_, _ = openFile.Write(BlockToBytes(newBlock))
		_ = openFile.Close()
		openIndex, e := os.OpenFile("./Index"+string('0'+bc.Channel)+"", os.O_RDWR|os.O_TRUNC, 777)
		if e != nil {
			fmt.Println(e)
		}
		Offset, _ := openIndex.Seek(0, 2)
		_, _ = openIndex.Write(Int64ToBytes(Offset))

		_ = openIndex.Close()
		bc.Blocks = append(bc.Blocks, newBlock)
		return
	}
	if isValid(*newBlock, *bc.Blocks[len(bc.Blocks)-1]) {
		openIndex, e := os.OpenFile("./Index"+string('0'+bc.Channel)+"", os.O_RDWR|os.O_APPEND, 777)
		if e != nil {
			fmt.Println(e)
		}
		openFile, e := os.OpenFile("./Ledger"+string('0'+bc.Channel)+"", os.O_RDWR|os.O_APPEND, 777)
		if e != nil {
			fmt.Println(e)
		}
		Offset, _ := openFile.Seek(0, 2)
		_, _ = openIndex.Write(Int64ToBytes(Offset))

		_, _ = openFile.Write(BlockToBytes(newBlock))

		_ = openIndex.Close()
		_ = openFile.Close()

		bc.Blocks = append(bc.Blocks, newBlock)
	} else {
		log.Fatal("invalid block")
	}
}

// 打印链上信息
//PRINT the information in BlockChain
func (bc *Blockchain) Print() {
	fmt.Printf("CHAIN" + string('0'+bc.Channel) + " \n")
	for _, block := range bc.Blocks {
		fmt.Printf("{ \n")
		fmt.Printf("	Index : %d \n", block.Index)
		fmt.Printf("	Prev Hash : %s \n", block.PrevBlockHash)
		fmt.Printf("	Curr Hash : %s \n", block.Hash)
		fmt.Printf("	Data : \n")
		for i, Inf := range block.Data {
			fmt.Printf("\tInf%d : ", i)
			fmt.Printf("IP: %d\tChannel: %d\n\t\tTime: %d\tIP: %s\n\t\tFunc: %d\t  Trans: %d\tOp: %d\n", Inf.ID, Inf.Channel, Inf.Time, InetNtoA(Inf.IP), Inf.FunCode, Inf.Trans, Inf.Op)
		}
		fmt.Printf("\n	Timestamp : %d \n", block.Timestamp)
		fmt.Printf("} \n")
	}
}

//检验新区块
func isValid(newBlock Block, oldBlock Block) bool {
	if newBlock.Index-1 != oldBlock.Index {
		//fmt.Println(newBlock.Index,oldBlock.Index,"1")
		return false
	}
	if newBlock.PrevBlockHash != oldBlock.Hash {
		//fmt.Println("2")
		return false
	}
	if CalculateHash(newBlock) != newBlock.Hash {
		//fmt.Println("3")
		return false
	}
	return true
}

//拓扑更新
func (bc *Blockchain) InfoUpdate(data string) int64 {
	newBlock := Block{}
	_ = json.Unmarshal([]byte(data), &newBlock)
	bc.AppendBlock(&newBlock)
	for _, ins := range newBlock.Data {
		key := append(Int64ToBytes(ins.IP), Int64ToBytes(ins.Time)...)
		value := append(Int64ToBytes(ins.Time), ins.FunCode, ins.Trans, ins.Op)
		//fmt.Println("store:",key,value)
		_ = Db.Put(key, value, nil)
	}
	//iter := Db.NewIterator(nil, nil)
	return newBlock.Index
}
func (bc *Blockchain) BlockGen(data []*Inf) (int64, []byte) {

	fmt.Printf("%d Purpose Info\n", time.Now().UnixNano()/1000000%100000)
	preBlock := bc.Blocks[len(bc.Blocks)-1]
	newBlock, _ := json.Marshal(GenerateNewBlock(*preBlock, data))

	return preBlock.Index, newBlock
}
func (bc *Blockchain) ExcuteQuery(key string) []byte {
	t := make([]byte, 0)
	iter := Db.NewIterator(nil, nil)
	//fmt.Println("seek:",Int64ToBytes(InetAtoN(key)))
	ok := iter.Seek(Int64ToBytes(InetAtoN(key)))
	if ok && strings.Contains(string(iter.Key()), string(Int64ToBytes(InetAtoN(key)))) {
		t = append(t, iter.Value()...)
		//fmt.Println("found",iter.Value())
	}
	ok = iter.Next()
	for ; ok; ok = iter.Next() {
		if !strings.Contains(string(iter.Key()), string(Int64ToBytes(InetAtoN(key)))) {
			break
		}
		t = append(t, byte('@'))
		t = append(t, iter.Value()...)
		//fmt.Println("found",iter.Value())
	}
	return t
}
func (bc *Blockchain) KeyUpdate() {
	for _, block := range bc.Blocks {
		for _, Inf := range block.Data {
			key := append(Int64ToBytes(Inf.IP), Int64ToBytes(Inf.Time)...)
			value := append(Int64ToBytes(Inf.Time), Inf.FunCode, Inf.Trans, Inf.Op)
			_ = Db.Put(key, value, nil)
		}
	}
}

//异常行为处理模块&*&*&*&*+新型异常
/* Report(Link *Link,ErrorType int){
	Etime:=time.Now().Format("15:04:05")
	var info string
	if Link!=nil{
		info =InetNtoA(Link.IP_A)+"-"+InetNtoA(Link.IP_B)
	}
	switch ErrorType {
	case 0:{
		fmt.Println(Etime,"冲突:",info)
	}
	case 1:{
		fmt.Println(Etime,"无效的删除:",info)
	}
	case 2:{
		fmt.Println(Etime,"超时:",info)
	}
	case 3:{
		fmt.Println(Etime,"配置无效")
	}
	case 4:{
		fmt.Println(Etime,"接收配置格式错误")
	}
	}

}*/
