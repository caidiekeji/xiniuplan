#!/usr/bin/env python3
"""i18n key 完整性校验：扫描 src 中所有静态 t('a.b.c') 引用，
逐个核对 zh-CN 消息树中存在（zh-CN 是 fallbackLocale，缺 key 会直接把 key 原文显示给用户）。

用法：python scripts/check-i18n.py  （退出码 0=全绿，1=有缺失）
"""
import os
import re
import sys

ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
SRC = os.path.join(ROOT, "frontend", "src")
ZH = os.path.join(SRC, "locales", "zh-CN")

T_KEY = re.compile(r"""\bt\(\s*['"]([A-Za-z0-9_.]+)['"]""")
IMPORT_RE = re.compile(r"""import\s+(\w+)\s+from\s+['"]\./([\w/]+)['"]""")
# 消息值字符串字面量——用于裸 @ 扫描
MSG_VALUE_RE = re.compile(r"""['"]([^'"\n]*)['"]""")
ESCAPED_AT_RE = re.compile(r"""\{'@'\}|\{@\}""")


def scan_bare_at():
    """扫描三个语言包消息值里未转义的裸 @。

    intlify 把 @ 视为 linked message 语法起点（@:key / @.lower:key），
    裸 @（如邮箱 noreply@example.com）会在运行时编译报 Invalid linked format，白屏。
    转义写法：{'@'}。
    """
    hits = []
    locales_dir = os.path.join(SRC, "locales")
    for lang in sorted(os.listdir(locales_dir)):
        lang_dir = os.path.join(locales_dir, lang)
        if not os.path.isdir(lang_dir):
            continue
        for dirpath, _dirnames, filenames in os.walk(lang_dir):
            for fn in filenames:
                if not fn.endswith(".ts"):
                    continue
                path = os.path.join(dirpath, fn)
                try:
                    lines = open(path, encoding="utf-8").read().splitlines()
                except UnicodeDecodeError:
                    continue
                for i, line in enumerate(lines, 1):
                    stripped = line.strip()
                    if stripped.startswith(("//", "*", "import")):
                        continue
                    if "@" not in stripped:
                        continue
                    # 已全部转义则跳过：剥掉 {'@'} 后不再有裸 @
                    residue = ESCAPED_AT_RE.sub("", stripped)
                    if "@" not in residue:
                        continue
                    hits.append(f"{os.path.relpath(path, SRC)}:{i}: {stripped[:90]}")
    return hits


def collect_keys():
    keys = set()
    for dirpath, dirnames, filenames in os.walk(SRC):
        dirnames[:] = [d for d in dirnames if d not in ("node_modules", "dist")]
        for fn in filenames:
            if not fn.endswith((".vue", ".ts")):
                continue
            path = os.path.join(dirpath, fn)
            try:
                text = open(path, encoding="utf-8").read()
            except UnicodeDecodeError:
                continue
            for m in T_KEY.finditer(text):
                keys.add(m.group(1))
    return keys


def load_zh_tree():
    """解析 locales/zh-CN/index.ts 的模块注册，返回 {top: 模块文件路径或 None}"""
    idx = open(os.path.join(ZH, "index.ts"), encoding="utf-8").read()
    top = {}
    for mod, rel in IMPORT_RE.findall(idx):
        p = os.path.join(ZH, rel + ".ts")
        top[mod] = p if os.path.exists(p) else None
    return top


def key_exists(mod_path, dotted):
    """在模块 ts 文件里做宽松属性链检查：每个段必须以 `seg:` / `'seg':` / `seg :` 形式出现"""
    try:
        text = open(mod_path, encoding="utf-8").read()
    except OSError:
        return False
    for seg in dotted.split("."):
        if not re.search(r"""['"]?%s['"]?\s*:""" % re.escape(seg), text):
            return False
        # 收窄到该段之后的嵌套体（近似：从首个匹配处向后查下一个段）
        m = re.search(r"""['"]?%s['"]?\s*:""" % re.escape(seg), text)
        text = text[m.end():]
    return True


def main():
    keys = collect_keys()
    top = load_zh_tree()
    missing = []
    unknown_top = []
    for k in sorted(keys):
        head, _, rest = k.partition(".")
        if head not in top:
            # 可能是局部函数返回的动态 key 拼接，标记为人工确认
            unknown_top.append(k)
            continue
        path = top[head]
        if path is None:
            missing.append(k)
            continue
        if rest and not key_exists(path, rest):
            missing.append(k)

    bare_at = scan_bare_at()

    print(f"引用 key 总数: {len(keys)}")
    if unknown_top:
        print(f"\n[人工确认] 顶层模块不存在（可能是动态 key）: {len(unknown_top)}")
        for k in unknown_top:
            print("  ?", k)
    if bare_at:
        print(f"\n[裸 @] locale 消息值含未转义 @（intlify 把 @ 当 linked 语法，运行时报 Invalid linked format）: {len(bare_at)}")
        for loc in bare_at:
            print("  x", loc)
    if missing:
        print(f"\n[缺失] zh-CN 中不存在: {len(missing)}")
        for k in missing:
            print("  x", k)
        return 1
    if bare_at:
        return 1
    print("✅ 全部 key 在 zh-CN 中存在")
    return 0


if __name__ == "__main__":
    sys.exit(main())
