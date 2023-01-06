package core

import (
	"crypto/sha256"
	"encoding/hex"
	"strconv"
	"time"
)

type Block struct {
	Index         int64  `json:"No"` // 区块编号
	Timestamp     int64  `json:"T"`  //区块时间戳
	PrevBlockHash string `json:"Ph"` //上一个区块hash
	Hash          string `json:"h"`  //当前区块哈希值
	Data          []*Inf `json:"D"`  //区块数据
}
type Inf struct {
	ID      uint8 `json:"ID"`
	Channel uint8 `json:"Mc"`
	Time    int64 `json:"t"`
	IP      int64 `json:"ip"`
	FunCode uint8 `json:"func"`
	Trans   uint8 `json:"ts"`
	Op      uint8 `json:"op"`
}

var GInf = Inf{48, 0, 1603113000, 10, 5, 1, 11}

// 计算区块hash
func CalculateHash(b Block) string {
	blockData := strconv.FormatInt(b.Index, 10) + strconv.FormatInt(b.Timestamp, 10) + b.PrevBlockHash
	for _, data := range b.Data {
		blockData += strconv.FormatInt(data.IP+data.Time, 10) + string(data.Channel+data.ID+data.FunCode+data.Trans+data.Op)
	}
	hashInByte := sha256.Sum256([]byte(blockData))
	hashInStr := hex.EncodeToString(hashInByte[:])
	return hashInStr
}

// 生成新区块
func GenerateNewBlock(preBlock Block, data []*Inf) Block {
	newBlock := Block{}
	newBlock.Index = preBlock.Index + 1
	newBlock.PrevBlockHash = preBlock.Hash
	if newBlock.Index == 0 {
		newBlock.Timestamp = 1603113000
	} else {
		newBlock.Timestamp = time.Now().Unix()
	}
	newBlock.Data = data
	newBlock.Hash = CalculateHash(newBlock)
	return newBlock
}

//创世区块
func GenerateGenesisBlock() Block {
	preBlock := Block{}
	preBlock.Index = -1
	preBlock.Hash = ""
	var genInf []*Inf
	genInf = append(genInf, &GInf)
	return GenerateNewBlock(preBlock, genInf)
}
