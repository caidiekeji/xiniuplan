import paramiko

HOST, USER, PASS = '192.168.31.154', 'jie', '12311'

ssh = paramiko.SSHClient()
ssh.set_missing_host_key_policy(paramiko.AutoAddPolicy())
ssh.connect(HOST, username=USER, password=PASS, timeout=10)
sftp = ssh.open_sftp()

# 下载 chunk 文件
sftp.get('/var/www/rhinopanel/assets/index-CEoSX4-N.js', r'E:\www\fuwuqi\index-CEoSX4-N.js')
print('下载完成')

# 读取文件内容，搜索 .map( 调用
with open(r'E:\www\fuwuqi\index-CEoSX4-N.js', 'r', encoding='utf-8') as f:
    content = f.read()

print(f'文件大小: {len(content)} 字符')

# 搜索 .map( 调用
import re
matches = [(m.start(), content[max(0,m.start()-100):m.start()+50]) for m in re.finditer(r'\.map\(', content)]
print(f'\n找到 {len(matches)} 个 .map() 调用:')
for i, (pos, ctx) in enumerate(matches[:10]):
    print(f'\n--- 调用 {i+1} (位置 {pos}) ---')
    print(ctx)

sftp.close()
ssh.close()
