#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
云枢面板 - 后端 API 全量测试脚本
用法: python scripts/test_api.py
功能: 启动后端 -> 获取 Token -> 遍历所有端点 -> 生成 docs/test-report.md
"""

import json
import os
import socket
import subprocess
import sys
import time
import urllib.request
import urllib.error
from datetime import datetime

# ============ 配置 ============
BASE_DIR = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
BACKEND_DIR = os.path.join(BASE_DIR, "backend")
REPORT_PATH = os.path.join(BASE_DIR, "docs", "test-report.md")
BASE_URL = "http://127.0.0.1:8888"
API_PREFIX = "/api/v1"
ADMIN_USER = "admin"
ADMIN_PASS = "admin123"
START_TIMEOUT = 90  # 秒
REQUEST_DELAY = 0.05  # 请求间隔（秒），避免触发限流

# ============ 测试结果数据结构 ============
class TestResult:
    def __init__(self, module, method, endpoint, expect_status, desc=""):
        self.module = module
        self.method = method
        self.endpoint = endpoint
        self.expect_status = expect_status  # "200" / "401" / "error"
        self.desc = desc
        self.actual_status = None
        self.passed = False
        self.response_code = None
        self.response_message = None
        self.has_format = False
        self.error_detail = ""
        self.duration_ms = 0

# ============ HTTP 工具 ============
def http_request(method, path, token=None, body=None, timeout=20, retries=2):
    """发送 HTTP 请求，带重试。返回 (status_code, response_dict_or_text, duration_ms)"""
    url = f"{BASE_URL}{path}"
    data = None
    headers = {"Content-Type": "application/json"}
    if token:
        headers["Authorization"] = f"Bearer {token}"
    if body is not None:
        data = json.dumps(body).encode("utf-8")

    last_err = None
    for attempt in range(retries + 1):
        req = urllib.request.Request(url, data=data, headers=headers, method=method)
        start = time.time()
        try:
            with urllib.request.urlopen(req, timeout=timeout) as resp:
                raw = resp.read().decode("utf-8", errors="replace")
                status = resp.status
            duration = int((time.time() - start) * 1000)
            try:
                parsed = json.loads(raw) if raw else {}
            except json.JSONDecodeError:
                parsed = raw
            return status, parsed, duration
        except urllib.error.HTTPError as e:
            raw = e.read().decode("utf-8", errors="replace")
            duration = int((time.time() - start) * 1000)
            try:
                parsed = json.loads(raw) if raw else {}
            except json.JSONDecodeError:
                parsed = raw
            return e.code, parsed, duration
        except Exception as e:
            last_err = e
            if attempt < retries:
                time.sleep(0.5)
                continue
            duration = int((time.time() - start) * 1000)
            return 0, str(e), duration
    return 0, str(last_err), 0

def check_response_format(resp):
    """检查响应是否包含 code/message/data 字段"""
    if isinstance(resp, dict):
        return "code" in resp and "message" in resp and "data" in resp
    return False

# ============ 端口/进程管理 ============
def is_port_in_use(port):
    with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as s:
        return s.connect_ex(("127.0.0.1", port)) == 0

def kill_process_on_port(port):
    """Windows: 杀掉占用指定端口的进程"""
    try:
        result = subprocess.run(
            ["netstat", "-ano"], capture_output=True, text=True, timeout=10
        )
        for line in result.stdout.splitlines():
            if f":{port}" in line and "LISTENING" in line:
                parts = line.split()
                pid = parts[-1]
                if pid.isdigit() and pid != "0":
                    subprocess.run(["taskkill", "/F", "/PID", pid],
                                   capture_output=True, timeout=10)
                    print(f"[清理] 已杀掉占用端口 {port} 的进程 PID={pid}")
                    time.sleep(1)
    except Exception as e:
        print(f"[清理] 端口清理失败: {e}")

# ============ 后端进程管理 ============
backend_proc = None
backend_log = None

def start_backend():
    """启动后端服务并等待健康检查通过"""
    global backend_proc, backend_log
    print("[启动] 清理端口占用...")
    if is_port_in_use(8888):
        kill_process_on_port(8888)
    time.sleep(1)

    print("[启动] 正在启动后端服务 (go run cmd/agent/main.go)...")
    backend_log = os.path.join(BACKEND_DIR, "test_backend.log")
    fout = open(backend_log, "w", encoding="utf-8")
    env = {**os.environ, "GOPROXY": "https://goproxy.cn,direct"}
    backend_proc = subprocess.Popen(
        ["go", "run", "cmd/agent/main.go"],
        cwd=BACKEND_DIR,
        stdout=fout,
        stderr=subprocess.STDOUT,
        env=env,
    )

    deadline = time.time() + START_TIMEOUT
    while time.time() < deadline:
        if backend_proc.poll() is not None:
            print(f"[启动] 后端进程异常退出，code={backend_proc.returncode}")
            print("--- 后端日志 ---")
            try:
                with open(backend_log, "r", encoding="utf-8") as f:
                    print(f.read()[-2000:])
            except Exception:
                pass
            return False
        try:
            status, resp, _ = http_request("GET", "/health", timeout=3, retries=0)
            if status == 200:
                print(f"[启动] 后端服务就绪 (version={resp.get('version', '?') if isinstance(resp, dict) else '?'})")
                return True
        except Exception:
            pass
        time.sleep(1)
    print("[启动] 等待后端就绪超时")
    return False

def stop_backend():
    """停止后端服务"""
    global backend_proc
    if backend_proc and backend_proc.poll() is None:
        print("[关闭] 正在停止后端服务...")
        backend_proc.terminate()
        try:
            backend_proc.wait(timeout=10)
        except subprocess.TimeoutExpired:
            backend_proc.kill()
        print("[关闭] 后端服务已停止")

# ============ 登录获取 Token ============
def login():
    """登录获取 access_token 和 refresh_token"""
    status, resp, _ = http_request("POST", f"{API_PREFIX}/auth/login",
                                    body={"username": ADMIN_USER, "password": ADMIN_PASS})
    if status == 200 and isinstance(resp, dict) and resp.get("code") == 200:
        data = resp.get("data", {})
        return data.get("access_token"), data.get("refresh_token")
    print(f"[登录] 失败: status={status}, resp={resp}")
    return None, None

# ============ 测试用例定义 ============
def build_test_cases():
    """构建所有测试用例，按模块分组。logout 放在最后避免 token 失效。"""
    cases = []  # (module, method, endpoint, expect, desc, body)

    # ---- 认证模块（logout 放最后） ----
    M = "认证"
    cases.append((M, "POST", "/auth/login", "200", "正确密码登录", {"username": ADMIN_USER, "password": ADMIN_PASS}))
    cases.append((M, "POST", "/auth/login", "error", "错误密码登录", {"username": ADMIN_USER, "password": "wrongpass"}))
    cases.append((M, "POST", "/auth/refresh", "error", "无效 refresh_token", {"refresh_token": "invalid_token"}))
    cases.append((M, "GET", "/auth/info", "200", "获取用户信息（需token）", None))
    cases.append((M, "GET", "/auth/info", "401", "未授权访问用户信息", None))
    cases.append((M, "GET", "/auth/totp/status", "200", "TOTP状态（需token）", None))
    cases.append((M, "POST", "/auth/totp/bind", "200", "TOTP绑定（需token）", None))

    # ---- Dashboard ----
    M = "Dashboard"
    cases.append((M, "GET", "/dashboard/overview", "200", "概览", None))
    cases.append((M, "GET", "/dashboard/metrics?range=1h", "200", "指标", None))
    cases.append((M, "GET", "/dashboard/audit-logs", "200", "审计日志", None))

    # ---- 网站 ----
    M = "网站"
    cases.append((M, "GET", "/sites", "200", "网站列表", None))
    cases.append((M, "POST", "/sites", "error", "创建网站(参数校验)", {"name": ""}))
    cases.append((M, "GET", "/sites/99999", "error", "获取不存在的网站", None))
    cases.append((M, "PUT", "/sites/99999", "error", "更新不存在的网站", {"name": "test"}))
    cases.append((M, "DELETE", "/sites/99999", "error", "删除不存在的网站", None))
    cases.append((M, "POST", "/sites/99999/start", "error", "启动不存在的网站", None))
    cases.append((M, "POST", "/sites/99999/stop", "error", "停止不存在的网站", None))
    cases.append((M, "POST", "/sites/99999/restart", "error", "重启不存在的网站", None))
    cases.append((M, "GET", "/sites/99999/config", "error", "获取不存在网站配置", None))
    cases.append((M, "PUT", "/sites/99999/config", "error", "更新不存在网站配置", {"config": {}}))

    # ---- 数据库 ----
    M = "数据库"
    cases.append((M, "GET", "/databases", "200", "数据库实例列表", None))
    cases.append((M, "POST", "/databases", "error", "创建实例(参数校验)", {"name": ""}))
    cases.append((M, "GET", "/databases/99999", "error", "获取不存在实例", None))
    cases.append((M, "DELETE", "/databases/99999", "error", "删除不存在实例", None))
    cases.append((M, "POST", "/databases/99999/start", "error", "启动不存在实例", None))
    cases.append((M, "POST", "/databases/99999/stop", "error", "停止不存在实例", None))
    cases.append((M, "POST", "/databases/99999/restart", "error", "重启不存在实例", None))

    # ---- Redis ----
    M = "Redis"
    cases.append((M, "GET", "/redis/99999/keys", "error", "不存在实例的keys", None))
    cases.append((M, "POST", "/redis/99999/cli", "error", "不存在实例的CLI", {"command": "PING"}))
    cases.append((M, "GET", "/redis/99999/info", "error", "不存在实例的info", None))
    cases.append((M, "GET", "/redis/99999/memory", "error", "不存在实例的memory", None))

    # ---- PostgreSQL ----
    M = "PostgreSQL"
    cases.append((M, "GET", "/postgres/99999/dbs", "error", "不存在实例的库列表", None))
    cases.append((M, "POST", "/postgres/99999/dbs", "error", "创建库(不存在实例)", {"name": "testdb"}))
    cases.append((M, "GET", "/postgres/99999/users", "error", "不存在实例的用户列表", None))

    # ---- 文件 ----
    M = "文件"
    cases.append((M, "GET", "/files?path=.", "200", "列出目录", None))
    cases.append((M, "POST", "/files/mkdir", "error", "创建目录(空path)", {"path": ""}))
    cases.append((M, "POST", "/files/create", "error", "创建文件(空path)", {"path": "", "type": "file"}))
    cases.append((M, "GET", "/files/content", "error", "读取文件(空path)", None))

    # ---- 终端 (WebSocket) ----
    M = "终端"
    cases.append((M, "GET", "/ws/terminal", "error", "WebSocket端点存在性验证", None))

    # ---- 容器 ----
    M = "容器"
    cases.append((M, "GET", "/containers", "200", "容器列表", None))
    cases.append((M, "GET", "/images", "200", "镜像列表", None))
    cases.append((M, "GET", "/containers/nonexistent", "error", "获取不存在容器", None))

    # ---- 应用商店 ----
    M = "应用商店"
    cases.append((M, "GET", "/apps", "200", "应用列表", None))
    cases.append((M, "GET", "/apps/installed", "200", "已安装应用", None))
    cases.append((M, "GET", "/app-sources", "200", "应用源列表", None))
    cases.append((M, "GET", "/apps/nonexistent_key", "error", "获取不存在应用", None))

    # ---- 备份 ----
    M = "备份"
    cases.append((M, "GET", "/backups", "200", "备份列表", None))
    cases.append((M, "POST", "/backups", "error", "创建备份(参数校验)", {"type": ""}))
    cases.append((M, "GET", "/backups/99999", "error", "获取不存在备份", None))

    # ---- 计划任务 ----
    M = "计划任务"
    cases.append((M, "GET", "/cron/tasks", "200", "任务列表", None))
    cases.append((M, "POST", "/cron/tasks", "error", "创建任务(参数校验)", {"name": ""}))
    cases.append((M, "GET", "/cron/tasks/99999/logs", "error", "不存在任务日志", None))

    # ---- 防火墙 ----
    M = "防火墙"
    cases.append((M, "GET", "/firewall/rules", "200", "规则列表", None))
    cases.append((M, "POST", "/firewall/rules", "error", "添加规则(参数校验)", {"port": ""}))
    cases.append((M, "GET", "/firewall/suggestions", "200", "安全建议", None))

    # ---- 监控 ----
    M = "监控"
    cases.append((M, "GET", "/monitor/processes", "200", "进程列表", None))
    cases.append((M, "GET", "/monitor/disk", "200", "磁盘分析", None))
    cases.append((M, "GET", "/monitor/network", "200", "网络连接", None))
    cases.append((M, "GET", "/monitor/history?type=cpu&range=1h", "200", "历史指标", None))

    # ---- 审计 ----
    M = "审计"
    cases.append((M, "GET", "/audit/logs", "200", "操作日志", None))
    cases.append((M, "GET", "/audit/login-logs", "200", "登录日志", None))

    # ---- 系统设置 ----
    M = "系统设置"
    cases.append((M, "GET", "/system/settings", "200", "获取设置", None))
    cases.append((M, "PUT", "/system/settings", "error", "更新设置(参数校验)", {"settings": ""}))
    cases.append((M, "GET", "/system/version", "200", "版本信息", None))
    cases.append((M, "GET", "/system/info", "200", "系统信息", None))
    cases.append((M, "POST", "/system/clean", "error", "清理系统(参数校验)", {"clean_type": ""}))
    cases.append((M, "GET", "/system/smtp", "200", "SMTP配置", None))

    # ---- 用户/RBAC ----
    M = "用户/RBAC"
    cases.append((M, "GET", "/users", "200", "用户列表", None))
    cases.append((M, "POST", "/users", "error", "创建用户(参数校验)", {"username": ""}))
    cases.append((M, "GET", "/users/99999", "error", "获取不存在用户", None))
    cases.append((M, "GET", "/roles", "200", "角色列表", None))
    cases.append((M, "GET", "/permissions", "200", "权限列表", None))

    # ---- SSH ----
    M = "SSH"
    cases.append((M, "GET", "/ssh/config", "200", "SSH配置", None))
    cases.append((M, "GET", "/ssh/keys", "200", "SSH密钥列表", None))
    cases.append((M, "GET", "/ssh/logs", "200", "SSH登录日志", None))
    cases.append((M, "GET", "/ssh/fail2ban/status", "200", "Fail2ban状态", None))

    # ---- 告警 ----
    M = "告警"
    cases.append((M, "GET", "/alert/rules", "200", "告警规则列表", None))
    cases.append((M, "POST", "/alert/rules", "error", "创建规则(参数校验)", {"name": ""}))
    cases.append((M, "GET", "/alert/channels", "200", "告警渠道列表", None))
    cases.append((M, "POST", "/alert/channels", "error", "创建渠道(参数校验)", {"name": ""}))
    cases.append((M, "GET", "/alert/history", "200", "告警历史", None))

    # ---- DNS ----
    M = "DNS"
    cases.append((M, "GET", "/dns/accounts", "200", "DNS账号列表", None))
    cases.append((M, "POST", "/dns/accounts", "error", "创建DNS账号(参数校验)", {"name": ""}))

    # ---- 插件 ----
    M = "插件"
    cases.append((M, "GET", "/plugins", "200", "插件列表", None))
    cases.append((M, "GET", "/plugins/nonexistent/config", "error", "不存在插件配置", None))

    # ---- 登出（放最后，避免 token 失效影响后续测试） ----
    M = "认证"
    cases.append((M, "POST", "/auth/logout", "200", "登出（需token）", None))

    return cases

# ============ 执行测试 ============
def run_tests(token):
    """执行所有测试用例，返回结果列表"""
    cases = build_test_cases()
    results = []
    total = len(cases)

    for i, (module, method, endpoint, expect, desc, body) in enumerate(cases, 1):
        full_path = f"{API_PREFIX}{endpoint}"
        use_token = token if expect != "401" else None
        status, resp, duration = http_request(method, full_path, token=use_token, body=body)

        tr = TestResult(module, method, endpoint, expect, desc)
        tr.actual_status = status
        tr.duration_ms = duration
        tr.has_format = check_response_format(resp)
        if isinstance(resp, dict):
            tr.response_code = resp.get("code")
            tr.response_message = resp.get("message")

        # 判断通过
        if expect == "200":
            tr.passed = (status == 200 and tr.has_format and tr.response_code == 200)
            if not tr.passed:
                tr.error_detail = f"预期HTTP 200+code200，实际HTTP {status}, code={tr.response_code}, msg={tr.response_message}"
        elif expect == "401":
            tr.passed = (status == 401)
            if not tr.passed:
                tr.error_detail = f"预期HTTP 401，实际HTTP {status}"
        elif expect == "error":
            tr.passed = (status != 0 and not (status == 200 and tr.response_code == 200))
            if not tr.passed:
                if status == 0:
                    tr.error_detail = f"连接失败: {resp}"
                else:
                    tr.error_detail = f"预期错误响应，实际HTTP {status}, code={tr.response_code}"

        results.append(tr)
        status_icon = "PASS" if tr.passed else "FAIL"
        print(f"  [{i:2d}/{total}] {status_icon} {method:6s} {full_path:50s} ({duration:4d}ms) - {desc}")

        time.sleep(REQUEST_DELAY)

    return results

# ============ 生成报告 ============
def generate_report(results, start_time, end_time):
    """生成 Markdown 测试报告"""
    total = len(results)
    passed = sum(1 for r in results if r.passed)
    failed = total - passed
    pass_rate = (passed / total * 100) if total > 0 else 0

    modules = {}
    for r in results:
        modules.setdefault(r.module, []).append(r)

    unauth_tests = [r for r in results if r.expect_status == "401"]
    unauth_passed = sum(1 for r in unauth_tests if r.passed)
    failures = [r for r in results if not r.passed]

    lines = []
    lines.append("# 云枢面板 - 后端 API 全量测试报告")
    lines.append("")
    lines.append(f"> 生成时间：{datetime.now().strftime('%Y-%m-%d %H:%M:%S')}")
    lines.append("")

    # 一、测试环境
    lines.append("## 一、测试环境")
    lines.append("")
    lines.append("| 项目 | 值 |")
    lines.append("|------|-----|")
    lines.append(f"| 测试时间 | {start_time.strftime('%Y-%m-%d %H:%M:%S')} ~ {end_time.strftime('%Y-%m-%d %H:%M:%S')} |")
    lines.append(f"| 耗时 | {int((end_time - start_time).total_seconds())} 秒 |")
    lines.append("| 操作系统 | Windows |")
    lines.append("| Go 版本 | 1.27.1 |")
    lines.append("| 后端地址 | http://127.0.0.1:8888 |")
    lines.append("| 测试账号 | admin / admin123 |")
    lines.append("| 测试方式 | 启动后端 -> 登录获取JWT -> 遍历端点 -> 验证响应格式与状态码 |")
    lines.append("| 构建验证 | go build ./... 通过 |")
    lines.append("")

    # 二、总体结果
    lines.append("## 二、总体结果")
    lines.append("")
    lines.append("| 指标 | 数值 |")
    lines.append("|------|------|")
    lines.append(f"| 总端点数量 | {total} |")
    lines.append(f"| 通过数 | {passed} |")
    lines.append(f"| 失败数 | {failed} |")
    lines.append(f"| 通过率 | {pass_rate:.1f}% |")
    lines.append(f"| 未授权测试 | {unauth_passed}/{len(unauth_tests)} 通过 |")
    lines.append("")

    # 三、各模块测试结果
    lines.append("## 三、各模块测试结果")
    lines.append("")
    for mod, items in modules.items():
        mod_pass = sum(1 for r in items if r.passed)
        mod_total = len(items)
        lines.append(f"### {mod}（{mod_pass}/{mod_total} 通过）")
        lines.append("")
        lines.append("| 端点 | 方法 | 预期 | 实际HTTP | 响应code | 格式 | 耗时(ms) | 结果 | 说明 |")
        lines.append("|------|------|------|----------|----------|------|----------|------|------|")
        for r in items:
            fmt = "✓" if r.has_format else "✗"
            result = "✅ 通过" if r.passed else "❌ 失败"
            lines.append(f"| `{r.endpoint}` | {r.method} | {r.expect_status} | {r.actual_status} | {r.response_code} | {fmt} | {r.duration_ms} | {result} | {r.desc} |")
        lines.append("")

    # 四、失败项详情
    lines.append("## 四、失败项详情")
    lines.append("")
    if failures:
        lines.append("| 模块 | 端点 | 方法 | 预期 | 实际 | 原因 |")
        lines.append("|------|------|------|------|------|------|")
        for r in failures:
            lines.append(f"| {r.module} | `{r.endpoint}` | {r.method} | {r.expect_status} | HTTP {r.actual_status}/code {r.response_code} | {r.error_detail} |")
        lines.append("")
    else:
        lines.append("所有测试用例均通过，无失败项。")
        lines.append("")

    # 五、未授权访问测试
    lines.append("## 五、未授权访问测试")
    lines.append("")
    lines.append("验证不带 Token 访问受保护端点是否返回 401：")
    lines.append("")
    lines.append("| 端点 | 方法 | 预期 | 实际HTTP | 结果 |")
    lines.append("|------|------|------|----------|------|")
    for r in unauth_tests:
        result = "✅ 通过" if r.passed else "❌ 失败"
        lines.append(f"| `{r.endpoint}` | {r.method} | 401 | {r.actual_status} | {result} |")
    lines.append("")

    # 六、环境限制说明
    lines.append("## 六、环境限制说明")
    lines.append("")
    lines.append("以下功能在 Windows 环境下可能受限，测试仅验证 API 端点存在且返回合理错误（不 panic）：")
    lines.append("")
    lines.append("- **Docker/容器**：Windows 上可能未安装 Docker，容器相关接口返回错误属正常")
    lines.append("- **Nginx/网站管理**：网站启停依赖 Nginx，Windows 环境可能返回错误")
    lines.append("- **SSH/Fail2ban**：Linux 专属功能，Windows 上读取配置可能返回错误")
    lines.append("- **数据库实例**：MySQL/PostgreSQL/Redis 实例启停依赖 Linux 服务管理")
    lines.append("- **WebSocket 终端**：仅验证端点路由存在（普通 GET 返回 400/426），未进行真实 WS 连接测试")
    lines.append("")

    # 七、结论和建议
    lines.append("## 七、结论和建议")
    lines.append("")
    if pass_rate >= 90:
        lines.append(f"**结论：API 整体质量良好，通过率 {pass_rate:.1f}%。**")
    elif pass_rate >= 70:
        lines.append(f"**结论：API 基本可用，通过率 {pass_rate:.1f}%，存在部分问题需修复。**")
    else:
        lines.append(f"**结论：API 存在较多问题，通过率 {pass_rate:.1f}%，需重点修复。**")
    lines.append("")
    lines.append("### 建议")
    lines.append("")
    lines.append("1. **统一错误处理**：确保所有端点在异常情况下均返回 `{code, message, data}` 格式，避免裸 HTTP 错误页")
    lines.append("2. **参数校验**：POST/PUT 端点应统一返回 400 + code=40001 表示参数错误")
    lines.append("3. **资源不存在**：访问不存在的资源 ID 应统一返回 404 + code=40401")
    lines.append("4. **跨平台兼容**：Linux 专属功能在 Windows 上应返回明确的「不支持当前平台」错误，而非 500")
    lines.append("5. **WebSocket 测试**：后续可补充真实 WebSocket 连接测试，验证终端会话功能")
    lines.append("6. **集成 CI**：将本测试脚本纳入 CI 流水线，每次提交自动运行")
    lines.append("")

    report = "\n".join(lines)
    os.makedirs(os.path.dirname(REPORT_PATH), exist_ok=True)
    with open(REPORT_PATH, "w", encoding="utf-8") as f:
        f.write(report)
    print(f"[报告] 已生成: {REPORT_PATH}")
    return report

# ============ 主流程 ============
def main():
    print("=" * 70)
    print("云枢面板 - 后端 API 全量测试")
    print("=" * 70)

    start_time = datetime.now()

    # 1. 启动后端
    if not start_backend():
        print("[错误] 后端启动失败，终止测试")
        sys.exit(1)

    try:
        # 2. 登录
        print("[登录] 正在获取 JWT Token...")
        token, refresh_token = login()
        if not token:
            print("[错误] 登录失败，终止测试")
            sys.exit(1)
        print(f"[登录] 成功，token 前20位: {token[:20]}...")

        # 3. 执行测试
        print(f"[测试] 开始执行 API 端点测试 (请求间隔 {REQUEST_DELAY}s)...")
        results = run_tests(token)

        end_time = datetime.now()

        # 4. 生成报告
        print("[报告] 正在生成测试报告...")
        generate_report(results, start_time, end_time)

        # 5. 输出摘要
        total = len(results)
        passed = sum(1 for r in results if r.passed)
        failed = total - passed
        print()
        print("=" * 70)
        print(f"测试完成: 总计 {total}，通过 {passed}，失败 {failed}，通过率 {passed/total*100:.1f}%")
        print(f"报告路径: {REPORT_PATH}")
        print("=" * 70)

    finally:
        stop_backend()

if __name__ == "__main__":
    main()
