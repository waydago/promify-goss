# Promify-goss

## Convert your json formated [goss](https://github.com/goss-org/goss) tests to something more prom friendly

Promify-goss is a simple text reformatter written in go to reformat json formatted goss tests into prometheus style metrics. Promified captures a wide variety of attributes to be used as query parameters.

- test-type
- resource-id
- resource-type
- property
- duration
- expected
- found
- result
- skipped
- successful
- total-duration
- failed-count
- test-count

These key data points make event corelation trivial. Additionally this allows just about any 3rd party monitoring tool another way to collect pass/fail data. Be it Alert Manager, Grafana and even Zabbix.

---

#### Let's look at a short **"two test"** example

> in our demo.yaml test we are expecting the file /srv/down not to exist and http://httpbun.org/get to return a 200 respose.

```bash
± ~/gocode/promify-goss (main ✓) $ goss -g ./examples/demo.yaml validate -f tap
1..2
ok 1 - File: /srv/down: exists: matches expectation: [false]
ok 2 - HTTP: http://httpbun.org/get: status: matches expectation: [200]
```

Below is the data returned with the json outputter.At first glance we can already see json exposes more details about each test.

```json
± ~/gocode/promify-goss (main ✓) $ goss -g ./examples/demo.yaml validate -f json -o pretty
{
    "results": [
        {
            "duration": 52381,
            "err": null,
            "expected": [
                "false"
            ],
            "found": [
                "false"
            ],
            "human": "",
            "meta": null,
            "property": "exists",
            "resource-id": "/srv/down",
            "resource-type": "File",
            "result": 0,
            "skipped": false,
            "successful": true,
            "summary-line": "File: /srv/down: exists: matches expectation: [false]",
            "test-type": 0,
            "title": ""
        },
        {
            "duration": 90017170,
            "err": null,
            "expected": [
                "200"
            ],
            "found": [
                "200"
            ],
            "human": "",
            "meta": null,
            "property": "status",
            "resource-id": "http://httpbun.org/get",
            "resource-type": "HTTP",
            "result": 0,
            "skipped": false,
            "successful": true,
            "summary-line": "HTTP: http://httpbun.org/get: status: matches expectation: [200]",
            "test-type": 0,
            "title": ""
        }
    ],
    "summary": {
        "failed-count": 0,
        "summary-line": "Count: 2, Failed: 0, Duration: 0.090s",
        "test-count": 2,
        "total-duration": 90394525
    }
}
```

Now if we pipe the validation command through promified we get a prom friendly file to be scraped and data to be queried like any other prometheus metric.

```bash
± ~/gocode/promify-goss (main ✓) $ goss -g ./examples/demo.yaml validate -f json | ./promify-goss -path ./ -name demo.prom ; cat ./demo.prom
goss_result_file{property="/srv/down",resource="exists",skipped="false"} 0
goss_result_file_duration{property="/srv/down",resource="exists",skipped="false"} 45118
goss_result_http{property="http://httpbun.org/get",resource="status",skipped="false"} 0
goss_result_http_duration{property="http://httpbun.org/get",resource="status",skipped="false"} 60305759
goss_results_summary{textfile="demo.prom",name="tested"} 2
goss_results_summary{textfile="demo.prom",name="failed"} 0
goss_results_summary{textfile="demo.prom",name="duration"} 60544267
```

## Basic Usage

Promify-goss doen't have many options since it's just supposed to do one thing pretty alright.

```bash
± ~/gocode/promify-goss (master U:2 ?:1 ✗) $ ./promify-goss --help
  -name string
     Name your .prom
  -path string
     Where to store the .prom file (default "/var/lib/node_exporter/textfile_collector")
  -uri string
     Goss endpoint to fetch data from.
```

You can either specify the url of your Goss endpoint, or, pipe a goss validation test into promify. Each method requires at least the name flag and worth noting an unspecified path will use the default textfile_collector path shipped by node_exporter. if your node_exporter deployment has a custom textfile_collector you will need to specify that path or update your version of the go code to make your path the default and rebuild the program.

### Using it yourself

1. Install `curl -fsSL https://goss.rocks/install | sh`
2. Clone this repo: 
3. Build promify-goss: `go build -o promify-goss .`

### Resources

### To-Dos

- add Taskfile
- write tests
- improve pipeline
