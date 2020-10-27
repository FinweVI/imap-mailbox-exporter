# imap-mailbox-exporter
Prometheus exporter for the status of an IMAP mailbox.

## Metrics

| Name | Description |
| -------- | -------- |
| imap_up | Wether the IMAP can be reached or not. |
| imap_total_message_count | Total number of messages in the mailbox. |
| imap_unseen_message | Sequence number of the first unseen message. Default to 0. |

## Arguments

| Name | Default | Description |
| -------- | -------- | -------- |
| imap.server | None | IMAP Server. |
| imap.username | None | IMAP Username. |
| imap.password | None | IMAP Password. |
| imap.mailbox | INBOX | IMAP Folder to watch. |
| imap.query.interval | 120 | Minimum interval between queries to the IMAP server in seconds. |
| listen.address | 127.0.0.1:9117 | Address to listen on for web interface and telemetry. |
| metrics.endpoint | /metrics | Path under which to expose metrics. |

### Env Var

| Name | Description |
| -------- | -------- |
| IMAP_SERVER | IMAP Server. |
| IMAP_USERNAME | IMAP Username. |
| IMAP_PASSWORD | IMAP Password. |
| IMAP_MAILBOX | IMAP Folder to watch. |
| IMAP_QUERY_INTERVAL | Minimum interval between queries to the IMAP server in seconds. |
| LISTEN_ADDRESS | Address to listen on for web interface and telemetry. |
| METRICS_ENDPOINT | Path under which to expose metrics. |

### Usage

```
Usage of ./imap-mailbox-exporter:
  -imap.mailbox string
    	IMAP mailbox to query
  -imap.password string
    	IMAP password for login
  -imap.query.interval string
    	Minimum interval between queries to IMAP server in seconds
  -imap.server string
    	IMAP server to query
  -imap.username string
    	IMAP username for login
  -listen.address string
    	Address to listen on for web interface and telemetry
  -metrics.endpoint string
    	Path under which to expose metrics
```

Note: Command Line Arguments takes precedence over ENV VARs.


## Issues & Contribution
All bug report, packaging requests, features requests or PR are accepted.
I mainly forked this exporter for my personal usage but I'll be happy to hear about your needs.
