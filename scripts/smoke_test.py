#!/usr/bin/env python3
"""Exercise the packaged scheduler + node using isolated SQLite data and loopback ports."""
import json, os, pathlib, socket, sqlite3, subprocess, tempfile, time, urllib.parse, urllib.request
ROOT = pathlib.Path(__file__).resolve().parents[1]

def port():
    with socket.socket() as s:
        s.bind(('127.0.0.1',0)); return s.getsockname()[1]

with tempfile.TemporaryDirectory(prefix='gocron-smoke-') as directory:
    webport, nodeport = port(), port()
    base = f'http://127.0.0.1:{webport}'
    env = dict(os.environ, GOCRON_DATA_DIR=directory)
    token = ''
    def api(path, data=None, authenticated=True):
        body = None if data is None else urllib.parse.urlencode(data).encode()
        req = urllib.request.Request(base+'/api'+path, data=body)
        if authenticated and token: req.add_header('Auth-Token',token)
        with urllib.request.urlopen(req,timeout=10) as response: result=json.load(response)
        assert result['code']==0, (path,result)
        return result['data']
    def start_web():
        p=subprocess.Popen([str(ROOT/'bin/gocron'),'web','--host','127.0.0.1','-p',str(webport)],cwd=directory,env=env,stdout=output,stderr=output)
        for _ in range(100):
            if p.poll() is not None: raise RuntimeError('scheduler failed to start')
            try: api('/install/status',authenticated=False); return p
            except OSError: time.sleep(.1)
        p.terminate(); p.wait(10); raise RuntimeError('startup timed out')
    def stop(p):
        if p.poll() is None: p.terminate(); p.wait(timeout=15)
    with open(pathlib.Path(directory)/'smoke.log','w+') as output:
        node=subprocess.Popen([str(ROOT/'bin/gocron-node'),'-s',f'127.0.0.1:{nodeport}'],cwd=directory,stdout=output,stderr=output)
        web=None
        try:
            for _ in range(100):
                if node.poll() is not None: raise RuntimeError('node failed to start')
                try:
                    with socket.create_connection(('127.0.0.1', nodeport), timeout=.1): break
                except OSError: time.sleep(.1)
            web=start_web()
            assert api('/install/status',authenticated=False) is False
            api('/install/store',dict(db_type='sqlite3',admin_username='tester',admin_password='test-password-123',confirm_admin_password='test-password-123',admin_email='tester@example.test'),False)
            token=api('/user/login',dict(username='tester',password='test-password-123'),False)['token']
            api('/host/store',dict(name='127.0.0.1',alias='Local smoke node',port=nodeport))
            hosts=api('/host')['data']; hid=hosts[0]['id']
            api(f'/host/ping/{hid}')
            common=dict(level=1,spec='0 0 0 1 1 *',http_method=1,timeout=5,multi=2,retry_times=0,retry_interval=0,notify_status=1,notify_type=1,dependency_status=1)
            api('/task/store',dict(common,name='Shell smoke',protocol=2,command='printf sqlite-shell-ok',host_id=hid))
            api('/task/store',dict(common,name='HTTP smoke',protocol=1,command=base+'/api/install/status'))
            tasks=api('/task')['data']; assert len(tasks)==2
            for task in tasks: api('/task/run/'+str(task['id']))
            for _ in range(100):
                logs=api('/task/log')['data']
                if len(logs)>=2 and all(l['status']==2 for l in logs): break
                time.sleep(.1)
            else: raise AssertionError(('tasks failed',logs))
            assert any('sqlite-shell-ok' in log['result'] for log in logs)
            api('/system/webhook/update',dict(url='http://127.0.0.1:1',template='test-template'))
            assert api('/system/webhook')['template']=='test-template'
            # Disabled users must not be able to log in; read-only users cannot mutate tasks.
            api('/user/store',dict(name='reader',password='reader-pass',confirm_password='reader-pass',email='reader@example.test',is_admin=0,status=1))
            for task in tasks: api('/task/disable/'+str(task['id']),{})
            stop(web); web=start_web()
            assert api('/install/status') is True
            assert len(api('/task')['data'])==2
            assert len(api('/task/log')['data'])==2
            assert api('/system/webhook')['template']=='test-template'
            db=sqlite3.connect(pathlib.Path(directory)/'data/gocron.db')
            assert db.execute('PRAGMA integrity_check').fetchone()[0]=='ok'
            assert db.execute('PRAGMA journal_mode').fetchone()[0]=='wal'
            db.close()
            print('PASS: fresh SQLite install, login, host RPC, HTTP/Shell execution, logs, settings, users, disable, restart persistence, SQLite integrity/WAL')
        except Exception:
            output.flush(); output.seek(0); print(output.read()[-10000:]); raise
        finally:
            if web: stop(web)
            stop(node)
