# promify-goss

## Convert your json formated [goss](https://github.com/goss-org/goss) tests to something more prom friendly

Goss is a fantastic server testing and validation tool that is blazing fast. What it lacked is a way to ship these individual test results into Prometheus.

Promify-goss is a simple text reformatter written in go to convert json formatted goss tests into prometheus style metrics. Promify-goss captures a wide variety of attributes to be used as query parameters.

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

These key data points make event corelation trivial. Additionally this allows just about any 3rd party monitoring tool another way to collect pass/fail data. Be it Prometheus Alert Manager, Grafana Alerts and even Zabbix.

---

## Basic Usage

Promified doen't have many options.

```bash
± ~/gocode/promify-goss (main U:2 ?:2 ✗) $ promify-goss --help
  -dir string
        Where do you want to store your .prom (default "/var/lib/node_exporter/textfile_collector")
  -name string
        Name your .prom here. Extension will be appended upon writing
  -uri string
        Goss endpoint to fetch the test results from.

```

You must specify the url of your Goss endpoint, or, pipe a goss validation test. Each method requires at least the name flag and worth noting an unspecified path will use the default        textfile_collector path shipped by node_exporter. If your node_exporter deployment has a custom textfile_collector you will need to specify that path or update your version of the go code to make your path the default and rebuild the program.

---

### Let's look at a short **"two test"** example

in our demo.yaml test we are expecting the file /srv/down not to exist and http://httpbun.org/get to return a 200 respose.

```bash
± ~/gocode/promify-goss (main U:2 ?:1 ✗) $ goss -g ./examples/demo.yaml validate -f tap
1..2
ok 1 - File: /srv/down: exists: matches expectation: [false]
ok 2 - HTTP: http://httpbun.org/get: status: matches expectation: [200]
```

Below is the data returned with the json outputter.At first glance we can already see json exposes more details about each test.

```bash
± ~/gocode/promify-goss (main U:2 ?:1 ✗) $ goss -g ./examples/demo.yaml validate -f json -o pretty
{
    "results": [
        {
            "duration": 52102,
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
            "duration": 523689683,
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
        "summary-line": "Count: 2, Failed: 0, Duration: 0.524s",
        "test-count": 2,
        "total-duration": 523920650
    }
}
```

Now if we pipe the validation command through promified we get a prom friendly file to be scraped and data to be queried like any other prometheus metric.

```bash
± ~/gocode/promify-goss (main U:2 ?:1 ✗) $ goss -g ./examples/demo.yaml validate -f json | ./promify-goss -path ./ -name ./demo.prom ; cat ./demo.prom
goss_result_file{property="/srv/down",resource="exists",skipped="false"} 0
goss_result_file_duration{property="/srv/down",resource="exists",skipped="false"} 53428
goss_result_http{property="http://httpbun.org/get",resource="status",skipped="false"} 0
goss_result_http_duration{property="http://httpbun.org/get",resource="status",skipped="false"} 525555851
goss_results_summary{textfile="./demo.prom",name="tested"} 2
goss_results_summary{textfile="./demo.prom",name="failed"} 0
goss_results_summary{textfile="./demo.prom",name="duration"} 525787516
```

### Using it yourself

1. Install: `curl -fsSL https://goss.rocks/install | sh`
2. Clone this repo:
3. Build promified: `go build -o promify-goss .`
4. Create a cron to continually run your test.


### To-Dos

- write tests
