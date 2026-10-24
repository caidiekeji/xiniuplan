package file

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
)

// SetPermission 修改文件/目录权限（chmod，F-005）
// mode 为八进制字符串，如 "644"、"0755"；recursive=true 时递归应用
func (m *Manager) SetPermission(path, mode string, recursive bool) error {
	abs, err := m.safePath(path)
	if err != nil {
		return err
	}
	mode = strings.TrimPrefix(mode, "0")
	parsed, err := strconv.ParseUint(mode, 8, 32)
	if err != nil || parsed > 0o7777 {
		return fmt.Errorf("无效的权限模式: %s（应为八进制，如 644）", mode)
	}
	fileMode := os.FileMode(parsed)

	if recursive {
		return filepath.Walk(abs, func(p string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			mode := fileMode
			if info.IsDir() {
				mode = fileMode | 0o111 // 目录默认保留执行位
			}
			return os.Chmod(p, mode)
		})
	}
	return os.Chmod(abs, fileMode)
}

// SetOwner 修改文件/目录属主（chown/chgrp，F-005）
// owner、group 传名称或数字 ID；recursive=true 时递归应用（仅 Unix 生效）
func (m *Manager) SetOwner(path, ownerName, groupName string, recursive bool) error {
	abs, err := m.safePath(path)
	if err != nil {
		return err
	}

	uid := -1
	if ownerName != "" {
		if u, err := user.Lookup(ownerName); err == nil {
			uid, _ = strconv.Atoi(u.Uid)
		} else if n, err := strconv.Atoi(ownerName); err == nil {
			uid = n
		} else {
			return fmt.Errorf("用户不存在: %s", ownerName)
		}
	}

	gid := -1
	if groupName != "" {
		if g, err := user.LookupGroup(groupName); err == nil {
			gid, _ = strconv.Atoi(g.Gid)
		} else if n, err := strconv.Atoi(groupName); err == nil {
			gid = n
		} else {
			return fmt.Errorf("用户组不存在: %s", groupName)
		}
	}

	if uid == -1 && gid == -1 {
		return fmt.Errorf("至少指定 owner 或 group 之一")
	}

	if recursive {
		return filepath.Walk(abs, func(p string, _ os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			return os.Chown(p, uid, gid)
		})
	}
	return os.Chown(abs, uid, gid)
}
