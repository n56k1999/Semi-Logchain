# Semi-Logchain
半中心化的区块链系统原型，采用了多通道设计，各通道间账本独立，通道内节点间基于PBFT机制达成共识

## Usage

生成可执行程序后放在不同文件夹下作为不同节点，调整conf文件以配置节点地址等

### 秘钥生成

```Semi-Logchain.exe KeyGen```

### 数据上传

```Semi-Logchain.exe Upload```
>
>程序将按行读取同目录下的info文件内容并发布

### 数据查询

```
Semi-Logchain.exe Query
Input the channel to query on: 00
Input the key to query: 0a
```
