import subprocess

import ols.base
import ols.render
import ols.postgres
import ols.http

from charms.reactive import (
    when,
    when_not,
    set_state,
    remove_state,
)

from charmhelpers.core import hookenv

@when('ols.configured', 'db.master.available')
def configure(pgsql):
    if hookenv.hook_name() == 'update-status':
        return

    config = hookenv.config().copy()
    maindb, standbys = ols.postgres.get_db_uris(pgsql)
    config['maindb'] = maindb
    if standbys:
        config['standbys_urls'] = standbys

    env = ols.base.env_vars()
    if env['SERVICE_ENVIRONMENT'] == 'staging':
        config['ssoBaseURL'] = 'https://login.staging.ubuntu.com/api/v2/'
        config['storeBaseURL'] = 'https://dashboard.staging.snapcraft.io/dev/api'

    name = ols.base.service_name()
    working_dir = ols.base.code_dir()
    config_path = ols.base.service_config_path()

    changed = ols.render.setup_charm_service(
        exec_path='{}/serial-vault'.format(working_dir),
        exec_args='--config {}'.format(config_path),
    )

    changed |= ols.render.render(
        template='settings.yaml',
        target=config_path,
        context=config,
    )

    if not changed:
        hookenv.log('nothing changed, doing nothing')
        # other layers set status to waiting, so we need to unset if we're good
        hookenv.status_set('active', 'ready')
        return

    restart_cmd = ['systemctl', 'restart', name]
    try:
        subprocess.check_output(restart_cmd)
    except subprocess.CalledProcessError as exc:
        if exc.output:
            hookenv.log(exc.output)
        hookenv.status_set('blocked', 'error restarting service')
        raise
    else:
        set_state('service.configured')
        hookenv.status_set('active', 'ready')
