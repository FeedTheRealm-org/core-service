BEGIN;

CREATE TABLE IF NOT EXISTS zones_subscriptions (
  id                     UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id                UUID        NOT NULL UNIQUE,
  stripe_customer_id     VARCHAR(255) NOT NULL,
  stripe_subscription_id VARCHAR(255),
  total_slots            INT         NOT NULL DEFAULT 0,
  used_slots             INT         NOT NULL DEFAULT 0,
  amount_due             DECIMAL(10,2) NOT NULL DEFAULT 0,
  status                 VARCHAR(50) NOT NULL DEFAULT 'inactive',
  next_billing_date      TIMESTAMPTZ,
  created_at             TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at             TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

COMMIT;
