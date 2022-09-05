import os
import subprocess
import sys


tmp_stdout_file_path = './file.tmp'

tmp_stdout_file = open(tmp_stdout_file_path, 'wb')

reg_cfg_docker = '{ "url": "file:///opt/stackql/registry", "localDocRoot": "/opt/stackql/registry", "verifyConfig": { "nopVerify": true } }'

auth_cfg_docker = '{}'

def start(cmd_arg_list):
  command = [item.encode(sys.getdefaultencoding()) for item in cmd_arg_list]
  return subprocess.Popen(
    command,
    stdin=subprocess.PIPE,
    stdout=tmp_stdout_file,
    stderr=subprocess.PIPE
  )

def start_docker_shell(cmd_arg_list):
  command = [item.encode(sys.getdefaultencoding()) for item in cmd_arg_list]
  return subprocess.Popen(
    command,
    stdin=subprocess.PIPE,
    stdout=subprocess.PIPE,
    stderr=subprocess.PIPE
  )


def read(process):
    return process.stdout.readline().decode("utf-8").strip()


def write(process, message):
    process.stdin.write(f"{message.strip()}\n".encode("utf-8"))
    process.stdin.flush()


def terminate(process):
    process.stdin.close()
    process.terminate()
    process.wait(timeout=0.2)


process = start(
  [ "./stackql",
    f"--registry={os.environ.get('REG_TEST')}",
    f"--auth={os.environ.get('AUTH_STR_INT')}",
    "shell"
  ]
)


docker_process = start_docker_shell(
  [ "docker-compose",
    "-p",
    "stackqlshell",
    "run",
    "--rm",
    "-e",
    f"OKTA_SECRET_KEY={os.environ.get('OKTA_SECRET_KEY')}", 
    "-e",
    f"GITHUB_SECRET_KEY={os.environ.get('GITHUB_CREDS')}",
    "-e",
    f"K8S_SECRET_KEY={os.environ.get('K8S_SECRET_KEY')}",
    "stackqlsrv",
    "bash",
    "-c",
    f"stackql --registry='{reg_cfg_docker}' --auth='{auth_cfg_docker}' shell",
  ]
)



write(process, "show providers;")

stdout_bytes, stderr_bytes = process.communicate()

tmp_stdout_file.close()

write(docker_process, "show providers;")

stdout_bytes_docker, stderr_bytes_docker = docker_process.communicate()

print()
print("### STDOUT ###")
print()

with open(tmp_stdout_file_path, 'rb') as f:
  for line in f.readlines():
    print(line.decode('utf-8').strip('\n'))

print()
print("### STDERR ###")
print()

print(stderr_bytes.decode('utf-8'))

print()

terminate(process)

print()
print("### DOCKER STDOUT ###")
print()

print(stdout_bytes_docker.decode('utf-8'))

print()
print("### DOCKER STDERR ###")
print()

print(stderr_bytes_docker.decode('utf-8'))

print()

terminate(process)