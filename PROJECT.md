# Netprobe · TWAMP 会话控制器（RFC 5357）

## 这不是什么

不是天气应用、不是「网络拨测大屏」、不是监控报表。产品主体是 **TWAMP 测试会话**的参数与会话态：mode、padding、反射器。

## 谁在用

城域网运维。需要在 PE 之间跑 TWAMP（Two-Way Active Measurement Protocol）测单向时延/抖动。本控制台登记会话、校验参数，会话等待走 context 取消，而不是睡死。

## 核心业务

1. 正文含 `mode=unauthenticated|authenticated|encrypted` 与 `pad=`（填充字节）。
2. `pad` 必须在 0..1472（以太网 UDP 实用上限，避免分片）。
3. 标签为反射器角色 `controller` / `reflector`。
4. 会话记录可按 slug 热缓存读取；修订保留 mode 变更。
5. 不实现完整 TWAMP 发包（避免变成发包工具/游戏），只做会话对象与参数闸门。

## 运行与验收

- `pad=9000` 拒绝；`mode=unauthenticated pad=64` 接受。
- `/healthz` 与首页可开。不提供 CSV 大盘。
