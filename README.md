# slack-bot-fleet [![ci](https://github.com/ishii1648/slack-bot-fleet/actions/workflows/ci.yml/badge.svg)](https://github.com/ishii1648/slack-bot-fleet/actions/workflows/ci.yml) [![codecov](https://codecov.io/gh/ishii1648/slack-bot-fleet/branch/main/graph/badge.svg?token=YBCUPT1WMP)](https://codecov.io/gh/ishii1648/slack-bot-fleet)

`slack-bot-fleet` is PoC for Slack App.

## test request

### service-broker

```
curl --dump-header - -XPOST \
-H "Content-Type: application/json" \
-H "X-Slack-Signature:0123456789abcdef" \
-H "X-Cloud-Trace-Context:0123456789abcdef0123456789abcdef/123;o=1" \
-H "X-Slack-Request-Timestamp:$(date +%s)" -d @- localhost:8080 <<EOF
{
    "type": "event_callback",
    "event": {
        "type": "reaction_added",
        "user": "U020VK32D63",
        "item": {
            "type": "message",
            "channel": "C0213JYV3HC",
            "ts": "1620581892.006600"
        },
        "reaction": "ok_hand",
        "item_user": "U020VK32D63",
        "event_ts": "1622040734.000100"
    }
}
EOF
```

### example

```
curl --dump-header - -XPOST \
-H "Content-Type: application/json" \
-H "X-Cloud-Trace-Context:0123456789abcdef0123456789abcdef/123;o=1" \
-d @- localhost:8080 <<EOF
{
    "reaction": "ok_hand",
    "user": "s_ishii",
    "itemChannel": "development",
    "itemTimestamp": "1622040734.000100"
}
EOF
```
