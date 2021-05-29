# What is this?

`slack-bot-fleet` is PoC for Slack App.

## service-broker

```
curl -XPOST -H "Content-Type: application/json" -H "X-Slack-Signature:0123456789abcdef" -H "X-Slack-Request-Timestamp:$(date +%s)" -d @- localhost:8080 <<EOF
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

## example

```
grpcurl -plaintext -d @ localhost:8080 example.Example.Reply <<EOM
{
    "reaction": "reaction_added",
    "user": "U020VK32D63",
    "item": {
        "channel": "C0213JYV3HC",
        "ts": "1620581892.006600"
    }
}
EOM
```
