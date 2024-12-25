import mistune

from typing import List, Tuple

import subprocess, os, sys, shutil, io

import json

_REPOSITORY_ROOT_PATH = os.path.abspath(os.path.join(os.path.dirname(os.path.abspath(__file__)), '..', '..', '..'))

"""
Intentions:

  - Support markdown parsing.
  - Support sequential markdown code block execution, leveraging [info strings](https://spec.commonmark.org/0.30/#info-string).
"""

class ASTNode(object):

    _STACKQL_SHELL_INVOCATION: str = 'stackql-shell'
    _BASH: str = 'bash'
    _SETUP: str = 'setup'
    _TEARDOWN: str = 'teardown'

    def __init__(self, node: dict):
        self.node = node
        self.children = []
        if 'children' in node:
            for child in node['children']:
                self.children.append(ASTNode(child))

    def get_type(self) -> str:
        return self.node.get('type', '')

    def get_text(self) -> str:
        return self.node.get('raw', '').strip()

    def is_executable(self) -> bool:
        return self.get_type() == 'block_code'
    
    def _get_annotations(self) -> List[str]:
        return self.node.get('attrs').get('info', '').split(' ')
    
    def is_stackql_shell_invocation(self) -> bool:
        return self._STACKQL_SHELL_INVOCATION in self._get_annotations()
    
    def is_bash(self) -> bool:
        return self._BASH in self._get_annotations()
    
    def is_setup(self) -> bool:
        return self._SETUP in self._get_annotations()
    
    def is_teardown(self) -> bool:
        return self._TEARDOWN in self._get_annotations()

    def get_execution_language(self) -> str:
        return self.node.get('lang', '')

    def __str__(self):
        return json.dumps(self.node, indent=2)
    
    def __repr__(self):
        return self.__str__()

class MdParser(object):

    def parse_markdown_file(self, file_path: str, lang=None) -> List[ASTNode]:
        markdown: mistune.Markdown = mistune.create_markdown(renderer='ast')
        with open(file_path, 'r') as f:
            txt = f.read()
        raw_list: List[dict] = markdown(txt)
        return [ASTNode(node) for node in raw_list]


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
    
    def __str__(self):
        return f'Setup: {self._setup}\nIn Session: {self._in_session}\nTeardown: {self._teardown}'
    
    def __repr__(self):
        return self.__str__()

class MdOrchestrator(object):

    def __init__(
        self,
        parser: MdParser, 
        max_setup_blocks: int = 1,
        max_invocations_blocks: int = 1,
        max_teardown_blocks: int = 1,
        setup_contains_shell_invocation: bool = True
    ):
        self._parser = parser
        self._max_setup_blocks = max_setup_blocks
        self._max_invocations_blocks = max_invocations_blocks
        self._max_teardown_blocks = max_teardown_blocks
        self._setup_contains_shell_invocation = setup_contains_shell_invocation

    def orchestrate(self, file_path: str) -> WorkloadDTO:
        setup_count: int = 0
        teardown_count: int = 0
        invocation_count: int = 0
        ast = self._parser.parse_markdown_file(file_path)
        print(f'AST: {ast}')
        setup_str: str = f'cd {_REPOSITORY_ROOT_PATH};\n'
        in_session_commands: List[str] = []
        teardown_str: str = f'cd {_REPOSITORY_ROOT_PATH};\n'
        for node in ast:
            if node.is_executable():
                if node.is_setup():
                    if setup_count < self._max_setup_blocks:
                        setup_str += f'{node.get_text()}'
                        setup_count += 1
                    else:
                        raise KeyError(f'Maximum setup blocks exceeded: {self._max_setup_blocks}')
                elif node.is_teardown():
                    if teardown_count < self._max_teardown_blocks:
                        teardown_str += f'{node.get_text()}'
                        teardown_count += 1
                    else:
                        raise KeyError(f'Maximum teardown blocks exceeded: {self._max_teardown_blocks}')
                elif node.is_stackql_shell_invocation():
                    if invocation_count < self._max_invocations_blocks:
                        all_commands: str = node.get_text().split('\n\n')
                        in_session_commands += all_commands
                        invocation_count += 1
                    else:
                        raise KeyError(f'Maximum invocation blocks exceeded: {self._max_invocations_blocks}')
        return WorkloadDTO(setup_str, in_session_commands, teardown_str)

class SimpleE2E(object):

    def __init__(self, workload: WorkloadDTO):
        self._workload = workload

    def run(self) -> Tuple[bytes, bytes]:
        bash_path = shutil.which('bash')
        pr: subprocess.Popen = subprocess.Popen(
            self._workload.get_setup(),
            stdin=subprocess.PIPE,
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
            shell=True,
            executable=bash_path
        )
        for cmd in self._workload.get_in_session():
            pr.stdin.write(f"{cmd}\n".encode(sys.getdefaultencoding()))
            pr.stdin.flush()
        stdoout_bytes, stderr_bytes = pr.communicate()
        return (stdoout_bytes, stderr_bytes,)

if __name__ == '__main__':
    md_parser = MdParser()
    orchestrator: MdOrchestrator = MdOrchestrator(md_parser)
    workload_dto: WorkloadDTO = orchestrator.orchestrate(os.path.join(_REPOSITORY_ROOT_PATH, 'docs', 'walkthroughs', 'get-google-vms.md'))
    print(f'Workload DTO: {workload_dto}')
    # print(json.dumps(parsed_file, indent=2))
    e2e: SimpleE2E = SimpleE2E(workload_dto)
    stdout_bytes, stderr_bytes = e2e.run()
    print(stdout_bytes.decode(sys.getdefaultencoding()))
    print(stderr_bytes.decode(sys.getdefaultencoding()))

    

