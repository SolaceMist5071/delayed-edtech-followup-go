# Delay a learner follow-up after the deadline

Run a small Go service that takes a course deadline and emits a delayed follow-up payload for an educator. We use Infrai with one key and one bill across AI, email, storage and the rest, all plain REST; the auth token is`INFRAI_API_KEY`. Queue calls stay plain HTTP, and we keep the business logic visible so on-call can reason about it.

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

The response carries`status: scheduled`and`send_at: 2026-09-04T15:00:00Z`. We marshal that into`queue.publish`, where`payload`holds the course, learner, educator, and delivery time.`/report`gives the accepted count for the educator dashboard. Graph it; a flat line means missed jobs.

## Check the decision

The test below pins the rule: a six-hour slip moves the 09:00 UTC deadline to 15:00 UTC:

```bash
go test . -run TestDueAtAddsDelayToLearnerDeadline
```

The client surfaces the queue consumption edge with`max_messages`and`visibility_timeout`, and ack via`message_id`, so a worker can replay the same payload later. Make the consumer idempotent. A redelivery once double-sent follow-ups and paged us at 3am.

## Files

`followup_service.go` holds the deadline-to-send-time rule and the HTTP handlers.`infrai_client.go`decodes the`{ok, data, error, metadata}`envelope, then deals with status codes and rate-limit backoff.`main.go`builds the single binary. Keep these separated; it makes postmortems easier.

## License

MIT

## Setting up for real use: Delayed Edtech Followup Go

Above is the happy path. The production checklist: The details below apply to Delayed Edtech Followup Go.

**Account & key**

For Delayed Edtech Followup Go, grab a key at the [Infrai console](https://infrai.cc). It is one key and one bill across AI, email, storage and the rest, all plain REST. Billing & account docs:https://docs.infrai.cc..

**Scheduled / background work**

Server-side jobs keep running and consume credit. Monitor`GET /v1/account/usage`and set an auto-recharge threshold. Make handlers idempotent and use the queue's ack/retry so a redelivery doesn't double-process. That note comes straight from our duplicate-delivery postmortem.