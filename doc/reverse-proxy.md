# 反向代理配置

把 wttr.in 放在 nginx / Caddy / 宝塔面板后面时，只有一件事必须改：
**告诉服务哪个地址是可信代理**，否则基于 IP 的自动定位会失效或可被伪造。

## 为什么必须配

服务需要客户端真实 IP 才能定位天气。经过反向代理后，`RemoteAddr` 是代理的地址，
真实 IP 只存在于 `X-Forwarded-For` / `X-Real-IP` 请求头里——而**这两个头客户端可以自己伪造**。

原实现直接取 `X-Forwarded-For` 的第一个值，有两个后果：

1. 攻击者加一个 `X-Forwarded-For: 1.1.1.1` 就能让别人看到堪萨斯的天气；
2. nginx 用 `$proxy_add_x_forwarded_for` 时是**追加**而非覆盖，头部变成
   `伪造值, 真实IP`，取最左值拿到的正是伪造值。

现在改为：只有直接对端（`RemoteAddr`）命中 `trustedProxies` 列表时，才解析转发头，
并且从**右往左**找第一个不属于可信代理的地址。代理是逐跳追加的，所以最右的不可信地址
才是代理真正看到的那个客户端。

## 配置

```yaml
server:
  portHttp: 8901
  trustedProxies:
    - 127.0.0.1/32      # 代理与服务同机（nginx / Caddy 在本机）
    - ::1/128           # 同上，IPv6
```

容器部署时填 Docker 网段：

```yaml
server:
  trustedProxies:
    - 172.17.0.0/16     # docker0 默认网段
```

也可以只写 IP，等价于 `/32`（IPv4）或 `/128`（IPv6）：

```yaml
server:
  trustedProxies:
    - 127.0.0.1
```

**留空则完全信任 `RemoteAddr`，忽略所有转发头**——这是安全默认值，适合服务直接暴露的场景。
配错了不会崩，非法网段会记一条 warn 日志后跳过。

## nginx

```nginx
server {
    listen 80;
    server_name wttr.example.com;

    location / {
        proxy_pass http://127.0.0.1:8901;

        # Host 必须原样传递：服务用它的前缀判断语言子域（zh-cn.example.com）
        proxy_set_header Host $host;

        # 关键：追加真实客户端地址
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;

        proxy_http_version 1.1;
        proxy_set_header Connection "";

        # 冷启动请求要拉上游天气，给足超时
        proxy_read_timeout 90s;
    }
}
```

`$proxy_add_x_forwarded_for` 保留客户端自带的 `X-Forwarded-For` 再追加真实地址。
这是正确的做法，服务的右→左解析正是为它设计的。

不要用 `proxy_set_header X-Forwarded-For $remote_addr;`（覆盖式）来"防伪造"：
那样会丢掉真实链路，多层代理时反而定位错误。

## Caddy

```caddyfile
wttr.example.com {
    reverse_proxy 127.0.0.1:8901
}
```

Caddy 默认就设置 `X-Forwarded-For` / `X-Forwarded-Proto` 并保留 `Host`，无需额外配置。
只需在服务端把 `127.0.0.1/32` 加入 `trustedProxies`。

## 宝塔面板

「网站 → 设置 → 反向代理」里新建代理，目标 `http://127.0.0.1:8901`。
面板生成的配置已包含 `proxy_set_header Host $host` 与 `X-Forwarded-For`，
但**默认可能使用覆盖式**。检查配置文件里是否为：

```nginx
proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
```

不是的话改过来，否则多级代理场景会丢失客户端 IP。

## HTTPS

两种做法，任选其一：

**1. 代理终止 TLS（推荐）**

代理监听 443 并持有证书，服务只开 HTTP：

```yaml
server:
  portHttp: 8901
  portHttps: 0
```

生成的链接与缓存都基于 `Host` / `X-Forwarded-Proto`，不需要额外配置。

**2. 服务自己终止 TLS**

```yaml
server:
  portHttp: 8901
  portHttps: 8443
  tlsCertFile: /wttr.in/etc/fullchain.pem
  tlsKeyFile: /wttr.in/etc/privkey.pem
```

或不写这两个字段、改用多域名/通配符证书：

```yaml
server:
  portHttps: 8443
  tlsCerts:
    - domain: "wttr.example.com"
      certFile: /wttr.in/etc/wttr.example.com/fullchain.pem
      keyFile:  /wttr.in/etc/wttr.example.com/privkey.pem
```

此时代理只做四层转发（stream / `proxy_pass` 到 8443），不要做七层拆包。

## 语言子域

服务用 `Host` 的第一段判断语言：

```
zh-cn.example.com/Beijing   → 简体中文
zh-tw.example.com/Taipei    → 繁体中文
```

所以要支持子域切换语言，代理必须传递原始 `Host`（nginx 用 `proxy_set_header Host $host;`，
不能写成 `$proxy_host`），并且为这些子域配置泛解析与证书（`*.example.com`）。

## 缓存

响应缓存键包含 `Host`、`URI`、客户端 IP、`Accept-Language` 与 `User-Agent`，
因此不同用户、不同语言、不同子域的缓存天然隔离，代理层无需额外处理。

被代理的请求如果带 `?debug=true`，会跳过缓存直接计算并在正文输出诊断信息
（含解析出的 IP 与定位结果），排查反代问题时很有用：

```bash
curl "https://wttr.example.com/?debug=true"
```

## 排错

**自动定位始终显示代理所在城市**

`trustedProxies` 没配。检查启动日志：

```
no trusted proxies configured: client IP taken from RemoteAddr
```

配上之后应变成：

```
trusting forwarding headers from configured proxies  proxies="[127.0.0.1/32 ::1/128]"
```

**配置了但定位仍然不对**

用 `?debug=true` 看 `Resolved IP:` 字段，确认它是不是真实客户端地址。
如果显示的是代理地址，说明代理没传头；如果显示的是伪造值，说明可信网段配得过宽
（例如写成了 `0.0.0.0/0`，等于信任一切）。

**502 / 504**

先直连服务确认它活着：

```bash
curl http://127.0.0.1:8901/?0
```

能通就是代理侧超时太短。冷请求要拉上游天气 API，把 `proxy_read_timeout` 放到 60–90 秒。

**中文全角字符显示错位**

这是浏览器字体问题，不是代理问题。服务端已保证每行按显示宽度对齐，
模板里也用 `@font-face` 把 ASCII / braille / CJK 归一化到 1:1:2 列宽。
若仍有错位，检查代理是否改写了 `Content-Type` 的 charset（应保留 `charset=utf-8`），
以及是否启用了会重排空白的 HTML 压缩模块。
