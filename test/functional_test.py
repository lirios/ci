import requests
import json
import unittest


class GoRunnerAPI(object):
    def __init__(self, host):
        self.host = host

    def list_jobs(self):
        r = requests.get("%s/jobs" % self.host)
        self._raise_if_status_not(r, 200)
        return r.json()

    def list_job_names(self):
        jobs = self.list_jobs()
        return [job['name'] for job in jobs]

    def add_job(self, name):
        r = requests.post("%s/jobs" % self.host, data=json.dumps({'name': name}))
        self._raise_if_status_not(r, 201)

    def get_job(self, name):
        r = requests.get("%s/jobs/%s" % (self.host, name))
        self._raise_if_status_not(r, 200)
        return r.json()

    def delete_job(self, name):
        r = requests.delete("%s/jobs/%s" % (self.host, name))
        self._raise_if_status_not(r, 200)

    def add_task_to_job(self, task, job):
        r = requests.post("%s/jobs/%s/tasks" % (self.host, job), json.dumps({'task': task}))
        self._raise_if_status_not(r, 201)

    def remove_task_from_job(self, task, job):
        r = requests.delete("%s/jobs/%s/tasks/%s" % (self.host, job, task))
        self._raise_if_status_not(r, 200)

    def list_tasks(self):
        r = requests.get("%s/tasks" % self.host)
        self._raise_if_status_not(r, 200)
        return r.json()

    def list_task_names(self):
        tasks = self.list_tasks()
        return [task['name'] for task in tasks]

    def add_task(self, name):
        r = requests.post("%s/tasks" % self.host, json.dumps({'name': name}))
        self._raise_if_status_not(r, 201)

    def get_task(self, name):
        r = requests.get("%s/tasks/%s" % (self.host, name))
        self._raise_if_status_not(r, 200)
        return r.json()

    def update_task(self, name, script):
        r = requests.put("%s/tasks/%s" % (self.host, name), data=json.dumps({'script': script}))
        self._raise_if_status_not(r, 200)

    def delete_task(self, name):
        r = requests.delete("%s/tasks/%s" % (self.host, name))
        self._raise_if_status_not(r, 200)

    def list_runs(self):
        r = requests.get("%s/runs" % self.host)
        self._raise_if_status_not(r, 200)
        return r.json()

    def list_run_ids(self):
        runs = self.list_runs()
        return [run['uuid'] for run in runs]

    def run_job(self, name):
        r = requests.post("%s/runs" % self.host, json.dumps({'job': name}))
        self._raise_if_status_not(r, 201)
        return r.json()

    def list_triggers(self):
        r = requests.get("%s/triggers" % self.host)
        self._raise_if_status_not(r, 200)
        return r.json()

    def list_trigger_names(self):
        triggers = self.list_triggers()
        return [trigger['name'] for trigger in triggers]

    def add_trigger(self, name):
        r = requests.post("%s/triggers" % self.host, data=json.dumps({'name': name}))
        self._raise_if_status_not(r, 201)

    def get_trigger(self, name):
        r = requests.get("%s/triggers/%s" % (self.host, name))
        self._raise_if_status_not(r, 200)
        return r.json()

    def update_trigger(self, name, cron):
        r = requests.put("%s/triggers/%s" % (self.host, name), data=json.dumps({'cron': cron}))
        self._raise_if_status_not(r, 200)

    def delete_trigger(self, name):
        r = requests.delete("%s/triggers/%s" % (self.host, name))
        self._raise_if_status_not(r, 200)

    def add_trigger_to_job(self, trigger, job):
        r = requests.post("%s/jobs/%s/triggers" % (self.host, job), data=json.dumps({'trigger': trigger}))
        self._raise_if_status_not(r, 201)

    def remove_trigger_from_job(self, trigger_idx, job):
        r = requests.delete("%s/jobs/%s/triggers/%s" % (self.host, job, trigger_idx))
        self._raise_if_status_not(r, 200)

    def _raise_if_status_not(self, r, status):
        if r.status_code != status:
            raise Exception(r.text)


class TestGoAPI(unittest.TestCase):
    def setUp(self):
        self.api = GoRunnerAPI("http://localhost:8090")

        self.test_job = "test_job999"
        self.test_task = "test_task999"
        self.test_trigger = "test_trigger999"
        self._clean()

    def tearDown(self):
        self._clean()

    def _clean(self):
        try:
            self.api.delete_job(self.test_job)
        except:
            pass
        try:
            self.api.delete_task(self.test_task)
        except:
            pass
        try:
            self.api.delete_trigger(self.test_trigger)
        except:
            pass

    def test_jobs(self):
        self.crud_test(self.api.list_job_names, self.api.delete_job, self.api.add_job, self.api.get_job)

    def test_tasks(self):
        self.crud_test(self.api.list_task_names, self.api.delete_task, self.api.add_task, self.api.get_task)

    def test_triggers(self):
        self.crud_test(self.api.list_trigger_names, self.api.delete_trigger, self.api.add_trigger, self.api.get_trigger)

    def test_adding_job_with_no_name(self):
        try:
            self.api.add_job("")
            self.fail()
        except Exception:
            pass

    def test_adding_job_with_no_payload(self):
        try:
            requests.post("%s/jobs" % self.api.host)
            self.fail()
        except Exception:
            pass

    def test_add_remove_task_to_job(self):
        self.api.add_job(self.test_job)
        self.api.add_task(self.test_task)
        self.api.add_task_to_job(self.test_task, self.test_job)

        job = self.api.get_job(self.test_job)
        self.assertIn(self.test_task, job['tasks'])

        self.api.remove_task_from_job(0, self.test_job)
        job = self.api.get_job(self.test_job)
        self.assertNotIn(self.test_task, job['tasks'])

    def test_add_remove_trigger_to_job(self):
        self.api.add_job(self.test_job)
        try:
            self.api.add_trigger(self.test_trigger)
        except:
            pass
        self.api.update_trigger(self.test_trigger, "* * * * * *")
        self.api.add_trigger_to_job(self.test_trigger, self.test_job)

        job = self.api.get_job(self.test_job)
        self.assertIn(self.test_trigger, job['triggers'])

        self.api.remove_trigger_from_job(self.test_trigger, self.test_job)
        job = self.api.get_job(self.test_job)
        self.assertNotIn(self.test_trigger, job['triggers'])

    def test_updating_task(self):
        self.api.add_task(self.test_task)
        task = self.api.get_task(self.test_task)
        self.assertEqual("", task['script'])

        self.api.update_task(self.test_task, "hello")
        task = self.api.get_task(self.test_task)
        self.assertEqual("hello", task['script'])

    def test_updating_trigger(self):
        self.api.add_trigger(self.test_trigger)
        trigger = self.api.get_trigger(self.test_trigger)
        self.assertEqual("", trigger['schedule'])
        self.api.update_trigger(self.test_trigger, "0 * * * *")
        trigger = self.api.get_trigger(self.test_trigger)
        self.assertEqual("0 * * * *", trigger['schedule'])

    def test_runs(self):
        self.api.add_job(self.test_job)
        self.api.add_task(self.test_task)
        self.api.add_task_to_job(self.test_task, self.test_job)
        uuid = self.api.run_job(self.test_job)['uuid']
        runs = self.api.list_run_ids()
        self.assertIn(uuid, runs)

    def crud_test(self, list_names, delete, add, get):
        test_name = "test999"

        names = list_names()
        if test_name in names:
            delete(test_name)
            names = list_names()
        self.assertNotIn(test_name, names)

        add(test_name)
        names = list_names()
        self.assertIn(test_name, names)

        thing = get(test_name)
        self.assertEqual(test_name, thing['name'])

        delete(test_name)
        names = list_names()
        self.assertNotIn(test_name, names)

if __name__ == "__main__":
    unittest.main()
