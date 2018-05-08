from __future__ import (absolute_import, division, print_function)
__metaclass__ = type

from ansible.plugins.connection.ssh import Connection as SSHConnection
from ansible import constants as C

class Connection(SSHConnection):
    ''' ssh based connections for powershell via packer'''

    transport = 'packer'
    module_implementation_preferences = ('.ps1', '.exe', '')
    become_methods = ['runas']
    allow_executable = False

    def __init__(self, *args, **kwargs):
        self._shell_type = 'powershell'

        super(Connection, self).__init__(*args, **kwargs)

        self.host = self._play_context.remote_addr
        self.port = self._play_context.port
        self.user = self._play_context.remote_user
        self.control_path = C.ANSIBLE_SSH_CONTROL_PATH
        self.control_path_dir = C.ANSIBLE_SSH_CONTROL_PATH_DIR
