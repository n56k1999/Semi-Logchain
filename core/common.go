package core

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math/big"
	"net"
	"regexp"
	"strconv"
	"strings"
)

//内容有效性检测
//(0a,01,1601778000,192.168.0.1,0f,07,09)
func DataValidCheck(blockData string, a *Inf) (b bool) {
	if strings.Count(blockData, string(',')) != 6 {
		return false
	}
	if blockData[0] != '(' || blockData[len(blockData)-1] != ')' {
		return false
	}
	reg := regexp.MustCompile(`[\w.]+`)
	arr0 := reg.FindAllString(blockData, -1)
	//fmt.Println(arr0)
	if len(arr0) != 7 {
		return false
	}
	aH, _ := hex.DecodeString(arr0[0])
	a.ID = aH[0]
	aH, _ = hex.DecodeString(arr0[1])
	a.Channel = aH[0]
	a.Time, _ = strconv.ParseInt(arr0[2], 10, 64)
	a.IP = InetAtoN(arr0[3])
	aH, _ = hex.DecodeString(arr0[4])
	a.FunCode = aH[0]
	aH, _ = hex.DecodeString(arr0[5])
	a.Trans = aH[0]
	aH, _ = hex.DecodeString(arr0[6])
	a.Op = aH[0]
	return true
}

func BlockToBytes(b *Block) []byte {
	var buffer bytes.Buffer
	buffer.Write(Int64ToBytes(b.Index))
	buffer.Write(Int64ToBytes(b.Timestamp))
	buffer.Write([]byte(b.PrevBlockHash))
	buffer.Write([]byte(b.Hash))
	for _, data := range b.Data {
		buffer.Write([]byte{data.ID})
		buffer.Write([]byte{data.Channel})
		buffer.Write(Int64ToBytes(data.Time))
		buffer.Write(Int64ToBytes(data.IP))
		buffer.Write([]byte{data.FunCode})
		buffer.Write([]byte{data.Trans})
		buffer.Write([]byte{data.Op})
	}
	return buffer.Bytes()
}
func Int64ToBytes(i int64) []byte {
	var buf = make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(i))
	return buf
}
func InetAtoN(ip string) int64 {
	ret := big.NewInt(0)
	ret.SetBytes(net.ParseIP(ip).To4())
	return ret.Int64()
}
func InetNtoA(ip int64) string {
	return fmt.Sprintf("%d.%d.%d.%d",
		byte(ip>>24), byte(ip>>16), byte(ip>>8), byte(ip))
}

/*func IdCheck(Tlink *Inf ,Nlink *Inf) bool{
	if Tlink.IP_B!=Nlink.IP_B ||Tlink.IP_A!=Nlink.IP_A {
		return false
	}
	return true
}
func PartCheck(Tlink *Inf ,Nlink *Inf) bool{
	if Tlink.NetAddr_A!=Nlink.NetAddr_A || Tlink.Linktype!=Nlink.Linktype|| Tlink.NetAddr_B!= Nlink.NetAddr_B {
		return false
	}
	return true
}
func InetNtoA(ip int64) string {
	return fmt.Sprintf("%d.%d.%d.%d",
		byte(ip>>24), byte(ip>>16), byte(ip>>8), byte(ip))
}
func InftoStr(str *Inf)string{
	return ","+string(str.Linktype+48)+ "," + InetNtoA(str.IP_A) + "," + hex.EncodeToString([]byte{str.NetAddr_A}) + "," + hex.EncodeToString([]byte{str.Port_A}) + "," + InetNtoA(str.IP_B) + "," + hex.EncodeToString([]byte{str.NetAddr_B}) + "," + hex.EncodeToString([]byte{str.Port_B}) + ")"
}*/
