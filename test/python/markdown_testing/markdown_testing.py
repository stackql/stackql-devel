import mistune

from typing import List, Tuple

import subprocess, os, sys

class ASTNode(object):

    def __init__(self, node: dict):
        self.node = node
        self.children = []
        if 'children' in node:
            for child in node['children']:
                self.children.append(ASTNode(child))

    def get_type(self) -> str:
        return self.node.get('type', '')

    def get_text(self) -> str:
        return self.node.get('text', '')

    def is_executable(self) -> bool:
        return self.get_type() == 'code_block'

    def get_execution_language(self) -> str:
        return self.node.get('lang', '')


class MdParser(object):

    def parse_markdown_file(self, file_path: str, lang=None) -> List[ASTNode]:
        markdown: mistune.Markdown = mistune.create_markdown(renderer='ast')
        with open(file_path, 'r') as f:
            txt = f.read()
        return markdown(txt)


class MdExecutor(object):

    def __init__(self, renderer: MdParser):
        self.renderer = renderer

    def execute(self, file_path: str) -> None:
        ast = self.renderer.parse_markdown_file(file_path)
        for node in ast:
            if node.is_executable():
                lang = node.get_execution_language()
                if lang == 'python':
                    exec(node.get_text())
                else:
                    print(f'Unsupported language: {lang}')

class WorkloadDTO(object):

    def __init__(self, setup: str, in_session: List[str], teardown: str):
        self._setup = setup
        self._in_session = in_session
        self._teardown = teardown

    def get_setup(self) -> List[str]:
        return self._setup
    
    def get_in_session(self) -> List[str]:
        return self._in_session
    
    def get_teardown(self) -> List[str]:
        return self._teardown

class SimpleE2E(object):

    def __init__(self, workload: WorkloadDTO):
        self._workload = workload

    def run(self) -> Tuple[bytes, bytes]:
        pr: subprocess.Popen = subprocess.Popen(
            self._workload.get_setup(),
            stdin=subprocess.PIPE,
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
            shell=True,
            executable='/bin/bash'
        )
        for cmd in self._workload.get_in_session():
            pr.stdin.write(f"{cmd}\n".encode(sys.getdefaultencoding()))
            pr.stdin.flush()
        stdoout_bytes, stderr_bytes = pr.communicate()
        return (stdoout_bytes, stderr_bytes,)

if __name__ == '__main__':
    workload_dto: WorkloadDTO = WorkloadDTO(
        setup="""
            export GOOGLE_CREDENTIALS="$(cat /Users/admin/stackql/secrets/concerted-testing/google-credentials.json)" && 
            stackql shell""",
        in_session=[
            'registry pull google;',
            'select name, id FROM google.compute.instances WHERE project = \'ryuki-it-sandbox-01\' AND zone = \'australia-southeast1-a\';'
        ],
        teardown='echo "Goodbye, World!"'
    )
    e2e: SimpleE2E = SimpleE2E(workload_dto)
    stdout_bytes, stderr_bytes = e2e.run()
    print(stdout_bytes.decode(sys.getdefaultencoding()))
    print(stderr_bytes.decode(sys.getdefaultencoding()))

    

