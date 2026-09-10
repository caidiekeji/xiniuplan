# -*- coding: utf-8 -*-
"""TS 接口字段 vs 后端 dto required 字段对比（处理 data 为变量的情况）"""
import re, os, io, sys
sys.stdout = io.TextIOWrapper(sys.stdout.buffer, encoding='utf-8')

BACKEND = os.path.join(os.path.dirname(__file__), '..', 'backend', 'internal')
FRONTEND = os.path.join(os.path.dirname(__file__), '..', 'frontend', 'src')
QUOTE = "`'\""

# 后端 dto
dto_fields = {}
for root, dirs, files in os.walk(os.path.join(BACKEND, 'dto')):
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

# 路由 -> handler -> dto
routes = []
handler_dto = {}
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

# 前端：TS 接口字段（每个 api 文件独立命名空间）
def parse_ts_interfaces(s):
    ifaces = {}
    for m in re.finditer(r'export\s+interface\s+(\w+)\s*\{(.*?)\n\}', s, re.S):
        name, body = m.group(1), m.group(2)
        fields = set()
        for line in body.split('\n'):
            fm = re.match(r'\s*(\w+)\??\s*:', line)
            if fm:
                fields.add(fm.group(1))
        ifaces[name] = fields
    return ifaces

# 前端 url + data 变量类型
fe = {}
for root, dirs, files in os.walk(FRONTEND):
    dirs[:] = [d for d in dirs if d != 'node_modules']
    for f in files:
        if not f.endswith('.ts'):
            continue
        s = open(os.path.join(root, f), encoding='utf-8').read()
        ifaces = parse_ts_interfaces(s)
        url_re = re.compile(r'url:\s*[' + re.escape(QUOTE) + r']([^' + re.escape(QUOTE) + r']+)[ ' + re.escape(QUOTE) + r']')
        # 函数签名: export function name(arg: Type) { ... url: ... data: arg }
        for m in url_re.finditer(s):
            url = re.sub(r'\$\{[^}]*\}', '{x}', m.group(1))
            # 向前找最近的 function 声明
            head = s[:m.start()]
            fn = None
            for fm in re.finditer(r'export\s+function\s+(\w+)\(([^)]*)\)', head):
                if fm.end() > m.start() - 800:
                    fn = fm
            if not fn:
                fe.setdefault(url, []).append((f, None, None))
                continue
            # data: 引用的变量
            seg = s[m.start():m.start() + 600]
            dm = re.search(r'data:\s*(\w+)', seg)
            var_name = dm.group(1) if dm else None
            # 找该变量在函数参数中的类型
            var_type = None
            if var_name and fn:
                for pm in re.finditer(r'(\w+)(\??)\s*:\s*([\w<>\[\]| ]+?)(?:,|\))', fn.group(2)):
                    if pm.group(1) == var_name:
                        var_type = pm.group(3).strip()
                        break
            # 对象字面量（简写/显式都算）
            literal_fields = set()
            seg2 = s[m.start():m.start() + 600]
            lm2 = re.search(r'data:\s*\{([^{}]*)\}', seg2)
            if lm2:
                for tok in re.findall(r'(\w+)\s*(?::|,|\})', lm2.group(1)):
                    literal_fields.add(tok)
            fe.setdefault(url, []).append((f, var_type, ifaces.get(var_type), literal_fields))

def norm_set(fields):
    out = set()
    def camel(s2):
        parts = s2.split('_')
        return parts[0] + ''.join(p.title() for p in parts[1:])
    for k in fields:
        out |= {k, camel(k), k.lower()}
    return out

print("=== [?] 项二次核验：TS 接口 vs 后端 dto required ===")
issues = 0
seen = set()
for method, path, handler in routes:
    dto_name = handler_dto.get(handler)
    if not dto_name:
        continue
    required = [k for k, req in dto_fields[dto_name].items() if req]
    if not required:
        continue
    key = (method, path)
    if key in seen:
        continue
    fe_url = re.sub(r'/:(\w+)', '{x}', path).replace('/api/v1', '')
    for url, entries in fe.items():
        u = re.sub(r'/\{x\}', '{x}', url).replace('/api/v1', '')
        if u != fe_url:
            continue
        for (f, var_type, iface_fields, literal_fields) in entries:
            if literal_fields:
                continue  # 字面量已核
            if var_type and iface_fields is not None:
                ns = norm_set(iface_fields)
                missing = [r for r in required if r not in ns and camel(r) not in ns]
                tag = f"{method} {path} ({dto_name}) <- {f}:{var_type}"
                if missing:
                    print(f"  [X] {tag}: 接口缺 required 字段 {missing}；接口字段: {sorted(iface_fields)}")
                    issues += 1
                seen.add(key)
                break
            elif var_type and iface_fields is None:
                print(f"  [?] {method} {path} ({dto_name}) <- {f}:{var_type}: 未找到接口定义，需人工核对 required={required}")
                issues += 1
                seen.add(key)
                break
print(f"\n疑似问题: {issues} 个")
