# 简介

基于 [v2fly/domain-list-community#256](https://github.com/v2fly/domain-list-community/issues/256) 的提议，重构 [v2fly/domain-list-community](https://github.com/v2fly/domain-list-community) 的构建流程，并添加新功能。

## 与官方版 `dlc.dat` 不同之处

- 将 `dlc.dat` 重命名为 `geosite.dat`
- 去除 `cn` 列表里带有 `@ads`、`@!cn` 属性的规则
- 去除 `geolocation-cn` 列表里带有 `@ads`、`@!cn` 属性的规则
- 去除 `geolocation-!cn` 列表里带有 `@ads`、`@cn` 属性的规则，尽量避免在中国大陆有接入点的海外公司的域名走代理，例如，避免国区 Steam 游戏下载服务走代理。

## 下载地址

> 如果无法访问域名 `raw.githubusercontent.com`，可以使用第二个地址（`cdn.jsdelivr.net`），但是内容更新会有 12 小时的延迟。

- **geosite.dat**：
  - [https://raw.githubusercontent.com/Loyalsoldier/domain-list-custom/release/geosite.dat](https://raw.githubusercontent.com/Loyalsoldier/domain-list-custom/release/geosite.dat)
  - [https://cdn.jsdelivr.net/gh/Loyalsoldier/domain-list-custom@release/geosite.dat](https://cdn.jsdelivr.net/gh/Loyalsoldier/domain-list-custom@release/geosite.dat)

## 使用本项目的项目

- [@Loyalsoldier/v2ray-rules-dat](https://github.com/Loyalsoldier/v2ray-rules-dat)
