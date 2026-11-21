# Migrating the ingest pipeline to streaming

This document proposes moving the nightly batch ingest to a streaming design. It covers the reasons, the target architecture, a rollout plan and the open questions. It also happens to exercise every bit of Markdown that **md2html** supports, so it doubles as the visual test page.

> The batch job has finished late on 11 of the last 30 nights. Each late finish delays the morning reports, and each delay costs someone an hour of re-checking numbers they already checked.

## Why change

The current job reads every file dropped into the bucket since the previous run, transforms it, and writes the result in one transaction. That worked when the volume was a few thousand rows. It is now about 40 million rows a night, and three things have gone wrong:

1. Run time grows linearly with volume, and volume doubled this year.
2. A single malformed file fails the *entire* run, not just that file.
3. Nothing is visible until everything is done, so a 6-hour job tells you nothing for 6 hours.

A streaming design fixes all three: work is done as files arrive, failures are isolated per file, and progress is visible continuously. It is not free, though. We give up the simple *all-or-nothing* guarantee and must handle ~~exactly-once~~ at-least-once delivery with idempotent writes instead.

### What stays the same

- The bucket layout and file formats.
- The downstream tables and their schemas.
- The morning report queries, which should not notice the change.

## Architecture

Files land in the bucket, an event goes onto a queue, and a pool of workers picks up events and processes one file each. Writes are keyed on `(source_id, row_hash)` so replays are harmless.

| Component     | Today              | Proposed                 | Owner     |
|---------------|--------------------|--------------------------|-----------|
| Trigger       | Cron at 01:00      | Bucket notification      | Platform  |
| Unit of work  | Whole night        | One file                 | Data      |
| Failure scope | Everything         | One file, sent to retry  | Data      |
| Throughput    | ~1.8k rows/s       | ~12k rows/s (est.)       | —         |
| Cost / month  | $410               | $520 (est.)              | Finance   |

### The worker loop

The worker is deliberately boring. It pulls one message, processes one file, and acknowledges only after the write commits.

```go
func (w *Worker) Run(ctx context.Context) error {
	for {
		msg, err := w.queue.Receive(ctx)
		if err != nil {
			return fmt.Errorf("receive: %w", err)
		}
		// Idempotent: replays of the same file are no-ops.
		if err := w.process(ctx, msg.FileKey); err != nil {
			w.log.Warn("file failed, will retry", "key", msg.FileKey, "err", err)
			msg.Nack()
			continue
		}
		msg.Ack()
	}
}
```

The upsert does the de-duplication:

```sql
INSERT INTO events (source_id, row_hash, payload, ingested_at)
VALUES ($1, $2, $3, now())
ON CONFLICT (source_id, row_hash) DO NOTHING;
```

Configuration lives in one small file:

```yaml
queue:
  name: ingest-events
  visibility_timeout: 300s
workers:
  min: 2
  max: 24
  scale_on: queue_depth   # one worker per 500 queued messages
retry:
  attempts: 5
  backoff: exponential
```

To run a worker locally, use `make worker` or call the binary directly:

```bash
INGEST_QUEUE=ingest-events-dev ./bin/worker --concurrency 4 --log-level debug
```

#### A note on ordering

Files may be processed out of order. Nothing downstream depends on arrival order today, but the reconciliation report does assume that a day's data is complete by 06:00. We keep that guarantee with a watermark check rather than with ordering.[^watermark]

## Rollout plan

The rollout runs in four steps over roughly three weeks. Each step can be reversed by flipping the `ingest.mode` flag back to `batch`.

- [x] Build the worker and deploy it to staging
- [x] Run the worker in **shadow mode** beside the batch job and diff the outputs
- [ ] Switch one low-risk source to streaming in production
- [ ] Switch the remaining sources, one per day
- [ ] Delete the batch job

Keyboard aside for whoever is on call: press <kbd>Ctrl</kbd> + <kbd>K</kbd> in the dashboard to jump straight to the queue depth panel.

## Risks

**Cost overrun.** The estimate assumes workers scale down to two overnight. If they don't, the bill roughly doubles. We will alert at $600/month.

**Poison files.** A file that fails all five retries goes to a dead-letter queue. Someone has to look at that queue, so it gets its own alert and a runbook entry.

**Duplicate rows from a hash change.** If anyone changes how `row_hash` is computed, every replayed row looks new. The hash function is therefore versioned, and the version is part of the key.

## Open questions

1. Do we need per-source rate limits, or is the global worker cap enough?
2. Should the dead-letter queue page someone, or just open a ticket?
3. Is $520/month acceptable to Finance, given the estimate's uncertainty?

---

See the [design review notes](https://example.com/review) for the discussion so far, or reach the team at <https://example.com/team>.

[^watermark]: The watermark is the oldest unprocessed file's arrival time. When it passes 06:00 for a given day, that day is complete.
