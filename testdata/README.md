# TestData
基于docker-compose一键启动PaddleDTX分布式机器学习所需的节点配置文件及docker-compose.yml文件.

## 一、文件目录说明
- blockchain【区块链网络】: 
    - xchain: 三个节点的xchain网络
        - xchain1/xchain2/xchain3: 区块链节点配置目录，包含各个网络节点账户配置
        - user: 用户安装合约所使用的区块链账户地址，包含用户的助记词、公私钥等信息
        - docker-compose.yml: xchain网络启动所需的配置
    - fabric: 4个peer节点，一个order节点, fabric版本为1.4.8
        - base: order和peer节点启动所需要的yaml文件
        - chaincode: 用户安装链码所需要go mod配置
        - conf: 证书和创世区块生成，以及fabric sdk调用所需要yaml配置文件
        - docker-compose.yml: fabric网络启动后创建通道、安装链码所需的文件

- xdb【去中心化存储服务】：三个数据持有节点，三个存储节点
    - data1: 数据持有节点1配置目录
      - authkeys为节点1授权的客户端公钥目录
      - keys为节点1公私钥配置目录
      - ukeys为节点1生成的客户端私钥对
    - data2: 数据持有节点2配置目录
      - authkeys为节点2授权的客户端公钥目录
      - keys为节点2公私钥配置目录
      - ukeys为节点2生成的客户端公私钥对
    - data3: 数据持有节点3配置目录
      - authkeys为节点3授权的客户端公钥目录
      - keys为节点3公私钥配置目录
      - ukeys为节点3生成的客户端公私钥对
    - storage1: 存储节点1配置目录
    - storage2: 存储节点2配置目录
    - storage3: 存储节点3配置目录
    - docker-compose.yml: 去中心化存储服务启动所需的配置
    
- executor【多方安全计算网络】：三个任务执行节点，一个计算需求方客户端
    - node1: 任务执行节点1配置目录
    - node2: 任务执行节点2配置目录
    - node3: 任务执行节点3配置目录
    - requester：计算需求方配置目录
    - docker-compose.yml: 多方安全计算网络启动所需的配置

- visualization【可视化服务】
    - docker-compose.yml: 启动可视化服务所需的配置


## 二、准备工作
服务启动依赖docker-compose，请先确认本地是否安装docker-compose

## 三、服务启动及任务执行
参考[README](../scripts/README.md)
