BEGIN;

CREATE TABLE IF NOT EXISTS processed_stripe_webhook_events (
  event_id TEXT PRIMARY KEY,
  session_id TEXT NOT NULL UNIQUE,
  event_type VARCHAR(100) NOT NULL,
  user_id UUID NOT NULL,
  created_at TIMESTAMPTZ DEFAULT NOW()
);

COMMIT;
