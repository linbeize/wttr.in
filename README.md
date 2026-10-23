
*wttr.in —— 用 `curl` 查天气的正确姿势！*

wttr.in 是一个面向命令行的天气预报服务，支持多种信息呈现方式：给 curl、httpie、wget
这类命令行 HTTP 客户端用的终端 ANSI 转义序列，给浏览器看的 HTML，以及给图形查看器用的 PNG。

它最初只是 [wego](https://github.com/schachmat/wego) 的一个小型封装，用来演示面向命令行的服务能有多强；
后来 *wttr.in* 成长为一个流行的天气查询服务，每天处理上千万次查询[¹](#wttrin-使用统计)。

> **本仓库**（`linbeize/wttr.in`）是 `chubin/wttr.in` 的简体中文本地化分支，
> IP 库已换成面向中国优化的 **bt.cn-ip** mmdb。完整改动清单见 [改造说明.md](改造说明.md)。

线上服务见：[wttr.in](https://wttr.in)。

[文档](https://wttr.in/:help) | [用法](#用法) | [单行输出](#单行输出) | [数据丰富输出格式-v2](#数据丰富输出格式-v2) | [地图视图](#地图视图-v3) | [输出格式](#其他输出格式) | [月相](#月相) | [国际化与本地化](#国际化与本地化) | [安装部署](#安装部署) | [反向代理](#反向代理)

## 用法

在命令行或浏览器里都可以这样访问：

    $ curl wttr.in
    北京 的天气

         \   /     晴
          .-.      10 – 11 °C
       ― (   ) ―   ↑ 11 km/h
          `-’      10 km
         /   \     0.0 mm


完整天气预报示例：

![天气预报](share/pics/San_Francisco.png)

PowerShell 里也可以：

```PowerShell
Invoke-RestMethod https://wttr.in
```

要查指定地点，把地点名拼到 URL 上即可：

    $ curl wttr.in/Beijing
    $ curl wttr.in/Shanghai
    $ curl wttr.in/Salt+Lake+City

省略地点名时，会根据你的 IP 地址自动定位到当前位置。

也可以用三字机场代码查询某个机场的天气：

    $ curl wttr.in/pek      # 首都国际机场
    $ curl wttr.in/pvg      # 浦东国际机场

如果想查的不是城市，而是某个景点、山峰或其他特殊地点：

	$ curl wttr.in/故宫
	$ curl wttr.in/Eiffel+Tower
	$ curl wttr.in/Kilimanjaro

这类查询会在预报下方多显示一行，给出地理编码的结果：

	Location: Vostok Station, станция Восток, AAT, Antarctica [-78.4642714,106.8364678]
	Location: Tour Eiffel, Paris, Île-de-France, 75007, France [48.8582602,2.29449905432]
	Location: Kilimanjaro, Northern, Tanzania [-3.4762789,37.3872648]

地点还可以直接写 IP 地址，或者用 `@` 前缀指定域名：

    $ curl wttr.in/@github.com
    $ curl wttr.in/@msu.ru

想看详细说明，访问 [/:help](https://wttr.in/:help) 页面：

    $ curl wttr.in/:help

### 单位

默认情况下，来自美国的请求使用 USCS 单位，世界其他地区使用公制。
可以用 `?u`、`?m` 或 `?M` 显式覆盖：

    $ curl wttr.in/Amsterdam?u  # USCS（美国默认）
    $ curl wttr.in/Amsterdam?m  # 公制（SI）（除美国外的默认值）
    $ curl wttr.in/Amsterdam?M  # 公制（SI），但风速用 m/s 显示

需要传多个选项时，单字母选项之间不加分隔符直接连写；
带值的长选项则用 `&` 分隔：

    $ curl 'wttr.in/Amsterdam?m2&lang=nl'

大致相当于 GNU 命令行语法里的 `-m2 --lang nl`。

## 支持的输出格式与视图

wttr.in 目前支持六种输出格式：

* ANSI —— 终端彩色输出；
* 纯文本 —— 终端与脚本；
* HTML —— 浏览器；
* PNG —— 图形查看器；
* JSON —— 脚本与 API；
* Prometheus 指标 —— 监控系统与 API。

ANSI 与 HTML 格式会根据 User-Agent 自动选择。

强制使用纯文本（关闭颜色）：

    $ curl wttr.in/?T

限制输出只使用标准控制台字体（例如 Consolas、Lucida Console）中存在的字形：

    $ curl wttr.in/?d

在查询地址末尾加 `.png` 可强制使用 PNG 格式：

    $ wget wttr.in/Beijing.png

PNG 格式下同样可以使用所有选项，但选项之间要用 `_` 而不是 `?` 和 `&` 分隔：

    $ wget wttr.in/Beijing_0tqp_lang=zh-cn.png

PNG 格式的常用选项：

* `t` 表示透明（等价于 `transparency=150`）；
* `transparency=0..255` 自定义透明程度。

把天气 PNG 叠到照片上时，透明背景非常有用：

    $ convert source.jpg <( curl wttr.in/Oymyakon_tqp0.png ) -geometry +50+50 -composite target.jpg

上例中：

* `source.jpg` —— 源图；
* `target.jpg` —— 目标图；
* `Oymyakon` —— 地点名；
* `tqp0` —— 选项（推荐写法）。

![带天气数据的图片](https://pbs.twimg.com/media/C69-wsIW0AAcAD5.jpg)

借助 [wttr-switcher](https://github.com/midzer/wttr-switcher)，可以把显示当前或指定地点天气的小组件
嵌入 HTML 页面。效果见：[wttr-switcher 示例](https://midzer.github.io/wttr-switcher/)，
真实站点案例：https://feuerwehr-eisolzried.de/。

![feuerwehr-eisolzried.de 上的 wttr.in 嵌入示例](https://user-images.githubusercontent.com/3875145/65265457-50eac180-db11-11e9-8f9b-2e1711dfc436.png)

## 单行输出

单行输出适合把天气塞进 *tmux*、*weechat* 等程序的状态栏。

使用 `format` 参数即可得到单行输出：

```
$ curl wttr.in/Nuremberg?format=3
Nuremberg: 🌦 +11⁰C
```

内置的预设格式有 1、2、3、4，也可以用百分号记法自定义（见下文）。
* 1：地点 + 当前天气：`🌦 +11⁰C`
* 2：地点 + 当前天气（含更多细节）：`🌦   🌡️+11°C 🌬️↓4km/h`
* 3：地点名 + 当前天气：`Nuremberg: 🌦 +11⁰C`
* 4：地点名 + 当前天气（含更多细节）：`Nuremberg: 🌦   🌡️+11°C 🌬️↓4km/h`

用 `:` 分隔可以指定多个地点（重复查询）：

```
$ curl wttr.in/Nuremberg:Hamburg:Berlin?format=3
Nuremberg: 🌦 +11⁰C
```
或一次性批量查询：

```
$ curl -s 'wttr.in/{Nuremberg,Hamburg,Berlin}?format=3'
Nuremberg: 🌦 +11⁰C
Hamburg: 🌦 +8⁰C
Berlin: 🌦 +8⁰C
```

自定义输出格式使用 `%` 记法：

```
    c    天气状况
    C    天气状况（文字名称）
    x    天气状况（纯文本符号）
    h    湿度
    t    温度（实际）
    f    温度（体感）
    H    最高温
    L    最低温
    w    风
    l    地点
    m    月相 🌑🌒🌓🌔🌕🌖🌗🌘
    M    月龄
    p    降水量（mm/3 小时）
    P    气压（hPa）
    e    露点
    u    紫外线指数（1-12）

    D    黎明*
    S    日出*
    z    正午*
    s    日落*
    d    黄昏*
    T    当前时间*
    Z    本地时区

（* 时间以当地时区显示）
```

所以下面两条命令等价：

```
    $ curl wttr.in/London?format=3
    London: ⛅️ +7⁰C
    $ curl wttr.in/London?format="%l:+%c+%t\n"
    London: ⛅️ +7⁰C
```

## 域名与可靠性

- **主域名**：[wttr.in](https://wttr.in)
- **备用域名**：[wttr.is](https://wttr.is) —— **完全等价的直接替换方案**

凡是用到 `wttr.in` 的地方都可以换成 `wttr.is`。
两个域名由同一套后端提供服务，并保持同步。

建议在脚本、状态栏、监控工具和 CI/CD 流水线中使用 `wttr.is`，以获得更好的可用性。

## 集成

得益于 *wttr.in* 极易集成，各种库、编程语言和系统里都有大量流行的集成方案。

*wttr.in* 可集成到：

* 终端管理器，
* 窗口管理器，
* 编辑器，
* 聊天客户端，

等等。这些集成把天气信息直接嵌进界面，提升了工作流的效率。

完整列表见：[wttr.in 集成清单](doc/integrations.md)，下面列举其中一部分。

### tmux

在 `tmux.conf` 中使用时，`%` 需要用 `%` 转义，也就是写成 `%%`。

使用 `%` 记法时输出默认不含换行，而使用预设格式（`1`、`2`、`3` 等）时会带上换行。
若要在 `%` 记法下也输出换行，写 `\n`，并在 shell 里用单引号包裹查询。

对于会自动轮询的程序（例如 tmux），最好设置一个合理的刷新间隔。tmux 里用 `status-interval` 配置。

如果查询里用 `:` 指定了多个地点，需要额外传入 `period=` 参数指定刷新周期：
```
set -g status-interval 60
WEATHER='#(curl -s wttr.in/London:Stockholm:Moscow\?format\="%%l:+%%c%%20%%t%%60%%w&period=60")'
set -g status-right "$WEATHER ..."
```
![tmux 状态栏中的 wttr.in](https://wttr.in/files/example-tmux-status-line.png)

### WeeChat

嵌入 IRC 客户端 [WeeChat](https://github.com/weechat/weechat) 的现有状态栏：

```
/alias add wttr /exec -pipe "/mute /set plugins.var.wttr" url:wttr.in/Montreal?format=%l:+%c+%f+%h+%p+%P+%m+%w+%S+%s;/wait 3 /item refresh wttr
/trigger add wttr timer 60000;0;0 "" "" "/wttr"
/item add wttr "" "${plugins.var.wttr}"
/eval /set weechat.bar.status.items ${weechat.bar.status.items},spacer,wttr
/eval /set weechat.startup.command_after_plugins ${weechat.startup.command_after_plugins};/wttr
/wttr
```
![WeeChat 状态栏中的 wttr.in](https://i.imgur.com/XkYiRU7.png)


### conky

conky 用法示例：

```
${texeci 1800 curl wttr.in/kyiv_0pq_lang=uk.png
  | convert - -transparent black $HOME/.config/conky/out.png}
${image $HOME/.config/conky/out.png -p 0,0}
```

![conky 中的 wttr.in](https://user-images.githubusercontent.com/3875145/172178453-9e9ed9e3-9815-426a-9a21-afdd6e279fc8.png)


### IRC

IRC 集成示例：

* https://github.com/OpenSourceTreasure/Mirc-ASCII-weather-translate-pixel-editor

### Emoji 支持

要在终端里看到 emoji，需要满足两个条件：

1. 终端支持 emoji（Cairo 1.15.8 起已支持）；
2. 使用带 emoji 字形的字体。

emoji 字体推荐 *Noto Color Emoji*，替代方案是 *Emoji One* 字体，两者都覆盖了所需的全部 emoji 字形。

字体配置：

```xml
$ cat ~/.config/fontconfig/fonts.conf
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE fontconfig SYSTEM "fonts.dtd">
<fontconfig>
  <alias>
    <family>serif</family>
    <prefer>
      <family>Noto Color Emoji</family>
    </prefer>
  </alias>
  <alias>
    <family>sans-serif</family>
    <prefer>
      <family>Noto Color Emoji</family>
    </prefer>
  </alias>
  <alias>
    <family>monospace</family>
    <prefer>
      <family>Noto Color Emoji</family>
    </prefer>
  </alias>
</fontconfig>
```

（改完执行 `fc-cache -f -v` 使其生效）。

某些情况下，`tmux` 与终端对个别 emoji 字符的理解会不一致，可能出现 #579 里描述的异常现象。

### Squeak

嵌入 Squeak 主停靠栏：

```smalltalk
wttr := (UpdatingStringMorph on: [(WebClient httpGet: 'https://wttr.in/?format=%20%20%l:%20%C+%t') content] selector: #value)
	stepTime: 60000;
	useStringFormat;
	yourself.
dockingBar := World mainDockingBars first.
dockingBar addMorph: wttr after: (dockingBar findA: ClockMorph).
```

![Squeak 主停靠栏中的 wttr.in](https://github.com/user-attachments/assets/4c2762b0-77ae-41a8-98db-3eb310d073bd)

## 数据丰富输出格式 v2

这个实验性的数据丰富格式对应视图代码 `v2`，提供大量额外的天气与天文信息：

* 全天温度与降水的逐时变化曲线；
* 今天及未来三天的月相；
* 当前天气状况、温度、湿度、风速风向、气压；
* 时区；
* 所选地点的黎明、日出、正午、日落、黄昏时间；
* 所选地点的精确经纬度。

```
  $ curl v2.wttr.in/Beijing
```

或者

```
  $ curl wttr.in/Beijing?format=v2
```

如果不想用 Emoji，而偏好 Nerd Fonts，可以用 `v2d`（白天）或 `v2n`（夜间）：

```
  $ curl v2d.wttr.in/Beijing
```


![数据丰富输出格式](https://wttr.in/files/example-wttr-v2.png)

（该模式仍属实验性质，目前有几处限制：

* 仅支持终端；
* 部分内容仅支持英文）。

部分终端需要做些调整才能得到最佳视觉效果。

### URXVT

视你的配置而定，下面的步骤可能全都要做，也可能只做一部分。URXVT 目前不支持 emoji 相关字体，
但用 *Font-Symbola* 可以获得几乎相同的效果。在 `.Xresources` 里加入这一行：
```
    xft:symbola:size=10:minspace=False
```
把它放在你首选字体的**后面**，只在需要时才会用到。
如果出现或感觉字距有问题，再加上 `URxvt.letterSpace: 0`。
URXVT 有时会判断错单词间距，需要这样强制一下。

效果大致如下：

![URXVT Emoji 行](https://user-images.githubusercontent.com/24360204/63842949-1d36d480-c975-11e9-81dd-998d1329bd8a.png)

## 地图视图 v3

这个实验性的地图视图对应视图代码 `v3`，提供某个地理区域的天气信息：

```
    $ curl v3.wttr.in/Bayern.sxl
```

![v3 地图视图示例](share/pics/San_Francisco.png)

也可以直接在浏览器打开：

*   https://v3.wttr.in/Bayern

地图视图目前支持三种格式：

* PNG（浏览器与即时通讯）；
* Sixel（终端内联图片）；
* IIP（支持 iterm2 内联图片协议的终端）。

各终端的内联图片协议支持情况：

⟶ *详见：[终端中的图片](doc/terminal-images.md)*

| 终端                   | 环境       | 图片支持 | 协议 |
| --------------------- | --------- | ------------- | --------- |
| uxterm                |   X11     |   支持         |   Sixel   |
| mlterm                |   X11     |   支持         |   Sixel   |
| kitty                 |   X11     |   支持         |   Kitty   |
| wezterm               |   X11     |   支持         |   IIP     |
| Darktile              |   X11     |   支持         |   Sixel   |
| Jexer                 |   X11     |   支持         |   Sixel   |
| GNOME Terminal        |   X11     |   [开发中](https://gitlab.gnome.org/GNOME/vte/-/issues/253) |   Sixel   |
| alacritty             |   X11     |   [开发中](https://github.com/alacritty/alacritty/issues/910) |  Sixel   |
| foot                  |  Wayland  |   支持         |   Sixel   |
| DomTerm               |   Web     |   支持         |   Sixel   |
| Yaft                  |   FB      |   支持         |   Sixel   |
| iTerm2                |   Mac OS X|   支持         |   IIP     |
| mintty                | Windows   |   支持         |   Sixel   |
| Windows Terminal  |   Windows     |   [开发中](https://github.com/microsoft/terminal/issues/448) |   Sixel   |
| [RLogin](http://nanno.dip.jp/softlib/man/rlogin/) | Windows | 支持         |   Sixel   |   |


## 其他输出格式

### JSON 输出

JSON 格式让程序能以易解析的方式访问 *wttr.in* 的数据，无需自己写复杂脚本去解析 wttr.in 的图形化输出。

获取 JSON 格式数据：

    $ curl wttr.in/Detroit?format=j1

这会以 JSON 返回底特律地区的信息。使用 `j1` 这个格式代码，是为了将来能给 JSON 输出换成其他版式。

返回结果大致如下：
```json
{
	"current_condition": [
		{
		    "FeelsLikeC": "25",
		    "FeelsLikeF": "76",
		    "cloudcover": "100",
		    "humidity": "76",
		    "observation_time": "04:08 PM",
		    "precipMM": "0.2",
		    "pressure": "1019",
		    "temp_C": "22",
		    "temp_F": "72",
		    "uvIndex": 5,
		    "visibility": "16",
		    "weatherCode": "122",
		    "weatherDesc": [
			{
			    "value": "Overcast"
			}
		    ],
		    "weatherIconUrl": [
			{
			    "value": ""
			}
		    ],
		    "winddir16Point": "NNE",
		    "winddirDegree": "20",
		    "windspeedKmph": "7",
		    "windspeedMiles": "4"
		}
	],
...
```

除 `weatherCode` 外，大部分字段含义一看即知。`weatherCode` 是一套枚举，可在
[WorldWeatherOnline 官网](https://www.worldweatheronline.com/developer/api/docs/weather-icons.aspx)
或 [wttr.in 源码](https://github.com/chubin/wttr.in/blob/master/lib/constants.py) 中查到。

另有不含逐时数据的精简版本 `format=j2`，适合内存受限的单片机使用。

### Prometheus 指标输出

[Prometheus](https://github.com/prometheus/prometheus) 指标格式让监控系统能以易解析的方式访问
*wttr.in* 的数据，无需自己写复杂脚本去解析 wttr.in 的图形化输出。

获取 Prometheus 格式数据：

    $ curl wttr.in/Detroit?format=p1

这会以 Prometheus 指标格式返回底特律地区的信息。使用 `p1` 格式代码，是为了将来能给输出换成其他版式。

Prometheus 的配置可以写成这样：

```yaml
    - job_name: 'wttr_in_detroit'
        static_configs:
            - targets: ['wttr.in']
        metrics_path: '/Detroit'
        params:
            format: ['p1']
```

返回结果大致如下：


    # HELP temperature_feels_like_celsius Feels Like Temperature in Celsius
    temperature_feels_like_celsius{forecast="current"} 7
    # HELP temperature_feels_like_fahrenheit Feels Like Temperature in Fahrenheit
    temperature_feels_like_fahrenheit{forecast="current"} 45
    [已截断]
...


## 月相

wttr.in 也可以用来查月相。下面是在完整输出模式下查看当前月相的例子：

    $ curl wttr.in/Moon

加上 `@YYYY-MM-DD` 可以查指定日期的月相：

    $ curl wttr.in/Moon@2016-12-25

月相计算使用 [pyphoon](https://github.com/chubin/pyphoon) 作为后端。

单行模式下用 `%m` 获取月相：

    $ curl wttr.in/London?format=%m
    🌖

需要注意，用 Unicode 字符表示月相有两个固有缺陷：

- 在某些字体下，`🌘` 的呈现是有歧义的：它看起来接近全暗还是接近全亮，
  取决于你的终端处于浅色模式还是深色模式。改用 `noto-fonts` 这类彩色字体可以绕过该问题。

- `🌘` 的含义同样有歧义：在北半球它表示「下弦月」，在南半球却表示「上弦月」，
  在热带地区则没有意义。这是 [Unicode](https://www.unicode.org/L2/L2017/17304-moon-var.pdf)
  已知的局限，而 wttr.in 目前尚未处理。

相关问题见 #247、#364，以及 pyphoon 的 [pyphoon#1](https://github.com/chubin/pyphoon/issues/1)。欢迎参与改进。

## 国际化与本地化

wttr.in 支持多语言地点名，可以用世界上任何语言书写
（说来有点意外，但世界上很多地点并没有英文名）。

查询串应以 Unicode 书写（可用十六进制转义，也可直接写）。查询串里的空格必须替换成 `+`：

    $ curl wttr.in/станция+Восток
    天气预报： станция Восток

                   阴
          .--.     -65 – -47 °C
       .-(    ).   ↑ 23 km/h
      (___.__)__)  15 km
                   0.0 mm

输出所用的语言（地点名除外）与输入语言无关，只取决于两种来源：默认的英语，
或者浏览器请求头 `Accept-Language` 中声明的首选语言（当请求来自浏览器时）。

用命令行客户端时，可以通过命令行选项显式指定语言：

    curl -H "Accept-Language: fr" wttr.in
    http GET wttr.in Accept-Language:ru

也可以用 `lang` 选项强制指定：

    $ curl wttr.in/Berlin?lang=de
    $ curl wttr.in/Beijing?lang=zh-cn      # 简体中文
    $ curl wttr.in/Taipei?lang=zh-tw       # 繁体中文

第三种方式是用查询所用的 DNS 名来选择语言：

    $ curl de.wttr.in/Berlin

wttr.in 目前已被翻译成 75 种语言，且支持的语言数量还在持续增加。

翻译流程、支持语言清单、贡献者名单，以及如何帮助翻译你的母语，见
[/:translation](https://wttr.in/:translation)。

![各种语言的 wttr.in 查询](https://pbs.twimg.com/media/C7hShiDXQAES6z1.jpg)

## 安装部署

本指南说明如何从源码安装 wttr.in。

它是一个**单一静态二进制**，所有资源（包括 PNG 渲染所需的字体）都已内嵌其中。
它**没有运行时依赖**，原生支持 HTTP 与 HTTPS，前面不需要 nginx、Apache 或任何其他 Web 服务器。

按本文装完后，你会得到：

- 一套干净、自包含的 wttr.in 服务，安装在 `/wttr.in/`
- 一个可执行文件（`/wttr.in/bin/srv`）
- 已配置好缓存、IP 地理位置（GeoIP2）与地点解析（OpenCage）
- 服务运行在 8080 端口（HTTP），功能与公开的 wttr.in 完全等价，可直接提供天气预报

### 系统准备

```bash
sudo mkdir -p /wttr.in/{bin,cache,log,data,etc}
```

如果希望用专用用户运行服务（推荐）：
```bash
sudo useradd -r -s /bin/false -u 1000 wttr
sudo chown -R wttr:wttr /wttr.in
```

### 下载 IP 数据库

用于自动 IP 定位：

**方式一：bt.cn-ip（推荐，中国地区更准确）**

- 获取 **bt_monitor.mmdb**（bt.cn-ip 格式），放到：`/wttr.in/data/bt_monitor.mmdb`
- 该库为 IPv4 专用，IPv6 会自动回退到 GeoIP2

**方式二：MaxMind GeoLite2**

- 在 [MaxMind](https://www.maxmind.com) 注册并下载 **GeoLite2-City.mmdb**
- 放到：`/wttr.in/data/GeoLite2-City.mmdb`

两个库可以同时配置：bt.cn-ip 优先，GeoIP2 作为全球与 IPv6 的兜底。

### 编译

```bash
git clone https://github.com/linbeize/wttr.in.git
cd wttr.in

# 编译期字体（会内嵌进最终二进制）
sudo apt-get install -y --no-install-recommends \
  fontconfig fonts-dejavu-core fonts-noto-core fonts-noto-cjk \
  fonts-wqy-zenhei fonts-symbola fonts-motoya-l-cedar fonts-lexi-gulim

bash build.sh build
```

注意：`go.mod` 要求 Go 1.25 及以上，用旧版本会直接报错。
`build.sh` 会先把 `share/` 复制到 `internal/assets/embed/` 供 `//go:embed` 使用。

### 安装二进制

```bash
sudo cp srv /wttr.in/bin/srv
sudo chown wttr:wttr /wttr.in/bin/srv
```

### 配置

创建 `/wttr.in/config.yaml`：

```yaml

cache:
  # 最终渲染结果缓存（HTML、JSON、PNG、文本等）
  responses:
    type: lru
    size: 50000
    ttl: 10m             # 渲染结果用较短的 TTL
    enabled: true

  # 上游 API 原始天气数据缓存（WWO 等）
  weather:
    type: disk
    dir: /var/cache/wttr.in/weather
    ttl: 45m
    max_size_mb: 2048     # 2 GB 软上限
    cleanup_interval: 15m
    enabled: true

ip:
  ipCacheDb: /wttr.in/cache/geoip.db
  ipCacheType: db
  # bt.cn-ip：中国地区更准，仅 IPv4，优先使用
  btGeoip: /wttr.in/data/bt_monitor.mmdb
  # GeoIP2：全球兜底，同时覆盖 IPv6
  geoip2: /wttr.in/data/GeoLite2-City.mmdb

geo:
  locationCacheDb: /wttr.in/cache/geoloc.db
  locationCacheType: db
  # 地理编码器按顺序尝试，第一个命中即返回
  nominatim:
    # 公开 Nominatim：免 key，lang 决定结果语言，中文地名靠它
    - name: osm
      type: osm
      url: https://nominatim.openstreetmap.org/search
      lang: zh
    - name: opencage
      type: opencage
      url: https://api.opencagedata.com/geocode/v1/json
      token: "YOUR_OPENCAGE_TOKEN_HERE"

weather:
  wwo:
    baseUrl: "http://wttr.in/{lat},{lon}?format=j1&lang={lang}"
    # key: "YOUR_WWO_TOKEN"   # 可选

# 站点级默认值
site:
  defaultLang: "zh-cn"        # 请求未指定语言时使用
  defaultLocation: "北京"      # 请求无地点且 IP 定位不到时使用

server:
  portHttp: 8080
  # 反向代理：只有来自这些网段的 X-Forwarded-For / X-Real-IP 才可信
  # 代理与服务同机填回环；容器部署填容器网段。留空则完全忽略转发头。
  trustedProxies:
    - 127.0.0.1/32
    - ::1/128
  # portHttps: 8443
  # tlsCertFile: /wttr.in/etc/fullchain.pem
  # tlsKeyFile: /wttr.in/etc/privkey.pem
```

上述配置默认用 OpenCage 做地理编码（需要 token），并用 wttr.in 作为天气数据源。
若想改用 *WorldWeatherOnline*，需在其官网注册获取 token，并相应修改 `baseUrl`。

### 启动服务

```bash
cd /wttr.in
sudo -u wttr ./bin/srv srv config.yaml
```

现在就可以通过 `http://你的服务器:8080` 访问自己的实例了（例如 `http://你的服务器:8080/London`）。

生产环境建议配一个 systemd 服务。缓存与日志分别位于 `/wttr.in/cache` 和 `/wttr.in/log`。

## 反向代理

放在 nginx / Caddy / 宝塔面板后面时，**必须配置 `trustedProxies`**，否则基于 IP 的自动定位
要么失效、要么可被伪造。

原因：经过代理后 `RemoteAddr` 是代理地址，真实客户端 IP 只在 `X-Forwarded-For`
或 `X-Real-IP` 里，而这两个头客户端可以自己伪造。原实现直接取 `X-Forwarded-For`
最左值，攻击者加一个头就能伪造定位；且 nginx 的 `$proxy_add_x_forwarded_for` 是**追加**语义，
头部会变成 `伪造值, 真实IP`，取最左值拿到的正是伪造值。

现在只有直接对端命中 `trustedProxies` 时才解析转发头，并从**右往左**取第一个不属于
可信代理的地址——代理逐跳追加，最右的不可信地址才是代理真正看到的客户端。

```yaml
server:
  portHttp: 8901
  trustedProxies:
    - 127.0.0.1/32      # 代理同机
    - ::1/128
```

nginx 侧（关键两行）：

```nginx
proxy_set_header Host $host;                              # 语言子域靠它
proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
```

完整说明（nginx / Caddy / 宝塔 / HTTPS 两种做法 / 语言子域 / 排错）见
[doc/reverse-proxy.md](doc/reverse-proxy.md)。

排查时加 `?debug=true`，正文会输出解析出的 IP 与定位结果：

```bash
curl "http://127.0.0.1:8901/?debug=true"
```

## 致谢

### OpenCage

![opencage 标志](share/pics/opencage-logo.png)

wttr.in 项目感谢 **[OpenCage](https://opencagedata.com/)** 通过 **OpenCage Geocoding API** 提供地理位置数据。

OpenCage 负责把地名（城市、地标、车站等）解析成坐标与规范化地址。正是这项能力让
`wttr.in/London`、`wttr.in/~Eiffel+Tower` 这类查询成为可能。

感谢 OpenCage 团队提供稳定可靠的地理编码服务、易用的 API，以及对开放工具的支持，
让天气预报得以走进终端。

### 上游项目

本项目基于 [chubin/wttr.in](https://github.com/chubin/wttr.in) 二次开发，
并依赖 [wego](https://github.com/schachmat/wego)、[pyphoon](https://github.com/chubin/pyphoon)
等开源项目。

## wttr.in 使用统计

截至 2026 年 4 月，根据访问日志，*wttr.in* 每天处理约一亿次查询，来自 40 万到 45 万名用户。

![wttr.in 使用统计](share/stats/stats.png)
