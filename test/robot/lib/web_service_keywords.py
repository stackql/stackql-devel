from robot.api.deco import library, keyword

from robot.libraries.Process import Process

import json

from requests import get, post, Response

import os

from typing import Union, Tuple, List

@library
class web_service_keywords(Process):

    _DEFAULT_SQLITE_DB_PATH: str = os.path.abspath(os.path.join(os.path.dirname(__file__), "..", "..", "tmp", "robot_cli_affirmation_store.db"))

    def _get_dsn(self) -> str:
        return self._DEFAULT_SQLITE_DB_PATH
    
    def __init__(self):
        self._affirmation_store_web_service = None
        self._web_server_app: str = 'test/python/flask/oauth2/token_srv'
        super().__init__()

    @keyword
    def create_oauth2_client_credentials_web_service(
        self,
        port: int
    ) -> None:
        """
        Sign the input.
        """
        return self.start_process(
            'flask',
            f'--app={self._web_server_app}',
            'run',
            f'--port={port}',
            stdout=os.path.abspath(os.path.join(os.path.dirname(__file__), '..', 'log', f'token-client-credentials-{port}-stdout.txt')),
            stderr=os.path.abspath(os.path.join(os.path.dirname(__file__), '..', 'log', f'token-client-credentials-{port}-stderr.txt'))
        )
    
    @keyword
    def send_get_request(
        self,
        address: str
    ) -> Response:
        """
        Send a simple get request.
        """
        return get(address)
    
    @keyword
    def send_json_post_request(
        self,
        address: str,
        input: dict
    ) -> Response:
        """
        Send a canonical json post request.
        """
        return post(address, json=input)
    
    @keyword
    def send_to_affirmation_store(
        self,
        frame_key_val: str,
        json_input: Union[str, dict],
        affirmation_store_url: str = 'http://127.0.0.1:5848',
        path_prefix: str = '/data/',
        frame_key: str = '_frame_hash'
    ) -> Response:
        """
        Send a canonical json post request.
        """
        if isinstance(json_input, str):
            input = json.loads(json_input)
        return post(f'{affirmation_store_url}{path_prefix}{frame_key_val}', json=input)
    
    @keyword
    def extract_frame_key(
        self,
        json_input: Union[str, dict],
        frame_key: str = '_frame_hash'
    ) -> Response:
        """
        Extract frame hash key.
        """
        if isinstance(json_input, str):
            json_input = json.loads(json_input)
        frame_key_val = json_input.get(frame_key)
        return frame_key_val
    
    @keyword
    def retrieve_from_affirmation_store(
        self,
        frame_key_val: str,
        affirmation_store_url: str = 'http://127.0.0.1:5848',
        path_prefix: str = '/data/',
        frame_key: str = '_frame_hash'
    ) -> Response:
        """
        Retrieve from affirmation store.
        """
        return get(f'{affirmation_store_url}{path_prefix}{frame_key_val}')