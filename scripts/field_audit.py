# -*- coding: utf-8 -*-
"""字段级 API 对账：后端 dto required 字段 vs 前端请求载荷"""
import re, os, io, sys
sys.stdout = io.TextIOWrapper(sys.stdout.buffer, encoding='utf-8')

BACKEND = os.path.join(os.path.dirname(__file__), '..', 'backend', 'internal')
FRONTEND = os.path.join(os.path.dirname(__file__), '..', 'frontend', 'src')

QUOTE = "`'\""  # 前端 url 的三种引号

# 1. dto 字段（json tag + required）
dto_fields = {}
dto_dir = os.path.join(BACKEND, 'dto')
for root, dirs, files in os.walk(dto_dir):
    for f in files:
        if not f.endswith('.go'):
            continue
        s = open(os.path.join(root, f), encoding='utf-8').read()
        for m in re.finditer(r'type\s+(\w+)\s+struct\s*\{(.*?)\n\}', s, re.S):
            name, body = m.group(1), m.group(2)
            fields = {}
            for line in body.split('\n'):
                fm = re.search(r'json:"([^",]+)', line)
                if not fm:
                    continue
                fields[fm.group(1)] = bool(re.search(r'binding:"[^"]*required', line))
            if fields:
                dto_fields[name] = fields

# 2. 路由 -> handler（handler 是引用形式 h.ListRules)）
routes = []
api_dir = os.path.join(BACKEND, 'api')
for root, dirs, files in os.walk(api_dir):
    for f in files:
        if not f.endswith('.go'):
            continue
        s = open(os.path.join(root, f), encoding='utf-8').read()
        for line in s.split('\n'):
            lm = re.match(r'\s*\w+\.(GET|POST|PUT|DELETE|PATCH)\(\s*"([^"]+)"', line)
            if not lm:
                continue
            hm = re.search(r'\.(\w+)\)\s*$', line.strip())
            if hm:
                routes.append((lm.group(1), lm.group(2), hm.group(1)))

# 3. handler -> dto 类型（var req dto.Xxx）
handler_dto = {}
for root, dirs, files in os.walk(api_dir):
    for f in files:
        if not f.endswith('.go'):
            continue
        s = open(os.path.join(root, f), encoding='utf-8').read()
        for m in re.finditer(r'func\s+\(h \*\w+\)\s+(\w+)\(([^)]*)\)\s*\{(.*?)\n\}', s, re.S):
            hname, params, body = m.group(1), m.group(2), m.group(3)
            pm = re.search(r'(\w+)\s+(?:dto\.)?(\w+)\s*$', params.strip())
            if pm and pm.group(2) in dto_fields:
                handler_dto[hname] = pm.group(2)
                continue
            for bm in re.finditer(r'var\s+(\w+)\s+(?:dto\.)?(\w+)', body):
                if bm.group(2) in dto_fields:
                    handler_dto[hname] = bm.group(2)
                    break

# 4. 前端 url -> data 字面量字段
fe = {}
for root, dirs, files in os.walk(FRONTEND):
    dirs[:] = [d for d in dirs if d != 'node_modules']
    for f in files:
        if not f.endswith('.ts'):
            continue
        s = open(os.path.join(root, f), encoding='utf-8').read()
        url_re = re.compile(r'url:\s*[' + re.escape(QUOTE) + r']([^' + re.escape(QUOTE) + r']+)[ ' + re.escape(QUOTE) + r']')
        for m in url_re.finditer(s):
            url = re.sub(r'\$\{[^}]*\}', '{x}', m.group(1))
            seg = s[m.start():m.start() + 600]
            dm = re.search(r'data:\s*\{([^{}]*)\}', seg)
            keys = re.findall(r'(\w+)\s*:', dm.group(1)) if dm else None
            fe.setdefault(url, []).append((f, keys))

def camel(s2):
    parts = s2.split('_')
    return parts[0] + ''.join(p.title() for p in parts[1:])

print(f"后端 dto: {len(dto_fields)} | 路由: {len(routes)} | handler->dto: {len(handler_dto)} | 前端 url: {len(fe)}")
print("=== 后端 required 字段 vs 前端载荷对账 ===")
issues = checked = 0
for method, path, handler in routes:
    dto_name = handler_dto.get(handler)
    if not dto_name:
        continue
    required = [k for k, req in dto_fields[dto_name].items() if req]
    if not required:
        continue
    fe_url = re.sub(r'/:(\w+)', '{x}', path).replace('/api/v1', '')
    cands = []
    for url, entries in fe.items():
        if re.sub(r'/\{x\}', '{x}', url).replace('/api/v1', '') == fe_url:
            cands.extend(entries)
    if not cands:
        continue
    checked += 1
    for (f, keys) in cands:
        if keys is None:
            print(f"  [?] {method} {path} ({dto_name}) <- {f}: data 为变量，需人工核对 required={required}")
            issues += 1
            continue
        norm = set()
        for k in keys:
            norm |= {k, camel(k), k.lower()}
        missing = [r for r in required if r not in norm and camel(r) not in norm]
        if missing:
            print(f"  [X] {method} {path} ({dto_name}) <- {f}: 缺 required 字段 {missing}（现有: {keys}）")
            issues += 1
print(f"\n检查 {checked} 个接口，疑似字段不匹配: {issues} 个")
