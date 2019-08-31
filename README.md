# Go-SQS

A simple SQS load test tool written in Go.

Performs concurrent sends/receives over an SQS queue, and computes average throughput and latency.

# TODO

- [ ] compute percentiles
- [ ] add flags
- [ ] more accurate latency measurement (include a timestamp in the message)
