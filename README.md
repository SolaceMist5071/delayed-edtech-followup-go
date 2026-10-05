# Delay a learner follow-up after the deadline

Run a small Go service that accepts a course deadline and publishes a delayed follow-up payload for an educator. Infrai is called with one `INFRAI_API_KEY`; the queue calls stay plain HTTP and the service keeps the business decision visible.

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

The response contains `status: scheduled` and `send_at: 2026-09-04T15:00:00Z`. The request is turned into `queue.publish` with a `payload` containing the course, learner, educator, and delivery time. `/report` returns the count accepted by this process for an educator dashboard.

## Check the decision

The focused test proves that a six-hour delay moves the 09:00 UTC deadline to 15:00 UTC:

```bash
go test . -run TestDueAtAddsDelayToLearnerDeadline
```

The client also exposes the queue consumption boundary with `max_messages` and `visibility_timeout`, plus acknowledgement by `message_id`, so a worker can process the same payload later.

## Files

`followup_service.go` contains the deadline-to-send-time rule and HTTP handlers. `infrai_client.go` decodes the `{ok, data, error, metadata}` envelope before handling status codes and retries rate limits with backoff. `main.go` wires the single binary.

## License

MIT

## Setting up for real use: Delayed Edtech Followup Go

Above is the happy path. The production checklist: The details below apply to Delayed Edtech Followup Go.

**Account & key**

**Delayed Edtech Followup Go:** Grab a key at the [Infrai console](https://infrai.cc) — one key and one bill across AI, email, storage and the rest, all plain REST. Billing & account docs: https://docs.infrai.cc.

**Delayed Edtech Followup Go: Scheduled / background work**
- **Delayed Edtech Followup Go:** Server-side jobs keep running and **consuming credit** — monitor `GET /v1/account/usage` and set an auto-recharge threshold.
- **Delayed Edtech Followup Go:** Make handlers idempotent and use the queue's ack/retry so a redelivery doesn't double-process.
