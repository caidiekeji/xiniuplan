#!/usr/bin/env python3
"""犀牛面板一键部署 v5：单二进制（前端已嵌入）
本地交叉编译 linux/amd64 二进制 → 上传 → nginx 整站反代 → 重启 → 健康检查
"""
import os
import subprocess
import sys
import time

import paramiko

HOST = "192.168.31.154"
PORT = 22
USER = "jie"
PASSWORD = "12311"

ROOT = r"E:\www\fuwuqi"
BACKEND = os.path.join(ROOT, "backend")
BIN_LOCAL = os.path.join(BACKEND, "rhinopanel-linux")
BIN_REMOTE = "/home/jie/rhinopanel/backend/rhinopanel"

NGINX_SITE = r"""server {
    listen 1990;
    server_name _;

    # 单二进制架构：整站反代到 Go 面板（页面 + API + WebSocket 同源）
    location / {
        proxy_pass http://127.0.0.1:8888;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_read_timeout 86400;
        proxy_send_timeout 86400;
        client_max_body_size 512m;
    }
}
"""


def log(msg):
    print(f"[deploy] {msg}", flush=True)


def run(ssh, cmd, timeout=120):
    stdin, stdout, stderr = ssh.exec_command(cmd, timeout=timeout)
    out = stdout.read().decode("utf-8", errors="replace")
    err = stderr.read().decode("utf-8", errors="replace")
    code = stdout.channel.recv_exit_status()
    if out.strip():
        print(out.strip()[:1200])
    if err.strip() and "sudo] password" not in err:
        print(f"[stderr] {err.strip()[:300]}")
    return code, out, err


def run_sudo(ssh, cmd, timeout=60):
    return run(ssh, f"echo '{PASSWORD}' | sudo -S bash -c {repr(cmd)}", timeout)


def build_local():
    log("=== 本地构建（前端嵌入 + linux 交叉编译）===")
    # 1. 前端构建
    log("npm run build ...")
    r = subprocess.run("npm run build", cwd=os.path.join(ROOT, "frontend"),
                       shell=True, capture_output=True, text=True)
    if r.returncode != 0:
        log("前端构建失败：\n" + r.stdout[-800:] + r.stderr[-800:])
        sys.exit(1)
    # 2. dist 注入 embed 目录
    dist_src = os.path.join(ROOT, "frontend", "dist")
    dist_dst = os.path.join(BACKEND, "web", "dist")
    if os.path.isdir(dist_dst):
        import shutil
        shutil.rmtree(dist_dst)
    subprocess.run(f'xcopy "{dist_src}" "{dist_dst}\\" /E /I /Q /Y', shell=True,
                   capture_output=True)
    # 3. 交叉编译
    log("GOOS=linux go build ...")
    env = os.environ.copy()
    env.update({"GOOS": "linux", "GOARCH": "amd64", "CGO_ENABLED": "0",
                "GOPROXY": "https://goproxy.cn,direct"})
    r = subprocess.run('go build -ldflags="-s -w" -o rhinopanel-linux ./cmd/agent',
                       cwd=BACKEND, shell=True, capture_output=True, text=True, env=env)
    if r.returncode != 0:
        log("后端编译失败：\n" + r.stderr[-1200:])
        sys.exit(1)
    size = os.path.getsize(BIN_LOCAL) // 1024 // 1024
    log(f"编译完成：rhinopanel-linux ({size} MB，已嵌入前端)")


def main():
    build_local()

    log("\n=== 上传并部署到服务器 ===")
    ssh = paramiko.SSHClient()
    ssh.set_missing_host_key_policy(paramiko.AutoAddPolicy())
    ssh.connect(HOST, port=PORT, username=USER, password=PASSWORD, timeout=30)
    log("SSH 连接成功")
    sftp = ssh.open_sftp()

    try:
        # 1. 上传二进制
        log("上传二进制 ...")
        sftp.put(BIN_LOCAL, BIN_REMOTE + ".new")
        run(ssh, f"chmod +x {BIN_REMOTE}.new", timeout=10)

        # 2. 原子替换二进制 + 更新 nginx 配置
        log("更新 nginx 为整站反代 ...")
        run_sudo(ssh, f"mv {BIN_REMOTE}.new {BIN_REMOTE}")
        cfg_tmp = "/tmp/rhinopanel.nginx.conf"
        sftp.open(cfg_tmp, "w").write(NGINX_SITE)
        run_sudo(ssh, f"cp {cfg_tmp} /etc/nginx/sites-enabled/rhinopanel && nginx -t")
        run_sudo(ssh, "systemctl reload nginx")

        # 3. 重启后端
        log("重启后端 ...")
        run(ssh, "pkill -f 'rhinopanel' || true", timeout=10)
        time.sleep(2)
        run(ssh, "cd /home/jie/rhinopanel/backend && setsid ./rhinopanel > /tmp/rhinopanel.log 2>&1 &", timeout=10)
        time.sleep(3)
        code, out, _ = run(ssh, "pgrep -f 'rhinopanel' && echo RUNNING || echo NOT_FOUND", timeout=10)
        if "RUNNING" not in out:
            log("后端未启动，查看日志")
            run(ssh, "tail -30 /tmp/rhinopanel.log", timeout=10)
            sys.exit(1)

        # 4. 健康检查
        log("\n=== 健康检查 ===")
        time.sleep(2)
        code, out, _ = run(ssh, "curl -s http://127.0.0.1:8888/health", timeout=15)
        if '"ok"' in out:
            log("✅ 后端健康检查通过")
        code, out, _ = run(ssh, "curl -s -o /dev/null -w '%{http_code}' http://127.0.0.1:1990/", timeout=15)
        log(f"外部访问 1990: HTTP {out.strip()}")
        log("=== 部署完成 ===")
    finally:
        sftp.close()
        ssh.close()


if __name__ == "__main__":
    main()
