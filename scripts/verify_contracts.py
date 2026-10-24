#!/usr/bin/env python3
"""最后一批契约修复的线上实测：database / backup / ssl / alert

不依赖外部库，纯 urllib。逐条打印请求与响应要点。
"""
import json
import urllib.request
import urllib.error

BASE = "http://192.168.31.154:8888/api/v1"
USER, PASS = "admin", "Aa12311./"

TOKEN = None
PASSED = 0
FAILED = 0


def call(method, path, body=None, params=None):
    url = BASE + path
    if params:
        url += "?" + "&".join(f"{k}={v}" for k, v in params.items())
    data = None
    headers = {"Accept": "application/json"}
    if TOKEN:
        headers["Authorization"] = "Bearer " + TOKEN
    if body is not None:
        data = json.dumps(body).encode()
        headers["Content-Type"] = "application/json"
    req = urllib.request.Request(url, data=data, headers=headers, method=method)
    try:
        with urllib.request.urlopen(req, timeout=20) as r:
            payload = r.read().decode("utf-8", "replace")
            status = r.status
    except urllib.error.HTTPError as e:
        payload = e.read().decode("utf-8", "replace")
        status = e.code
    except Exception as e:
        return -1, {"message": str(e)}
    try:
        return status, json.loads(payload)
    except Exception:
        return status, payload


def check(title, ok, detail=""):
    global PASSED, FAILED
    if ok:
        PASSED += 1
        print(f"  [PASS] {title}  {detail}")
    else:
        FAILED += 1
        print(f"  [FAIL] {title}  {detail}")


def data_of(p):
    return p.get("data") if isinstance(p, dict) else p


def items_of(p):
    d = data_of(p)
    if isinstance(d, dict):
        return d.get("items") or d.get("list") or []
    return d if isinstance(d, list) else []


def main():
    global TOKEN
    st, p = call("POST", "/auth/login", {"username": USER, "password": PASS})
    if st != 200:
        print("登录失败:", st, p)
        return
    TOKEN = data_of(p).get("access_token")
    print(f"登录成功\n")

    # ================= 1. 数据库：类型错误映射 =================
    print("=== 1. 数据库：实例类型不支持应返回 400 而非 500 ===")
    for path in ["/databases/1/users", "/databases/1/dbs",
                 "/databases/1/variables", "/databases/1/status"]:
        st, p = call("GET", path)
        code = p.get("code") if isinstance(p, dict) else None
        check(f"{path} → 400/40001", st == 400 and code == 40001,
              f"HTTP {st} code {code}")

    # ================= 2. 数据库：分页字段名 =================
    print("\n=== 2. 数据库：分页响应使用 pageSize（camelCase）===")
    st, p = call("GET", "/databases", params={"page": 1, "pageSize": 20})
    d = data_of(p)
    check("列表返回 pageSize 字段", isinstance(d, dict) and "pageSize" in d,
          f"字段: {list(d.keys()) if isinstance(d, dict) else d}")

    # ================= 3. 导入 MySQL 实例 =================
    print("\n=== 3. 数据库：导入 MySQL 实例并验证 username 契约 ===")
    st, p = call("POST", "/databases/import", {
        "type": "mysql", "host": "127.0.0.1", "port": 3306,
        "username": "root", "password": "12311", "remark": "contract-verify",
    })
    check("导入 MySQL 实例", st == 200, f"HTTP {st}")
    if st != 200:
        return
    mid = data_of(p).get("id")

    st, p = call("GET", f"/databases/{mid}/users")
    users = items_of(p)
    keys = set()
    for u in users:
        if isinstance(u, dict):
            keys |= set(u.keys())
    check("用户列表字段为 username（非 user）",
          bool(users) and "username" in keys and "user" not in keys,
          f"字段集: {sorted(keys)}")

    # ================= 4. 创建用户 =================
    print("\n=== 4. 数据库：创建用户（CREATE USER 占位符修复）===")
    # 先删同名用户，保证脚本可重复运行（否则第二次报 1396 用户已存在）
    call("DELETE", f"/databases/{mid}/users/contract_test", params={"host": "%"})
    st, p = call("POST", f"/databases/{mid}/users",
                 {"username": "contract_test", "password": "Test@12345", "host": "%"})
    check("创建用户成功", st == 200 and isinstance(p, dict) and p.get("code") == 200,
          f"HTTP {st} {str(p)[:120]}")

    st, p = call("GET", f"/databases/{mid}/users")
    check("新用户出现在列表中",
          any(u.get("username") == "contract_test" for u in items_of(p) if isinstance(u, dict)))

    # ================= 5. 授权 =================
    print("\n=== 5. 数据库：授权 / 回收（GrantRequest 契约）===")
    st, p = call("POST", f"/databases/{mid}/users/contract_test/grant",
                 {"host": "%", "db_name": "app", "privileges": ["SELECT"]})
    check("授权成功（body 传 db_name）", st == 200, f"HTTP {st} {str(p)[:120]}")

    st, p = call("POST", f"/databases/{mid}/users/contract_test/grant",
                 {"username": "contract_test", "host": "%", "db_name": "app",
                  "privileges": ["SELECT"]})
    check("授权成功（body 冗余传 username）", st == 200, f"HTTP {st}")

    st, p = call("POST", f"/databases/{mid}/users/contract_test/grant",
                 {"host": "%", "db_name": "app", "privileges": ["DROP"]})
    check("非法权限被白名单拦截", st == 400, f"HTTP {st}")

    st, p = call("POST", f"/databases/{mid}/users/contract_test/revoke",
                 params={"host": "%", "db_name": "app"})
    check("回收权限成功（query 传 db_name）", st == 200, f"HTTP {st} {str(p)[:120]}")

    # ================= 6. 表结构 columns =================
    print("\n=== 6. 数据库：表结构返回 columns 数组 ===")
    st, p = call("GET", f"/databases/{mid}/dbs/app/tables")
    tables = items_of(p)
    check("表列表非空", bool(tables), f"表数 {len(tables)}")
    if tables:
        tname = tables[0].get("name")
        st, p = call("GET", f"/databases/{mid}/dbs/app/tables/{tname}")
        d = data_of(p)
        check("表结构含 columns 数组",
              isinstance(d, dict) and isinstance(d.get("columns"), list),
              f"表 {tname} 字段: {list(d.keys()) if isinstance(d, dict) else d}")

    # ================= 7. 备份 =================
    print("\n=== 7. 备份：状态值与分页 ===")
    st, p = call("GET", "/backups", params={"page": 1, "pageSize": 20})
    d = data_of(p)
    check("备份列表 200", st == 200, f"HTTP {st}")
    if isinstance(d, dict):
        check("备份分页使用 pageSize", "pageSize" in d, f"字段: {list(d.keys())}")
        bks = items_of(p)
        statuses = sorted({b.get("status") for b in bks if isinstance(b, dict)})
        check("备份状态值为 success/pending（无 completed）",
              all(s in ("success", "pending", "running", "failed") for s in statuses),
              f"状态集合: {statuses}")

    # ================= 8. 告警 =================
    print("\n=== 8. 告警：pageSize=100 被接受 ===")
    st, p = call("GET", "/alert/channels", params={"page": 1, "pageSize": 100})
    d = data_of(p)
    check("告警通道列表 200", st == 200, f"HTTP {st}")
    check("回显 pageSize=100", isinstance(d, dict) and d.get("pageSize") == 100,
          f"pageSize={d.get('pageSize') if isinstance(d, dict) else None}")

    st, p = call("GET", "/alert/rules", params={"page": 1, "pageSize": 50})
    check("告警规则列表 200", st == 200, f"HTTP {st}")

    # ================= 9. SSL（站点级） =================
    print("\n=== 9. SSL：站点级证书接口 ===")
    st, p = call("GET", "/sites", params={"page": 1, "pageSize": 50})
    sites = items_of(p)
    check("站点列表 200", st == 200, f"HTTP {st}")
    if sites:
        sid = sites[0].get("id")
        st, p = call("GET", f"/sites/{sid}/ssl")
        d = data_of(p)
        check("SSL 详情 200 且字段完整",
              st == 200 and isinstance(d, dict) and "enabled" in d and "provider" in d,
              f"字段: {list(d.keys()) if isinstance(d, dict) else str(d)[:120]}")

    # ================= 10. 类型错误映射（Postgres/Redis） =================
    print("\n=== 10. 数据库：错误码映射不影响正常路径 ===")
    st, p = call("GET", "/databases/discover")
    check("实例发现 200", st == 200, f"HTTP {st}")

    # ================= 清理测试数据 =================
    # 删除测试用户与测试实例，避免污染线上环境
    call("DELETE", f"/databases/{mid}/users/contract_test", params={"host": "%"})
    st, _ = call("DELETE", f"/databases/{mid}")
    if st == 200:
        print("\n（已清理测试用户与测试实例）")

    print("\n" + "=" * 60)
    print(f"结果：PASS {PASSED} / FAIL {FAILED}")
    print("=" * 60)


if __name__ == "__main__":
    main()
