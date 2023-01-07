# Semi-Logchain
半中心化的区块链系统原型，采用了多通道设计，各通道间账本独立，通道内节点间基于PBFT机制达成共识

## Usage

生成可执行程序后放在不同文件夹下作为不同节点，调整conf文件以配置节点地址等

>conf文件中各字段含义
>
>localAddress  本地地址（IP)
>
>rmAddress 初始主节点位置（SCMS的IP）
>
>isInit  初始化标识
>
>onChain 所在的通道列表

0.网络启动前，各节点分别执行keyGen进行身份认证凭据的生成和本地保存；

1.主节点率先启动，之后各初始节点完成启动，以执行初始化阶段；
启动方法为直接运行可执行程序：
```Semi-Logchain.exe```

2.如需上传信息，使用upload方法，如需查询已在链上的信息则使用query方法；

3.后续节点不设初始化标签直接启动，程序将自动访问主节点并请求加入配置文件中设置的通道。



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

### 宕机后恢复

重新启动时,```isInit```应当置为false，在启动之初的配置读取阶段会首先读取相应配置，检测到初始化标签为false时进行rollback。
