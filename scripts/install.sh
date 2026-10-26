#!/bin/bash
# ============================================================
# 犀牛面板安装脚本
# 支持: CentOS 7+ / Ubuntu 20.04+ / Debian 11+
# 架构: x86_64 / arm64
# ============================================================

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 安装配置
INSTALL_DIR="/opt/rhinopanel"
BIN_PATH="/usr/local/bin/rhinopanel"
SERVICE_NAME="rhinopanel"
SERVICE_FILE="/etc/systemd/system/${SERVICE_NAME}.service"
# 面板监听端口：与 backend/internal/config/config.go 的默认值一致（8888）
PANEL_PORT="8888"
# GitHub Releases 下载地址（真实仓库：caidiekeji/xiniupanl）
# 产物由 .github/workflows/release.yml 打 tag 后自动构建发布，命名：
#   rhinopanel-linux-amd64 / rhinopanel-linux-arm64
DOWNLOAD_BASE_URL="https://github.com/caidiekeji/xiniupanl/releases/latest/download"
# 允许通过环境变量自定义安装参数：
#   ./install.sh 或 PANEL_PORT=9000 bash install.sh
PANEL_PORT="${PANEL_PORT:-8888}"

# ============================================================
# 工具函数
# ============================================================

info()    { echo -e "${BLUE}[INFO]${NC} $*"; }
success() { echo -e "${GREEN}[OK]${NC} $*"; }
warn()    { echo -e "${YELLOW}[WARN]${NC} $*"; }
error()   { echo -e "${RED}[ERROR]${NC} $*"; exit 1; }

# 检查是否为 root 用户
check_root() {
    if [ "$(id -u)" -ne 0 ]; then
        error "请使用 root 用户运行此脚本"
    fi
}

# ============================================================
# 1. 检测 OS 版本和架构
# ============================================================

detect_os() {
    info "检测操作系统版本..."

    ARCH=$(uname -m)
    case "${ARCH}" in
        x86_64|amd64)
            ARCH="amd64"
            ;;
        aarch64|arm64)
            ARCH="arm64"
            ;;
        *)
            error "不支持的架构: ${ARCH}，仅支持 x86_64 和 arm64"
            ;;
    esac

    if [ -f /etc/os-release ]; then
        . /etc/os-release
        OS_ID="${ID}"
        OS_VERSION="${VERSION_ID}"
    elif [ -f /etc/redhat-release ]; then
        OS_ID="centos"
        OS_VERSION=$(rpm -q --queryformat '%{VERSION}' centos-release 2>/dev/null || echo "7")
    else
        error "无法识别操作系统，请手动安装"
    fi

    # 版本校验
    case "${OS_ID}" in
        centos|rhel|rocky|almalinux)
            MAJOR_VER=$(echo "${OS_VERSION}" | cut -d. -f1)
            if [ "${MAJOR_VER}" -lt 7 ]; then
                error "CentOS/RHEL 版本过低，需要 7+，当前: ${OS_VERSION}"
            fi
            OS_FAMILY="rhel"
            ;;
        ubuntu)
            MAJOR_VER=$(echo "${OS_VERSION}" | cut -d. -f1)
            if [ "${MAJOR_VER}" -lt 20 ]; then
                error "Ubuntu 版本过低，需要 20.04+，当前: ${OS_VERSION}"
            fi
            OS_FAMILY="debian"
            ;;
        debian)
            MAJOR_VER=$(echo "${OS_VERSION}" | cut -d. -f1)
            if [ "${MAJOR_VER}" -lt 11 ]; then
                error "Debian 版本过低，需要 11+，当前: ${OS_VERSION}"
            fi
            OS_FAMILY="debian"
            ;;
        *)
            warn "未完全验证的系统: ${OS_ID} ${OS_VERSION}，尝试继续安装..."
            if command -v apt-get &>/dev/null; then
                OS_FAMILY="debian"
            elif command -v yum &>/dev/null || command -v dnf &>/dev/null; then
                OS_FAMILY="rhel"
            else
                error "无法确定包管理器，请手动安装"
            fi
            ;;
    esac

    success "操作系统: ${OS_ID} ${OS_VERSION} (${OS_FAMILY})，架构: ${ARCH}"
}

# ============================================================
# 2. 检测/安装 Docker
# ============================================================

install_docker() {
    info "检测 Docker 安装状态..."

    if command -v docker &>/dev/null; then
        DOCKER_VER=$(docker --version 2>/dev/null | awk '{print $3}' | tr -d ',')
        success "Docker 已安装: ${DOCKER_VER}"
    else
        info "正在安装 Docker..."
        if command -v curl &>/dev/null; then
            curl -fsSL https://get.docker.com | sh
        elif command -v wget &>/dev/null; then
            wget -qO- https://get.docker.com | sh
        else
            # 安装 curl
            if [ "${OS_FAMILY}" = "debian" ]; then
                apt-get update -qq && apt-get install -y -qq curl
            else
                yum install -y -q curl
            fi
            curl -fsSL https://get.docker.com | sh
        fi
        success "Docker 安装完成"
    fi

    # 启动 Docker 并设置开机自启
    info "启动 Docker 服务..."
    systemctl enable docker 2>/dev/null || true
    systemctl start docker 2>/dev/null || true

    # 验证 Docker 运行状态
    if docker info &>/dev/null; then
        success "Docker 服务运行正常"
    else
        warn "Docker 服务可能未正常启动，请手动检查: systemctl status docker"
    fi
}

# ============================================================
# 3. 创建安装目录
# ============================================================

create_directories() {
    info "创建安装目录..."
    mkdir -p "${INSTALL_DIR}/data"
    mkdir -p "${INSTALL_DIR}/backup"
    mkdir -p "${INSTALL_DIR}/logs"
    success "目录创建完成: ${INSTALL_DIR}"
}

# ============================================================
# 4. 下载二进制文件
# ============================================================

download_binary() {
    info "下载犀牛面板二进制文件..."

    BINARY_NAME="rhinopanel-linux-${ARCH}"
    DOWNLOAD_URL="${DOWNLOAD_BASE_URL}/${BINARY_NAME}"
    TEMP_FILE="/tmp/${BINARY_NAME}"

    info "下载地址: ${DOWNLOAD_URL}"

    if command -v curl &>/dev/null; then
        curl -fsSL -o "${TEMP_FILE}" "${DOWNLOAD_URL}"
    elif command -v wget &>/dev/null; then
        wget -q -O "${TEMP_FILE}" "${DOWNLOAD_URL}"
    else
        error "未找到 curl 或 wget，无法下载文件"
    fi

    if [ ! -f "${TEMP_FILE}" ] || [ ! -s "${TEMP_FILE}" ]; then
        error "二进制文件下载失败，请检查网络连接或下载地址"
    fi

    success "二进制文件下载完成"
}

# ============================================================
# 5. 安装二进制文件
# ============================================================

install_binary() {
    info "安装二进制文件..."

    BINARY_NAME="rhinopanel-linux-${ARCH}"
    TEMP_FILE="/tmp/${BINARY_NAME}"

    # 如果之前已下载则使用，否则重新下载
    if [ ! -f "${TEMP_FILE}" ]; then
        download_binary
    fi

    cp "${TEMP_FILE}" "${BIN_PATH}"
    chmod +x "${BIN_PATH}"
    rm -f "${TEMP_FILE}"

    # 验证
    if [ -x "${BIN_PATH}" ]; then
        success "二进制安装完成: ${BIN_PATH}"
    else
        error "二进制安装失败"
    fi
}

# ============================================================
# 6. 初始化数据库
# ============================================================

init_database() {
    info "初始化数据库（首次运行自动创建）..."
    # 数据库由程序首次启动时自动创建，这里仅确保数据目录存在
    mkdir -p "${INSTALL_DIR}/data"
    success "数据目录就绪: ${INSTALL_DIR}/data"
}

# ============================================================
# 7. 生成配置文件
# ============================================================

generate_config() {
    info "生成配置文件（端口 ${PANEL_PORT}）..."
    cat > "${INSTALL_DIR}/config.yaml" <<EOF
server:
  port: ${PANEL_PORT}
  mode: release
database:
  path: ./data/rhinopanel.db
EOF
    success "配置文件已生成: ${INSTALL_DIR}/config.yaml"
}

# ============================================================
# 8. 生成随机安全入口路径
# ============================================================

generate_security_path() {
    info "生成安全入口路径..."
    RANDOM_SUFFIX=$(head -c 6 /dev/urandom | base64 | tr -dc 'a-z0-9' | head -c 6)
    if [ -z "${RANDOM_SUFFIX}" ]; then
        RANDOM_SUFFIX=$(date +%s | sha256sum | tr -dc 'a-z0-9' | head -c 6)
    fi
    SECURITY_PATH="panel-${RANDOM_SUFFIX}"
    success "安全入口路径: /${SECURITY_PATH}"
}

# ============================================================
# 9. 创建 systemd service
# ============================================================

create_systemd_service() {
    info "创建 systemd 服务..."

    cat > "${SERVICE_FILE}" <<EOF
[Unit]
Description=Rhinopanel Server Management Panel
After=network.target docker.service
Wants=network.target docker.service

[Service]
Type=simple
WorkingDirectory=${INSTALL_DIR}
ExecStart=${BIN_PATH}
Restart=always
RestartSec=5
LimitNOFILE=65536
Environment=GIN_MODE=release
Environment=SECURITY_ENTRY=${SECURITY_PATH}

[Install]
WantedBy=multi-user.target
EOF

    success "systemd 服务文件已创建: ${SERVICE_FILE}"
}

# ============================================================
# 10. 启动服务
# ============================================================

start_service() {
    info "启动犀牛面板服务..."

    systemctl daemon-reload
    systemctl enable "${SERVICE_NAME}"
    systemctl start "${SERVICE_NAME}"

    # 等待服务启动
    sleep 3

    if systemctl is-active --quiet "${SERVICE_NAME}"; then
        success "犀牛面板服务启动成功"
    else
        warn "服务可能未正常启动，请检查: systemctl status ${SERVICE_NAME}"
        warn "查看日志: journalctl -u ${SERVICE_NAME} -f"
    fi
}

# ============================================================
# 11. 配置防火墙
# ============================================================

configure_firewall() {
    info "配置防火墙放行面板端口 ${PANEL_PORT}..."

    if command -v firewall-cmd &>/dev/null; then
        # firewalld (CentOS/RHEL)
        if systemctl is-active --quiet firewalld 2>/dev/null; then
            firewall-cmd --permanent --add-port="${PANEL_PORT}/tcp" 2>/dev/null || true
            firewall-cmd --reload 2>/dev/null || true
            success "firewalld 已放行端口 ${PANEL_PORT}"
        else
            warn "firewalld 未运行，跳过防火墙配置"
        fi
    elif command -v ufw &>/dev/null; then
        # ufw (Ubuntu/Debian)
        if ufw status | grep -q "Status: active"; then
            ufw allow "${PANEL_PORT}/tcp" 2>/dev/null || true
            success "ufw 已放行端口 ${PANEL_PORT}"
        else
            warn "ufw 未激活，跳过防火墙配置"
        fi
    else
        warn "未检测到 firewalld 或 ufw，请手动放行端口 ${PANEL_PORT}"
    fi
}

# ============================================================
# 12. 提取初始管理员密码
#
# 密码由程序在首次启动时随机生成（internal/service/auth InitAdmin），
# 打印在服务日志里 —— 不在脚本侧预生成，保证两者一致（否则脚本展示的
# 密码是摆设，真实密码藏在日志里，用户根本登不进去）。
# ============================================================

extract_admin_password() {
    info "提取初始管理员密码..."
    ADMIN_PASSWORD=""
    for _ in 1 2 3 4 5; do
        ADMIN_PASSWORD=$(journalctl -u "${SERVICE_NAME}" --since "5 minutes ago" --no-pager 2>/dev/null \
            | grep "初始管理员已创建" | tail -1 \
            | sed -n 's/.*admin \/ \([^，, ]*\).*/\1/p')
        [ -n "${ADMIN_PASSWORD}" ] && break
        sleep 2
    done
    if [ -n "${ADMIN_PASSWORD}" ]; then
        success "初始管理员密码已提取"
    else
        warn "未能自动提取初始密码，请手动执行以下命令查看："
        warn "  journalctl -u ${SERVICE_NAME} | grep '初始管理员'"
        ADMIN_PASSWORD="（见上方命令输出）"
    fi
}

# ============================================================
# 13. 打印安装信息
# ============================================================

print_install_info() {
    SERVER_IP=$(hostname -I 2>/dev/null | awk '{print $1}' || echo "YOUR_SERVER_IP")

    echo ""
    echo "============================================================"
    echo -e "${GREEN}  犀牛面板安装完成!${NC}"
    echo "============================================================"
    echo ""
    echo -e "  ${BLUE}访问地址:${NC}"
    echo -e "    http://${SERVER_IP}:${PANEL_PORT}/${SECURITY_PATH}"
    echo ""
    echo -e "  ${BLUE}初始账号:${NC}"
    echo -e "    用户名: admin"
    echo -e "    密  码: ${ADMIN_PASSWORD}"
    echo ""
    echo -e "  ${YELLOW}重要提示:${NC}"
    echo -e "    1. 请妥善保管初始密码，首次登录后将强制修改"
    echo -e "    2. 安全入口路径用于隐藏面板地址，请牢记"
    echo -e "    3. 如无法访问，请检查防火墙是否放行端口 ${PANEL_PORT}"
    echo -e "    4. 服务管理命令: systemctl {start|stop|restart|status} ${SERVICE_NAME}"
    echo -e "    5. 查看日志: journalctl -u ${SERVICE_NAME} -f"
    echo ""
    echo "============================================================"
    echo ""

    # 保存安装信息到文件
    INFO_FILE="${INSTALL_DIR}/install_info.txt"
    cat > "${INFO_FILE}" <<EOF
犀牛面板安装信息
====================
安装时间: $(date '+%Y-%m-%d %H:%M:%S')
访问地址: http://${SERVER_IP}:${PANEL_PORT}/${SECURITY_PATH}
管理员账号: admin
管理员密码: ${ADMIN_PASSWORD}
安全入口: /${SECURITY_PATH}
面板端口: ${PANEL_PORT}
安装目录: ${INSTALL_DIR}
二进制路径: ${BIN_PATH}
EOF
    chmod 600 "${INFO_FILE}"
    info "安装信息已保存到: ${INFO_FILE}"
}

# ============================================================
# 主流程
# ============================================================

main() {
    echo ""
    echo "============================================================"
    echo "  犀牛面板安装脚本"
    echo "  支持: CentOS 7+ / Ubuntu 20.04+ / Debian 11+"
    echo "============================================================"
    echo ""

    check_root
    detect_os
    install_docker
    create_directories
    download_binary
    install_binary
    init_database
    generate_security_path
    create_systemd_service
    start_service
    # 从 journalctl 提取程序随机生成的初始密码，必须在服务启动之后
    extract_admin_password
    configure_firewall
    print_install_info
}

main "$@"
