# promify-goss

## What's promify-goss?

`promify-goss` is a nifty tool written in Go that turns your JSON-formatted goss test results into Prometheus-friendly metrics. It's super handy for getting detailed test data into your monitoring setup.
Key Features

- Converts goss test results to Prometheus metrics.
- Captures lots of details like test type, resource ID, duration, and more.
- Easy to integrate with monitoring tools like Prometheus, Grafana, etc.

## How to Use

1. Set Up: First, you need a goss test output in JSON. You can get this by running your goss tests with the -f json flag.
2. Run promify-goss: Pipe the JSON output to promify-goss. Don't forget to name your output file using the -name flag. If you don't specify a path, it uses the default /var/lib/node_exporter/textfile_collector.

### Example:

```bash
    goss -g your_test.yaml validate -f json | promify-goss -name your_output.prom
```

Check the Output: You'll get a .prom file with all your test data in a format that Prometheus loves.

## Behind the Scenes

promify-goss.go does the magic. It reads the piped JSON, extracts key info, and formats everything into Prometheus metrics. It's got some neat Go code handling JSON unmarshalling, command-line arguments, and file writing.
Building and Installing
1. Clone this repo.    Build it with Go: go build -o promify-goss .
2. Set up a cron job or similar to run your tests regularly.

## To-Dos

- Write some tests for the tool.

