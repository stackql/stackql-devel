from robot.api.deco import keyword, library

from robot.libraries.BuiltIn import BuiltIn

from robot.libraries.Process import Process

from robot.libraries.OperatingSystem import OperatingSystem 

import os



@library(scope='SUITE', version='0.1.0', doc_format='reST')
class StackQLInterfaces(OperatingSystem, Process, BuiltIn):
  ROBOT_LISTENER_API_VERSION = 2

  def __init__(self):
    self._counter = 0
    self.ROBOT_LIBRARY_LISTENER = self
    Process.__init__(self)

  def _end_suite(self, name, attrs):
    print('Suite %s (%s) ending.' % (name, attrs['id']))

  def count(self):
    self._counter += 1
    print(self._counter)

  def clear_counter(self):
    self._counter = 0


  def _run_PG_client_command(self, curdir :str, psql_exe :str, psql_conn_str :str, query :str):
    _mod_conn =  psql_conn_str.replace("\\", "/")
    # bi = BuiltIn().get_library_instance('Builtin')
    self.log_to_console(f"curdir = '{curdir}'")
    self.log_to_console(f"psql_exe = '{psql_exe}'")
    result = super().run_process(
      psql_exe, 
      '-d', _mod_conn, 
      '-c', query
    )
    self.log(result.stdout)
    self.log(result.stderr)
    return result

  def _run_stackql_exec_command(
    self,  
    stackql_exe :str, 
    okta_secret_str :str,
    github_secret_str :str,
    k8s_secret_str :str,
    registry_cfg_str :str, 
    auth_cfg_str :str, 
    query,
    **cfg
  ):
    self.set_environment_variable("OKTA_SECRET_KEY", okta_secret_str)
    self.set_environment_variable("GITHUB_SECRET_KEY", github_secret_str)
    self.set_environment_variable("K8S_SECRET_KEY", k8s_secret_str)
    res = super().run_process(
      stackql_exe,
      "exec",
      f"--registry={registry_cfg_str}",
      f"--auth={auth_cfg_str}",
      "--tls.allowInsecure=true",
      query,
      **cfg
    )
    self.log(res.stdout)
    self.log(res.stderr)
    return res
  
  @keyword
  def should_PG_client_error_inline_contain(self, curdir :str, psql_exe :str, psql_conn_str :str, query :str, expected_output :str):
    result =    self._run_PG_client_command(
      curdir,
      psql_exe,
      psql_conn_str,
      query
    ) #    ${CURDIR}    ${PSQL_EXE}    ${_CONN_STR}    ${_QUERY}
    return self.should_contain(result.stderr, expected_output)
  
  @keyword
  def should_PG_client_inline_contain(self, curdir :str, psql_exe :str, psql_conn_str :str, query :str, expected_output :str):
    result =    self._run_PG_client_command(
      curdir,
      psql_exe,
      psql_conn_str,
      query
    )
    return self.should_contain(result.stdout, expected_output)
  
  @keyword
  def should_PG_client_inline_equal(self, curdir :str, psql_exe :str, psql_conn_str :str, query :str, expected_output :str):
    result =    self._run_PG_client_command(
      curdir,
      psql_exe,
      psql_conn_str,
      query
    )
    return self.should_be_equal(result.stdout, expected_output)
  
  @keyword
  def should_stackql_inline_equal(
    self, 
    stackql_exe :str, 
    okta_secret_str :str,
    github_secret_str :str,
    k8s_secret_str :str,
    registry_cfg_str :str, 
    auth_cfg_str :str, 
    query :str,
    expected_output :str
  ):
    result = self._run_stackql_exec_command(
      stackql_exe, 
      okta_secret_str,
      github_secret_str,
      k8s_secret_str,
      registry_cfg_str, 
      auth_cfg_str, 
      query,
    )
    return self.should_be_equal(result.stdout, expected_output)

  @keyword
  def should_horrid_query_stackql_inline_equal(
    self, 
    stackql_exe :str, 
    okta_secret_str :str,
    github_secret_str :str,
    k8s_secret_str :str,
    registry_cfg_str :str, 
    auth_cfg_str :str,
    expected_output :str,
    stdout_tmp_file :str,
  ):
    _SELECT_GOOGLE_JOIN_CONCATENATED_SELECT_EXPRESSIONS :bytes =  b"""SELECT i.zone, i.name, i.machineType, i.deletionProtection, '[{"subnetwork":"' || JSON_EXTRACT(i.networkInterfaces, '$[0].subnetwork') || '"}]', '[{"boot": true, "initializeParams": { "diskSizeGb": "' || JSON_EXTRACT(i.disks, '$[0].diskSizeGb') || '", "sourceImage": "' || d.sourceImage || '"}}]', i.labels FROM google.compute.instances i INNER JOIN google.compute.disks d ON i.name = d.name WHERE i.project = 'testing-project' AND i.zone = 'australia-southeast1-a' AND d.project = 'testing-project' AND d.zone = 'australia-southeast1-a' AND i.name LIKE '%' order by i.name DESC;"""
    if os.name == 'nt':
      _SELECT_GOOGLE_JOIN_CONCATENATED_SELECT_EXPRESSIONS =  _SELECT_GOOGLE_JOIN_CONCATENATED_SELECT_EXPRESSIONS
    result = self._run_stackql_exec_command(
      stackql_exe, 
      okta_secret_str,
      github_secret_str,
      k8s_secret_str,
      registry_cfg_str, 
      auth_cfg_str, 
      _SELECT_GOOGLE_JOIN_CONCATENATED_SELECT_EXPRESSIONS,
      **{"stdout": stdout_tmp_file }
    )
    return self.should_be_equal(result.stdout, expected_output)