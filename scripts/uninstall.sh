#!/bin/bash
# ============================================================
# CloudPivot 面板卸载脚本
# 支持: CentOS 7+ / Ubuntu 20.04+ / Debian 11+
# ============================================================

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# 安装配置（与 install.sh 保持一致）
INSTALL_DIR="/opt/cloudpivot"
BIN_PATH="/usr/local/bin/cloudpivot"
SERVICE_NAME="cloudpivot"
SERVICE_FILE="/etc/systemd/system/${SERVICE_NAME}.service"

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
# 1. 停止并禁用 systemd service
# ============================================================

stop_service() {
    info "停止 CloudPivot 服务..."

    if systemctl is-active --quiet "${SERVICE_NAME}" 2>/dev/null; then
        systemctl stop "${SERVICE_NAME}"
        success "服务已停止"
    else
        warn "服务未在运行"
    fi

    info "禁用开机自启..."
    if systemctl is-enabled --quiet "${SERVICE_NAME}" 2>/dev/null; then
        systemctl disable "${SERVICE_NAME}" 2>/dev/null || true
        success "已禁用开机自启"
    else
        warn "服务未设置开机自启"
    fi
}

# ============================================================
# 2. 删除 systemd unit 文件
# ============================================================

remove_service_file() {
    info "删除 systemd 服务文件..."

    if [ -f "${SERVICE_FILE}" ]; then
        rm -f "${SERVICE_FILE}"
        success "已删除: ${SERVICE_FILE}"
    else
        warn "服务文件不存在: ${SERVICE_FILE}"
    fi

    # 重载 systemd
    systemctl daemon-reload 2>/dev/null || true
    systemctl reset-failed 2>/dev/null || true
}

# ============================================================
# 3. 停止 cloudpivot 进程（兜底）
# ============================================================

kill_process() {
    info "检查并停止残留进程..."

    # 查找 cloudpivot 进程
    PIDS=$(pgrep -f "${BIN_PATH}" 2>/dev/null || true)
    if [ -n "${PIDS}" ]; then
        for pid in ${PIDS}; do
            kill "${pid}" 2>/dev/null || true
        done
        sleep 1
        # 强制杀死仍在运行的进程
        PIDS=$(pgrep -f "${BIN_PATH}" 2>/dev/null || true)
        if [ -n "${PIDS}" ]; then
            for pid in ${PIDS}; do
                kill -9 "${pid}" 2>/dev/null || true
            done
        fi
        success "残留进程已终止"
    else
        info "无残留进程"
    fi
}

# ============================================================
# 4. 询问是否删除数据
# ============================================================

confirm_data_removal() {
    echo ""
    warn "数据目录: ${INSTALL_DIR}/data"
    warn "包含: 数据库文件、配置、备份、日志等"
    echo ""
    read -r -p "是否删除所有数据？[y/N] (默认保留): " answer
    case "${answer}" in
        [yY]|[yY][eE][sS])
            REMOVE_DATA="yes"
            warn "将删除所有数据！此操作不可恢复！"
            ;;
        *)
            REMOVE_DATA="no"
            info "数据将被保留"
            ;;
    esac
}

# ============================================================
# 5. 删除二进制和目录
# ============================================================

remove_files() {
    info "删除二进制文件..."

    if [ -f "${BIN_PATH}" ]; then
        rm -f "${BIN_PATH}"
        success "已删除: ${BIN_PATH}"
    else
        warn "二进制文件不存在: ${BIN_PATH}"
    fi

    # 删除安装目录
    if [ -d "${INSTALL_DIR}" ]; then
        if [ "${REMOVE_DATA}" = "yes" ]; then
            rm -rf "${INSTALL_DIR}"
            success "已删除安装目录（含数据）: ${INSTALL_DIR}"
        else
            # 仅删除程序文件，保留 data 目录
            find "${INSTALL_DIR}" -mindepth 1 -maxdepth 1 ! -name "data" -exec rm -rf {} + 2>/dev/null || true
            success "已删除程序文件，数据目录已保留: ${INSTALL_DIR}/data"
        fi
    else
        warn "安装目录不存在: ${INSTALL_DIR}"
    fi
}

# ============================================================
# 6. 打印卸载完成
# ============================================================

print_uninstall_complete() {
    echo ""
    echo "============================================================"
    echo -e "${GREEN}  CloudPivot 面板卸载完成!${NC}"
    echo "============================================================"
    echo ""
    if [ "${REMOVE_DATA}" = "no" ] && [ -d "${INSTALL_DIR}/data" ]; then
        echo -e "  ${YELLOW}数据已保留在: ${INSTALL_DIR}/data${NC}"
        echo -e "  ${YELLOW}如需彻底删除，请手动执行: rm -rf ${INSTALL_DIR}${NC}"
    fi
    echo ""
    echo -e "  ${BLUE}已执行的操作:${NC}"
    echo "    1. 停止并禁用 systemd 服务"
    echo "    2. 删除 systemd unit 文件"
    echo "    3. 终止残留进程"
    echo "    4. 删除二进制文件"
    if [ "${REMOVE_DATA}" = "yes" ]; then
        echo "    5. 删除所有数据（含数据库、配置、备份）"
    else
        echo "    5. 数据目录已保留"
    fi
    echo ""
    echo "============================================================"
    echo ""
}

# ============================================================
# 主流程
# ============================================================

main() {
    echo ""
    echo "============================================================"
    echo "  CloudPivot 面板卸载脚本"
    echo "============================================================"
    echo ""

    check_root

    # 确认卸载
    read -r -p "确认要卸载 CloudPivot 面板吗？[y/N]: " confirm
    case "${confirm}" in
        [yY]|[yY][eE][sS])
            info "开始卸载..."
            ;;
        *)
            info "卸载已取消"
            exit 0
            ;;
    esac

    stop_service
    remove_service_file
    kill_process
    confirm_data_removal
    remove_files
    print_uninstall_complete
}

main "$@"
