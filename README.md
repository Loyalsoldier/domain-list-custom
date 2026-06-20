# 简介

基于 [v2fly/domain-list-community#256](https://github.com/v2fly/domain-list-community/issues/256) 的提议，重构 [v2fly/domain-list-community](https://github.com/v2fly/domain-list-community) 的构建流程，并添加新功能。

本项目采用插件化架构，参考 [Loyalsoldier/geoip](https://github.com/Loyalsoldier/geoip) 项目的设计模式，提供命令行界面（CLI）工具，支持灵活的配置文件，方便用户自定义域名列表的处理和输出。

## 与官方版 `dlc.dat` 不同之处

- 将 `dlc.dat` 重命名为 `geosite.dat`
- 去除 `cn` 列表里带有 `@ads`、`@!cn` 属性的规则
- 去除 `geolocation-cn` 列表里带有 `@ads`、`@!cn` 属性的规则
- 去除 `geolocation-!cn` 列表里带有 `@ads`、`@cn` 属性的规则，尽量避免在中国大陆有接入点的海外公司的域名走代理。例如，避免国区 Steam 游戏下载服务走代理。

## 下载地址

[https://github.com/Loyalsoldier/domain-list-custom/releases/latest/download/geosite.dat](https://github.com/Loyalsoldier/domain-list-custom/releases/latest/download/geosite.dat)

## 使用方法

### 命令行工具

```bash
# 查看帮助
./domain-list-custom --help

# 使用配置文件进行转换
./domain-list-custom convert -c config.json

# 使用远程配置文件
./domain-list-custom convert -c https://example.com/config.json
```

### 配置文件

配置文件采用 JSON 格式，包含 `input` 和 `output` 两个部分：

```json
{
  "input": [
    {
      "type": "domainlist",
      "action": "add",
      "args": {
        "dataDir": "./data"
      }
    }
  ],
  "output": [
    {
      "type": "v2rayGeoSite",
      "action": "output",
      "args": {
        "outputDir": "./output",
        "outputName": "geosite.dat",
        "excludeAttrs": "cn@!cn@ads,geolocation-cn@!cn@ads,geolocation-!cn@cn@ads",
        "gfwlistOutput": "geolocation-!cn"
      }
    },
    {
      "type": "text",
      "action": "output",
      "args": {
        "outputDir": "./output",
        "wantedList": ["cn", "google", "apple"]
      }
    }
  ]
}
```

更多配置示例请参考 `config.example.json`。

## 项目结构

```
.
├── lib/                 # 核心库
│   ├── lib.go          # 接口定义
│   ├── config.go       # 配置解析
│   ├── container.go    # 数据容器
│   ├── entry.go        # 条目管理
│   ├── instance.go     # 实例管理
│   └── common.go       # 通用函数
├── plugin/             # 插件目录
│   ├── plaintext/      # 文本格式插件
│   └── v2ray/          # V2Ray 格式插件
├── main.go             # 主程序入口
├── convert.go          # 转换命令
├── init.go             # 插件注册
└── config.json         # 配置文件
```

## 使用本项目的项目

[@Loyalsoldier/v2ray-rules-dat](https://github.com/Loyalsoldier/v2ray-rules-dat)
