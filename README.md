# NewChainAPIExpress

## 目标

1，简化NewChain客户端的开发，并提高运行速度。

2，部分资源受限客户端难以实现计算签名Recovery ID，需要服务器端协助。

## 需求描述

### 获取基础数据的API： newton_getBaseInfo
1，用户提供账户地址，服务器返回一次性获取Nonce、Gas Price、Chain ID、当前余额。

### 客户提交TX相关信息的API： newton_sendTransaction
1. 客户端输入数据包括：
    1. 未签名的RAW TX，RLP HEX格式。
    2. 签名数据，HEX格式。
    3. FROM地址。
2. 服务端需要根据用户提供的数据，计算出Recovery ID，并组装为签名后的RAW TX。
3. 服务器端将上述RAW TX发送到NewChain RPC。
4. 服务器端返回给客户端TX HASH，如果有错误，返回错误代码。
5. 提供wait参数：
    1. 如果客户端设置wait为0，则服务端验证客户提供的数据合法后，
            即可返回TX HASH，然后服务器端再提交TX到NewChain RPC。
    2. 如果客户端设置wait为1，则服务器端提交TX到NewChain RPC后即可返回。
    3. 如果客户端设置wait为2，则服务器端需等待TX被确认后才可返回。

### 客户提交TX相关信息的API： newton_sendRawTransaction
1. 客户端输入数据包括（同eth_sendRawTransaction）：
    1. 签名后的RAW TX，RLP HEX格式。
2. 服务器端将上述RAW TX发送到NewChain RPC。
3. 服务器端返回给客户端TX HASH，如果有错误，返回错误代码。
4. 提供wait参数：
    1. 如果客户端设置wait为0，则服务端验证客户提供的数据合法后，
        即可返回TX HASH，然后服务器端再提交TX到NewChain RPC。
    2. 如果客户端设置wait为1，则服务器端提交TX到NewChain RPC后即可返回。
    3. 如果客户端设置wait为2，则服务器端需等待TX被确认后才可返回。


### 到账通知
提供三个级别的mqtt到账通知。  
* 0: 收到合法数据。
* 1: 合法的tx提交到NewChain。
* 2: tx被确认至少1个区块。


### 备注
1. 客户端根据实际情况通过get_base_info同步基础信息。
2. 由于目前GAS Price非常稳定，客户端可以设置为固定值，无需向服务端询问。
3. 客户端可以将Gas Limit设置为一个较大的值，无需事先评估。
4. 通讯使用HTTP POST JSON格式数据进行通讯，兼容jsonrpc 2.0和NewChain RPC。
5. 最终效果：客户端通过一次通讯即可完成一次交易，整体时间低于0.5秒。

## 安装

安装 newchain-api-express 程序：

```bash
git clone https://github.com/newtonproject/newchain-api-express.git && cd newchain-api-express && make install
```

服务器端需要配置MQTT服务，可参考[MQTT](http://mqtt.org/)或者使用[AWS MQ](https://aws.amazon.com/amazon-mq)


## API

### newton_getBaseInfo

获取基础数据

* 请求参数
    * address: 用户地址, hex格式
* 返回参数
    * JSON结构体
        * networkID: ChainID
        * gasPrice: 当前Gas费用
        * nonceLatest: 地址address的 latest nonce
        * noncePending: 地址address的 pending nonce
        * balance: 地址address的latest余额
* 示例

```
// Request
curl -X POST --data '{"jsonrpc":"2.0","method":"newton_getBaseInfo","params":{"address":"0x97549E368AcaFdCAE786BB93D98379f1D1561a29"},"id":1}' -H "Content-Type: application/json" http://127.0.0.1:8888

// Result
{
    "jsonrpc":"2.0",
    "id":1,
    "result":{
        "nonceLatest": "0x543",
        "noncePending": "0x543",
        "gasPrice":"0x64",
        "networkID":1007,
        "balance":"0x32b6fbe3b559ae26fceaf1"
    }
}
```


### newton_sendRawTransaction

客户端提交签名后的RawTransaction

* 请求参数
    * Transaction结构体
        * tx: 签名后的RawTransaction，RLP HEX格式
        * wait: 0,1,2，需要wait的参数
* 返回参数
    * 交易Hash

```
// Request
curl -X POST --data '{"jsonrpc":"2.0","method":"newton_sendRawTransaction","params":{"tx":"0xf86b8204de648252089497549e368acafdcae786bb93d98379f1d1561a29880de0b6b3a764000080820801a04177f15eec3c930644f4964feaf7b73c6b4d28bb59394ec4c70e3d8d6812f9f4a03fea89e167ca55787c62ee992f857457f2f3b5a36d7e452758654fc5dcdfe1e5","wait":0},"id":67}'  -H "Content-Type: application/json" http://127.0.0.1:8888

// Result
{
    "jsonrpc":"2.0",
    "id":67,
    "result":"0x85ea238671582e93bbcfffa94b09c00cee35d7ae46a38e547f5f234ebcbd0dc1"
}
```


### newton_sendTransaction

客户端提交签名后但未组装的RawTransaction

* 请求参数
    * Transaction结构体
        * message: 未签名的RAW TX，RLP HEX格式
        * tx: 签名结果，HEX格式
        * from: 发送者地址，HEX格式
        * wait: 0,1,2，需要wait的参数
* 返回参数
      * 交易Hash
* 示例

```
// Request
curl -X POST --data '{"jsonrpc":"2.0","method":"newton_sendTransaction","params":{"from":"0x97549e368acafdcae786bb93d98379f1d1561a29","tx":"0xe98204e0648252089497549e368acafdcae786bb93d98379f1d1561a29880de0b6b3a764000080808080","signature":"0x2bfdd5d619d589e5c3d389affbab514ec3d36fe1e21b42d6e09b059e98d7202a7d3c7a5f0325a72cc17ff5b7d436a6d562f27ff1608bc1df60c7166c81a4a948","wait":1},"id":67}'  -H "Content-Type: application/json" http://127.0.0.1:8888

// Result
{
    "jsonrpc":"2.0",
    "id":67,
    "result":"0xf172da87fc390f9b57205fd5ebb6bf2a716635951dffde28fe93be7ad2ec1b77"
}
```

## Test

### info

```bash
# Get base info of address 0xd639a62be604374ff04af4112a555890bd822a03
newchain-api-express info 0xd639a62be604374ff04af4112a555890bd822a03

# Get base info and save to config file
newchain-api-express info 0xd639a62be604374ff04af4112a555890bd822a03 --update
```

### pay

```bash
# pay 1 NEW to  0x97549e368acafdcae786bb93d98379f1d1561a29
newchain-api-express pay 0x97549e368acafdcae786bb93d98379f1d1561a29 1 --from 0xd639A62Be604374fF04aF4112a555890Bd822a03

# Pay and waiting to be mined
newchain-api-express pay 0x97549e368acafdcae786bb93d98379f1d1561a29 1 --from 0xd639A62Be604374fF04aF4112a555890Bd822a03 --wait 2
```

