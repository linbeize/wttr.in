# systemd 部署

把 wttr.in 作为系统服务运行，监听 **8091** 端口，开机自启、崩溃自动重启。

## 一键部署

```bash
# 1. 目录与专用用户
sudo mkdir -p /wttr.in/{bin,cache,log,data,etc}
sudo useradd -r -s /usr/sbin/nologin -d /wttr.in wttr

# 2. 二进制与 IP 库
sudo cp srv /wttr.in/bin/srv
sudo chmod 755 /wttr.in/bin/srv
sudo cp bt_monitor.mmdb /wttr.in/data/

# 3. 配置（内容见下）
sudo cp config.yaml /wttr.in/etc/config.yaml

# 4. 目录属主：只有 cache/log 归 wttr，其余归 root
sudo chown root:root /wttr.in /wttr.in/{bin,data,etc}
sudo chown -R wttr:wttr /wttr.in/{cache,log}
sudo chmod 755 /wttr.in /wttr.in/{bin,data,etc}
sudo chmod 600 /wttr.in/etc/config.yaml

# 5. 安装并启动
sudo cp share/systemd/wttr.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable --now wttr

# 6. 验证
curl http://127.0.0.1:8091/?0
```

## 日常操作

```bash
systemctl status wttr          # 状态
systemctl restart wttr         # 重启（改配置后必须）
systemctl stop wttr            # 停止
systemctl disable wttr         # 取消开机自启
journalctl -u wttr -f          # 实时日志
journalctl -u wttr -n 100      # 最近 100 行
journalctl -u wttr --since "10 min ago"
```

## 为什么目录属主要那样分

`ProtectSystem=strict` 会把整个文件系统对服务进程挂成只读，
再通过 `ReadWritePaths` 放开 `/wttr.in/cache` 与 `/wttr.in/log`。

但**这只约束 systemd 启动的进程**。如果 `/wttr.in` 本身属于 `wttr` 用户，
那么任何以 wttr 身份运行的进程（包括被攻破后拿到的 shell）都能改写二进制
和配置——那就等于沙箱白做了。

所以：`/wttr.in`、`bin`、`etc`、`data` 归 root 只读；只有 `cache`、`log` 归 wttr 可写。

验证沙箱确实生效（在服务进程的命名空间里试写）：

```bash
PID=$(systemctl show -p MainPID --value wttr)
sudo nsenter -t $PID -m touch /wttr.in/sandbox_test
# touch: cannot touch '/wttr.in/sandbox_test': Read-only file system
```

## 配置要点

完整配置见仓库根目录的 `share/config/config.yaml`，与生产部署相关的几项：

```yaml
server:
  portHttp: 8091
  portHttps: 0
  trustedProxies:
    - 127.0.0.1/32
    - ::1/128

logging:
  accessLog: /wttr.in/log/access.log
  errorsLog: /wttr.in/log/errors.log
  interval: 300      # 访问日志按此间隔批量落盘

cache:
  weather:
    dir: /wttr.in/cache/weather

site:
  defaultLang: "zh-cn"
  defaultLocation: "北京"
```

**`trustedProxies` 必须配**，否则反代后面拿不到真实客户端 IP，
基于 IP 的自动定位会全部落到代理所在地址。详见 [reverse-proxy.md](reverse-proxy.md)。

**`interval: 300`** 意味着访问日志每 5 分钟才写一次盘。想看实时请求看 journal
（`journalctl -u wttr -f`），别盯着 `access.log` 以为没生效。

## 资源占用

实测（Go 静态二进制 + CGO sqlite3）：

| 项目 | 数值 |
|---|---|
| 二进制大小 | 约 50 MB（字体已内嵌） |
| 常驻内存 | 约 165 MB |
| 启动耗时 | 约 1 秒（主要花在解压字体） |
| 冷请求 | < 0.5 秒（拉上游天气 API） |
| 命中缓存 | < 5 毫秒 |

unit 里限制了 `MemoryMax=2G`、`CPUQuota=200%`，够用且有防护余量。

## 升级

```bash
sudo systemctl stop wttr
sudo cp srv /wttr.in/bin/srv
sudo systemctl start wttr
curl http://127.0.0.1:8091/?0
```

二进制是单一静态文件（仅动态依赖 libc），替换即可，无需编译环境。

## 前端反代

服务监听 8091，建议前面挂 nginx / Caddy 做 TLS 终止。nginx 示例：

```nginx
server {
    listen 443 ssl http2;
    server_name wttr.example.com;

    ssl_certificate     /etc/letsencrypt/live/wttr.example.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/wttr.example.com/privkey.pem;

    location / {
        proxy_pass http://127.0.0.1:8091;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_http_version 1.1;
        proxy_set_header Connection "";
        proxy_read_timeout 90s;
    }
}
```

## 排错

**服务起不来**

```bash
journalctl -u wttr -n 50 --no-pager
```

常见原因：配置文件不存在或 YAML 语法错（`LoadFromYAML` 用严格模式，
多余字段会报错）、端口被占、`/wttr.in/cache` 不可写。

**端口被占**

```bash
ss -tlnp | grep 8091
```

**改了配置没生效**

```bash
sudo systemctl restart wttr
```

配置只在启动时读取，没有热重载。

**自动定位始终是北京**

说明 IP 定位没拿到客户端地址。检查 journal 里是否有：

```
trusting forwarding headers from configured proxies  proxies="[...]"
```

没有的话就是 `trustedProxies` 没配或没生效。

**返回 500 且日志有 `location not found`**

地理编码服务不可达。服务会回退到奥伊米亚康并给出本地化提示，不会白屏；
但如果所有请求都这样，检查 `/wttr.in/log/errors.log` 与出网连通性。

## 卸载

```bash
sudo systemctl disable --now wttr
sudo rm /etc/systemd/system/wttr.service
sudo systemctl daemon-reload
sudo rm -rf /wttr.in
sudo userdel wttr
```
