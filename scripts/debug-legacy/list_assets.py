import paramiko

HOST, USER, PASS = '192.168.31.154', 'jie', '12311'

ssh = paramiko.SSHClient()
ssh.set_missing_host_key_policy(paramiko.AutoAddPolicy())
ssh.connect(HOST, username=USER, password=PASS, timeout=10)

def run(cmd):
    _, stdout, _ = ssh.exec_command(cmd)
    return stdout.read().decode().strip()

print('=== assets 目录 ===')
print(run('ls -la /var/www/rhinopanel/assets/ | head -30'))

print('\n=== index-*.js 文件 ===')
print(run('ls /var/www/rhinopanel/assets/index-*.js'))

ssh.close()
