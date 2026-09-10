# 插件开发指南

云枢面板提供插件扩展机制，允许开发者通过编写插件扩展面板功能，包括自定义应用模板、监控指标、通知渠道和操作扩展点。

## 插件架构

### 插件类型

| 类型 | 说明 | 示例 |
|------|------|------|
| 应用模板插件 | 扩展应用商店的应用列表 | 添加自定义应用一键安装模板 |
| 监控指标插件 | 扩展监控采集指标 | 添加自定义服务监控 |
| 通知渠道插件 | 扩展告警通知方式 | 添加企业微信/钉钉/飞书通知 |
| 操作扩展插件 | 扩展各模块的操作能力 | 添加自定义网站部署流程 |

### 插件目录结构

```
plugins/
└── my-plugin/
    ├── plugin.yaml          # 插件元信息
    ├── main.go              # 插件入口（Go 插件）
    ├── apps/                # 应用模板（应用模板插件）
    │   └── myapp.yaml
    ├── metrics/             # 自定义指标（监控插件）
    │   └── collector.go
    ├── notifiers/           # 通知渠道（通知插件）
    │   └── wecom.go
    └── static/              # 前端静态资源
        └── index.js
```

## 插件接口定义

### 插件元信息（plugin.yaml）

```yaml
# 插件基本信息
name: my-plugin
version: 1.0.0
description: 我的自定义插件
author: Your Name
license: MIT

# 插件类型（可多选）
types:
  - app-template
  - notifier

# 依赖的面板版本
min_panel_version: "1.0.0"

# 前端资源入口
frontend_entry: static/index.js

# 后端配置
backend:
  enabled: true
  entry: main.go
```

### 插件主接口

```go
// Plugin 是所有插件必须实现的主接口
type Plugin interface {
    // Init 初始化插件，传入插件上下文
    Init(ctx PluginContext) error

    // Name 返回插件名称
    Name() string

    // Version 返回插件版本
    Version() string

    // Types 返回插件类型列表
    Types() []PluginType

    // Cleanup 插件卸载时的清理工作
    Cleanup() error
}

// PluginContext 插件运行上下文
type PluginContext struct {
    Config    map[string]interface{}  // 插件配置
    DB        *gorm.DB                // 数据库连接
    Logger    *log.Logger             // 日志记录器
    EventBus  *EventBus               // 事件总线
    API       *PluginAPI              // 面板 API 封装
}

// PluginType 插件类型
type PluginType string

const (
    PluginTypeAppTemplate PluginType = "app-template"
    PluginTypeMetric      PluginType = "metric"
    PluginTypeNotifier    PluginType = "notifier"
    PluginTypeOperation   PluginType = "operation"
)
```

### 应用模板插件接口

```go
// AppTemplateProvider 应用模板提供者接口
type AppTemplateProvider interface {
    // GetApps 返回应用模板列表
    GetApps() []AppTemplate

    // GetApp 获取单个应用模板
    GetApp(key string) (*AppTemplate, error)

    // OnInstall 应用安装前的钩子
    OnInstall(appKey string, envParams map[string]string) error

    // OnUninstall 应用卸载后的钩子
    OnUninstall(appKey string) error
}

// AppTemplate 应用模板定义
type AppTemplate struct {
    Key          string            `yaml:"key"`
    Name         string            `yaml:"name"`
    Description  string            `yaml:"description"`
    Category     string            `yaml:"category"`
    Version      string            `yaml:"version"`
    Icon         string            `yaml:"icon"`
    Image        string            `yaml:"image"`         // Docker 镜像
    Ports        []PortMapping     `yaml:"ports"`         // 端口映射
    EnvParams    []EnvParam        `yaml:"env_params"`    // 环境变量定义
    Volumes      []VolumeMount     `yaml:"volumes"`       // 数据卷挂载
    Dependencies []string          `yaml:"dependencies"`  // 依赖应用
    InstallSteps []InstallStep     `yaml:"install_steps"` // 安装步骤
    PostInstall  string            `yaml:"post_install"`  // 安装后脚本
}
```

### 应用模板示例（apps/myapp.yaml）

```yaml
key: my-custom-app
name: 我的自定义应用
description: 一个示例自定义应用
category: development
version: 1.0.0
icon: https://example.com/icon.png
image: myregistry/myapp:1.0.0

ports:
  - container: 8080
    host: 8080
    protocol: tcp

env_params:
  - name: APP_PORT
    label: 应用端口
    default: "8080"
    required: true
  - name: DATABASE_URL
    label: 数据库连接串
    required: false
    description: 格式: mysql://user:pass@host:3306/dbname

volumes:
  - container_path: /data
    host_path: /data/apps/my-custom-app
    description: 应用数据目录

install_steps:
  - step: 检查端口
    command: "netstat -tlnp | grep 8080 || true"
  - step: 创建数据目录
    command: "mkdir -p /data/apps/my-custom-app"

post_install: |
  echo "应用安装完成，请访问 http://<服务器IP>:8080"
```

### 通知渠道插件接口

```go
// Notifier 通知渠道接口
type Notifier interface {
    // Type 返回通知渠道类型标识
    Type() string

    // Name 返回渠道显示名称
    Name() string

    // Send 发送通知
    Send(alert AlertMessage) error

    // Test 发送测试通知
    Test(config map[string]string) error

    // ConfigSchema 返回配置表单定义
    ConfigSchema() []ConfigField
}

// AlertMessage 告警消息
type AlertMessage struct {
    RuleName    string
    Metric      string
    Value       float64
    Threshold   float64
    Level       string  // info / warning / critical
    Message     string
    TriggeredAt time.Time
    ServerName  string
}

// ConfigField 配置字段定义
type ConfigField struct {
    Name        string `json:"name"`
    Label       string `json:"label"`
    Type        string `json:"type"`        // text / password / number / select / textarea
    Required    bool   `json:"required"`
    Default     string `json:"default"`
    Placeholder string `json:"placeholder"`
    Description string `json:"description"`
    Options     []Option `json:"options"`   // select 类型的选项
}
```

### 通知渠道示例（企业微信）

```go
package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
)

// WeComNotifier 企业微信通知渠道
type WeComNotifier struct{}

func (n *WeComNotifier) Type() string { return "wecom" }
func (n *WeComNotifier) Name() string { return "企业微信" }

func (n *WeComNotifier) ConfigSchema() []ConfigField {
    return []ConfigField{
        {
            Name:        "webhook_url",
            Label:       "Webhook 地址",
            Type:        "text",
            Required:    true,
            Placeholder: "https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=xxx",
            Description: "在企业微信群中添加机器人获取",
        },
    }
}

func (n *WeComNotifier) Send(alert AlertMessage) error {
    // 从配置中获取 webhook_url
    webhookURL := alert.Config["webhook_url"]

    content := fmt.Sprintf(
        "【云枢告警】%s\n级别: %s\n指标: %s\n当前值: %.2f\n阈值: %.2f\n时间: %s\n服务器: %s",
        alert.RuleName, alert.Level, alert.Metric,
        alert.Value, alert.Threshold,
        alert.TriggeredAt.Format("2006-01-02 15:04:05"),
        alert.ServerName,
    )

    payload := map[string]interface{}{
        "msgtype": "text",
        "text":    map[string]string{"content": content},
    }

    body, _ := json.Marshal(payload)
    resp, err := http.Post(webhookURL, "application/json", bytes.NewReader(body))
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("企业微信通知失败: %d", resp.StatusCode)
    }
    return nil
}

func (n *WeComNotifier) Test(config map[string]string) error {
    return n.Send(AlertMessage{
        RuleName:    "测试规则",
        Metric:      "cpu",
        Value:       50.0,
        Threshold:   80.0,
        Level:       "info",
        Message:     "这是一条测试告警消息",
        TriggeredAt: time.Now(),
        ServerName:  "测试服务器",
        Config:      config,
    })
}
```

### 监控指标插件接口

```go
// MetricCollector 自定义指标采集器接口
type MetricCollector interface {
    // Name 返回采集器名称
    Name() string

    // Collect 采集指标数据
    Collect() ([]MetricPoint, error)

    // Interval 返回采集间隔（秒）
    Interval() int
}

// MetricPoint 指标数据点
type MetricPoint struct {
    Metric    string  `json:"metric"`
    Value     float64 `json:"value"`
    Timestamp int64   `json:"timestamp"`
    Tags      map[string]string `json:"tags"`
}
```

### 监控指标示例（Nginx 状态采集）

```go
package main

import (
    "io/ioutil"
    "net/http"
    "strconv"
    "strings"
    "time"
)

// NginxStatusCollector Nginx 状态采集器
type NginxStatusCollector struct {
    URL string
}

func (c *NginxStatusCollector) Name() string { return "nginx_status" }
func (c *NginxStatusCollector) Interval() int { return 30 }

func (c *NginxStatusCollector) Collect() ([]MetricPoint, error) {
    resp, err := http.Get(c.URL)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    body, _ := ioutil.ReadAll(resp.Body)
    lines := strings.Split(string(body), "\n")

    now := time.Now().Unix()
    points := []MetricPoint{}

    // 解析 Active connections
    if len(lines) > 0 {
        parts := strings.Fields(lines[0])
        if len(parts) >= 3 {
            active, _ := strconv.ParseFloat(parts[2], 64)
            points = append(points, MetricPoint{
                Metric:    "nginx_active_connections",
                Value:     active,
                Timestamp: now,
            })
        }
    }

    return points, nil
}
```

## 扩展点说明

### 事件总线

插件可以通过事件总线监听和触发面板事件：

```go
// 可用事件
const (
    EventSiteCreated     = "site.created"
    EventSiteDeleted     = "site.deleted"
    EventSiteStarted     = "site.started"
    EventSiteStopped     = "site.stopped"
    EventDatabaseCreated = "database.created"
    EventBackupCreated   = "backup.created"
    EventBackupRestored  = "backup.restored"
    EventAlertTriggered  = "alert.triggered"
    EventUserLogin       = "user.login"
    EventCronTaskExecuted = "cron.executed"
)

// 监听事件
ctx.EventBus.On(EventSiteCreated, func(event Event) {
    site := event.Data.(*model.Site)
    // 自定义处理逻辑，如自动配置 DNS、发送通知等
})

// 触发事件
ctx.EventBus.Emit(Event{
    Name: "myplugin.custom_event",
    Data: customData,
})
```

### 面板 API 封装

插件可以通过 `PluginAPI` 调用面板内部能力：

```go
type PluginAPI struct {
    // 网站操作
    CreateSite(req CreateSiteRequest) (*model.Site, error)
    GetSite(id uint) (*model.Site, error)
    ListSites(page, pageSize int) ([]model.Site, int64, error)

    // 数据库操作
    CreateDatabaseInstance(req CreateDBInstanceRequest) (*model.DatabaseInstance, error)
    ExecuteSQL(instanceID uint, dbName, sql string) (*SQLResult, error)

    // 文件操作
    ReadFile(path string) (string, error)
    WriteFile(path, content string) error
    ExecuteCommand(name string, args ...string) (string, error)

    // 容器操作
    ListContainers(all bool) ([]ContainerInfo, error)
    StartContainer(id string) error
    StopContainer(id string) error

    // 通知操作
    SendNotification(channelID uint, message string) error

    // 数据库
    DB() *gorm.DB
}
```

### 前端扩展

插件可以通过前端 JS 文件扩展面板 UI：

```javascript
// static/index.js
// 注册自定义菜单项
window.CloudPivot.registerMenu({
  id: 'my-plugin',
  title: '我的插件',
  icon: 'extension',
  path: '/my-plugin',
  component: () => import('./views/MyPlugin.vue'),
  permission: 'myplugin:view'
})

// 注册自定义设置面板
window.CloudPivot.registerSettings({
  id: 'my-plugin-settings',
  title: '我的插件设置',
  component: MyPluginSettings
})

// 监听面板事件
window.CloudPivot.on('site.created', (site) => {
  console.log('网站已创建:', site)
})

// 调用面板 API
const sites = await window.CloudPivot.api.site.list()
```

## 插件配置

### 插件配置文件

用户可以在面板设置中配置插件参数，配置以键值对形式存储：

```yaml
# 插件配置示例（在面板设置中管理）
plugins:
  my-plugin:
    enabled: true
    config:
      api_key: "your-api-key"
      api_secret: "your-api-secret"
      log_level: "info"
```

### 插件配置读取

```go
func (p *MyPlugin) Init(ctx PluginContext) error {
    // 读取配置
    apiKey, _ := ctx.Config["api_key"].(string)
    logLevel, _ := ctx.Config["log_level"].(string)

    // 初始化
    p.client = NewClient(apiKey)
    p.logger = ctx.Logger

    return nil
}
```

## 插件开发流程

### 1. 创建插件目录

```bash
mkdir -p plugins/my-plugin/{apps,metrics,notifiers,static}
cd plugins/my-plugin
```

### 2. 编写 plugin.yaml

参考上方插件元信息模板。

### 3. 实现插件接口

根据插件类型实现对应接口：
- 应用模板：编写 `apps/*.yaml` + 实现 `AppTemplateProvider`
- 通知渠道：实现 `Notifier` 接口
- 监控指标：实现 `MetricCollector` 接口
- 操作扩展：监听事件总线 + 调用 PluginAPI

### 4. 编写前端扩展（可选）

在 `static/` 目录编写前端 JS/Vue 组件。

### 5. 本地测试

```bash
# 将插件目录复制到面板插件目录
cp -r my-plugin /opt/cloudpivot/plugins/

# 重启面板服务
systemctl restart cloudpivot

# 查看日志确认插件加载
journalctl -u cloudpivot -f | grep my-plugin
```

### 6. 打包发布

```bash
# 打包插件
tar czf my-plugin-1.0.0.tar.gz my-plugin/

# 发布到插件市场（或手动分发）
```

## 注意事项

1. **安全性**：插件代码运行在面板进程中，具有与面板相同的系统权限，安装第三方插件需谨慎
2. **性能**：监控采集插件的 Collect 方法应快速返回，避免阻塞采集循环
3. **错误处理**：插件方法应返回有意义的错误，面板会记录到日志并展示给用户
4. **资源清理**：插件的 Cleanup 方法应释放所有资源（关闭连接、停止 goroutine 等）
5. **版本兼容**：使用 `min_panel_version` 声明最低面板版本，避免 API 不兼容
6. **数据库**：插件如需自定义表，使用 GORM AutoMigrate，表名建议加插件前缀避免冲突
7. **前端**：前端扩展使用全局 `window.CloudPivot` 对象，避免与面板内部代码冲突
