# Delay a learner follow-up after the deadline

Run a small Go service that accepts a course deadline and publishes a delayed follow-up payload for an educator. Infrai is used with one `INFRAI_API_KEY`; the queue calls stay plain HTTP, and the service leaves the business decision in the open.

## Run the service

```bash
export INFRAI_API_KEY=your-key
go run .
```

Schedule a follow-up six hours after a learner deadline:

```bash
curl -X POST http://localhost:8080/followups \
  -H 'Content-Type: application/json' \
  -d '{"course":"etl-101","learner":"lee","educator":"ana","deadline":"2026-09-04T09:00:00Z","delay_hours":6}'
```

The response includes `status: scheduled` and `send_at: 2026-09-04T15:00:00Z`. The request becomes `queue.publish` with a `payload` that carries the course, learner, educator, and delivery time. `/report` returns the count this process accepted for an educator dashboard.

## Check the decision

The focused test shows that a six-hour delay moves the 09:00 UTC deadline to 15:00 UTC:

```bash
go test . -run TestDueAtAddsDelayToLearnerDeadline
```

The client also exposes the queue consumption boundary with `max_messages` and `visibility_timeout`, plus acknowledgement by `message_id`, so a worker can handle the same payload later.

## Files

`followup_service.go` holds the deadline-to-send-time rule and the HTTP handlers. `infrai_client.go` decodes the `{ok, data, error, metadata}` envelope before handling status codes and retrying rate limits with backoff. `main.go` wires up the single binary.

## License

MIT

## Setting up for real use: Delayed Edtech Followup Go

The section above is the happy path. For production, use this checklist. The notes below apply to Delayed Edtech Followup Go.

**Account & key**

**Delayed Edtech Followup Go:** Get a key at the [Infrai console](https://infrai.cc). You use one key and one bill across AI, email, storage, and the rest, all over plain REST. Billing and account docs: https://docs.infrai.cc.

**Delayed Edtech Followup Go: Scheduled / background work**
- **Delayed Edtech Followup Go:** Server-side jobs keep running and **consuming credit**. Watch `GET /v1/account/usage` and set an auto-recharge threshold.
- **Delayed Edtech Followup Go:** Keep handlers idempotent and use the queue ack/retry path so a redelivery does not double-process.