# 终端 API

Web 终端基于 WebSocket 协议实现，提供浏览器内的 SSH 终端功能。终端连接到 `localhost:22`，使用当前系统用户。

## 连接终端

`GET /api/v1/ws/terminal`

建立 WebSocket 连接，创建或恢复终端会话。

### 查询参数

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| session_id | string | 空 | 会话 ID，传入可恢复已有会话 |
| cols | int | 80 | 终端列数 |
| rows | int | 24 | 终端行数 |

### 连接示例（JavaScript）

```javascript
const ws = new WebSocket('ws://localhost:8888/api/v1/ws/terminal?cols=120&rows=40');

ws.onopen = () => {
  console.log('终端已连接');
};

ws.onmessage = (event) => {
  const msg = JSON.parse(event.data);
  if (msg.type === 'output') {
    terminal.write(msg.data);
  } else if (msg.type === 'error') {
    console.error(msg.data);
  }
};
```

### 连接成功响应

连接建立后，服务端发送会话创建确认消息：

```json
{
  "type": "output",
  "data": "\r\n[终端已连接] session_id=abc123def456\r\n"
}
```

恢复已有会话时发送：

```json
{
  "type": "output",
  "data": "\r\n[会话已恢复] abc123def456\r\n"
}
```

## 消息协议

客户端和服务端通过 JSON 格式消息通信。非 JSON 消息将被直接当作终端输入处理。

### 消息类型

| 类型 | 方向 | 说明 |
|------|------|------|
| input | 客户端→服务端 | 用户输入内容 |
| output | 服务端→客户端 | 终端输出内容 |
| resize | 客户端→服务端 | 调整终端窗口大小 |
| error | 服务端→客户端 | 错误信息 |
| close | 客户端→服务端 | 关闭会话 |

### 输入消息

客户端发送用户输入：

```json
{
  "type": "input",
  "data": "ls -la\n"
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| type | string | 固定为 "input" |
| data | string | 输入内容，包含换行符 `\n` 表示回车 |

也可以直接发送原始字节（非 JSON），将被直接写入终端 stdin。

### 输出消息

服务端返回终端输出：

```json
{
  "type": "output",
  "data": "total 24\ndrwxr-xr-x 2 root root 4096 Sep  9 10:00 .\n"
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| type | string | 固定为 "output" |
| data | string | 终端输出内容（含 ANSI 转义序列） |

### 调整窗口大小

```json
{
  "type": "resize",
  "cols": 120,
  "rows": 40
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| type | string | 固定为 "resize" |
| cols | int | 新的列数（大于 0） |
| rows | int | 新的行数（大于 0） |

### 错误消息

```json
{
  "type": "error",
  "data": "SSH 连接失败: dial tcp [::1]:22: connect: connection refused"
}
```

### 关闭会话

```json
{
  "type": "close"
}
```

## 会话管理

### 会话创建

- 每次 WebSocket 连接如未传入 `session_id`，将创建新会话
- 会话分配唯一的 `session_id`
- 系统建立 SSH 连接到 `localhost:22`
- 启动 PTY（伪终端），使用连接时指定的 cols/rows
- 启动输出流转发 goroutine，将 PTY 输出转发到 WebSocket

### 会话恢复

- 连接时传入有效的 `session_id` 可恢复已有会话
- 恢复条件：会话存在且未关闭
- 恢复后继续使用之前的 SSH 连接和 PTY
- 会话中运行的程序不会中断

### 会话清理

- 每 5 分钟执行一次会话清理检查
- 超过 30 分钟未活跃的会话将被自动关闭
- 活跃时间通过消息交互和心跳更新

## 心跳保活

- 服务端每 30 秒发送一次 WebSocket Ping 帧
- 客户端应自动响应 Pong 帧（标准 WebSocket 行为）
- 60 秒内未收到 Pong 则认为连接断开
- 收到 Pong 时更新会话活跃时间

## 完整使用示例

### xterm.js 集成示例

```javascript
import { Terminal } from 'xterm';
import { FitAddon } from 'xterm-addon-fit';

const term = new Terminal({ cursorBlink: true });
const fitAddon = new FitAddon();
term.loadAddon(fitAddon);
term.open(document.getElementById('terminal'));
fitAddon.fit();

const ws = new WebSocket(
  `ws://localhost:8888/api/v1/ws/terminal?cols=${term.cols}&rows=${term.rows}`
);

ws.onopen = () => {
  term.onData(data => {
    ws.send(JSON.stringify({ type: 'input', data }));
  });
};

ws.onmessage = (event) => {
  const msg = JSON.parse(event.data);
  if (msg.type === 'output' || msg.type === 'error') {
    term.write(msg.data);
  }
};

window.addEventListener('resize', () => {
  fitAddon.fit();
  ws.send(JSON.stringify({
    type: 'resize',
    cols: term.cols,
    rows: term.rows
  }));
});
```

## 注意事项

1. **认证**：WebSocket 连接需要携带有效的 JWT Token（通过查询参数 `token` 或 Cookie）
2. **SSH 连接**：终端连接到 `localhost:22`，确保 SSH 服务正在运行
3. **权限**：终端权限等同于运行面板的系统用户，请谨慎授予终端访问权限
4. **生产环境**：建议使用 WSS（WebSocket Secure）加密传输，配置 `CheckOrigin` 校验来源
5. **审计**：终端连接和操作会被审计日志记录
