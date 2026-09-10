# -*- coding: utf-8 -*-
"""前后端 API 对账：找出后端有路由但前端没调用、以及前端调用但后端不存在的接口"""
import re, os, json, io, sys
sys.stdout = io.TextIOWrapper(sys.stdout.buffer, encoding='utf-8')

BACKEND = r"E:\www\fuwuqi\backend"
FRONTEND = r"E:\www\fuwuqi\frontend\src"

# ---------- 1. 后端路由收集 ----------
go_files = []
for root, dirs, files in os.walk(BACKEND):
    dirs[:] = [d for d in dirs if d not in ('node_modules', 'data', 'runtime', 'sandbox', 'backup')]
    for f in files:
        if f.endswith('.go'):
            go_files.append(os.path.join(root, f))

routes = {}  # fullpath -> set(methods)
groups = {}  # varname -> prefix

def norm(prefix, path):
    if path in ('', '/'):
        p = prefix
    elif path.startswith(':'):
        p = prefix + '/{' + path[1:] + '}'
    else:
        p = prefix + path
    p = re.sub(r'/:([A-Za-z_]\w*)', r'/{\1}', p)
    p = re.sub(r'/+', '/', p)
    return p if p else '/'

for fp in go_files:
    try:
        src = open(fp, encoding='utf-8').read()
    except Exception:
        continue
    # group 定义（每行重置作用域太粗，简单按文件顺序覆盖）
    for m in re.finditer(r'(\w+)\s*:?=\s*\w+\.Group\("([^"]*)"', src):
        var, prefix = m.group(1), m.group(2)
        # 继承父组前缀：如 g := r.Group(...) 中 r 已知则拼接
        if var in ('r', 'v1'):
            groups[var] = norm('', prefix)
        else:
            base = groups.get(var, '')  # 不会命中，下面处理父 var
    # 更精确：捕获 "x := parentVar.Group("pattern"
    for m in re.finditer(r'(\w+)\s*:?=\s*(\w+)\.Group\("([^"]*)"\)', src):
        var, parent, prefix = m.group(1), m.group(2), m.group(3)
        base = groups.get(parent, '')
        groups[var] = norm(base, prefix)
    # v1 / r 顶层
    for m in re.finditer(r'(\w+)\s*:?=\s*gin\.New\(\)|(\w+)\s*:=\s*r\.Group\("([^"]*)"\)', src):
        pass
    # 路由注册: var.METHOD("path"
    for m in re.finditer(r'(\w+)\.(GET|POST|PUT|DELETE|PATCH|Any)\(\s*"([^"]*)"', src):
        var, method, path = m.group(1), m.group(2).upper(), m.group(3)
        if method == 'ANY':
            method = 'GET'  # 归一
        prefix = groups.get(var, '')
        # 顶层 gin engine r 通常无前缀；authGroup 等已算好
        full = norm(prefix, path)
        routes.setdefault(full, set()).add(method)

# 前缀可能因文件顺序有误，特判常见组
for var, expect in [('authGroup', ''), ('v1', '/api/v1')]:
    pass

# ---------- 2. 前端调用收集 ----------
fe_calls = {}  # normalized path -> set(methods)
url_re = re.compile(r"url:\s*[`'\"]([^`'\"]+)[`'\"]")
method_re = re.compile(r"method:\s*['\"](\w+)['\"]")

def norm_fe(u):
    u = re.sub(r'\$\{[^}]*\}', '{x}', u)
    u = re.sub(r'/\{x\}/\{x\}', '/{x}', u)
    if not u.startswith('/'):
        u = '/' + u
    return u

# axios baseURL='/api/v1'，前端 url 需补前缀
API_PREFIX = '/api/v1'

for root, dirs, files in os.walk(FRONTEND):
    dirs[:] = [d for d in dirs if d != 'node_modules']
    for f in files:
        if not (f.endswith('.ts') or f.endswith('.vue')):
            continue
        fp = os.path.join(root, f)
        try:
            src = open(fp, encoding='utf-8').read()
        except Exception:
            continue
        for m in url_re.finditer(src):
            u = m.group(1)
            method = 'GET'
            # 找该 url 块内的 method
            seg = src[m.start():m.start()+400]
            mm = method_re.search(seg)
            if mm:
                method = mm.group(1).upper()
            full = norm_fe(u)
            # 前端统一 /api/v1 前缀（axios baseURL）
            if not full.startswith(API_PREFIX):
                full = API_PREFIX + full
            fe_calls.setdefault(full, set()).add(method)

# websocket 前端调用（new WebSocket）
ws_re = re.compile(r"new WebSocket\([^)]*\)")
ws_files = []
for root, dirs, files in os.walk(FRONTEND):
    for f in files:
        if f.endswith(('.ts', '.vue')):
            fp = os.path.join(root, f)
            try:
                if ws_re.search(open(fp, encoding='utf-8').read()):
                    ws_files.append(os.path.relpath(fp, FRONTEND))
            except Exception:
                pass

# ---------- 3. 对账 ----------
def canon(p):
    # {x}/{x} 等占位符统一为 {*}，便于匹配
    return re.sub(r'\{[^}]*\}', '{*}', p)

be = {canon(k): sorted(v) for k, v in routes.items()}
fe = {canon(k): sorted(v) for k, v in fe_calls.items()}

print("=== 后端路由（去重后 %d 条）===" % len(be))
print("=== 前端调用（去重后 %d 条）===" % len(fe))

print("\n=== A. 前端调用了但后端没有对应路由（疑似断裂）===")
for p in sorted(fe):
    if p not in be:
        # 宽松匹配：后端有同结构但占位符数量不同
        loose = [b for b in be if re.sub(r'\{\*\}', '*', b) == re.sub(r'\{\*\}', '*', p)]
        tag = '  (结构疑似匹配: %s)' % loose[0] if loose else ''
        print("  %-55s %s%s" % (p, fe[p], tag))

print("\n=== B. 后端有路由但前端从未调用（功能不可达/未做前端）===")
for p in sorted(be):
    if p not in fe:
        # 排除 ws 和内部
        print("  %-55s %s" % (p, be[p]))
