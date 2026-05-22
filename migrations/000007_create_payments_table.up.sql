create table payments (
  id bigint generated always as identity primary key,

  order_id bigint not null references orders (id),

  idempotency_key varchar(128) not null,

  provider varchar(50) not null,
  method varchar(50) not null,

  provider_payment_id varchar(255) not null,

  status varchar(20) not null default 'pending',

  amount integer not null,
  currency char(3) not null default 'THB',

  payment_url text null,
  qr_code_url text null,

  failure_reason text null,

  paid_at timestamptz null,
  failed_at timestamptz null,
  cancelled_at timestamptz null,
  expired_at timestamptz null,

  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now(),

  constraint payments_order_id_idempotency_key_unique
    unique (order_id, idempotency_key),

  constraint payments_provider_payment_id_unique
    unique (provider, provider_payment_id),

  constraint payments_idempotency_key_not_blank_check
    check (length(trim(idempotency_key)) > 0),

  constraint payments_provider_not_blank_check
    check (length(trim(provider)) > 0),

  constraint payments_method_not_blank_check
    check (length(trim(method)) > 0),

  constraint payments_status_check
    check (status in ('pending', 'succeeded', 'failed', 'cancelled', 'expired')),

  constraint payments_amount_check
    check (amount > 0),

  constraint payments_currency_check
    check (currency = 'THB'),

  constraint payments_paid_consistency_check
    check (
      (status = 'succeeded' and paid_at is not null)
      or
      (status <> 'succeeded' and paid_at is null)
    ),

  constraint payments_failed_consistency_check
    check (
      (status = 'failed' and failed_at is not null)
      or
      (status <> 'failed' and failed_at is null)
    ),

  constraint payments_cancelled_consistency_check
    check (
      (status = 'cancelled' and cancelled_at is not null)
      or
      (status <> 'cancelled' and cancelled_at is null)
    ),

  constraint payments_expired_consistency_check
    check (
      (status = 'expired' and expired_at is not null)
      or
      (status <> 'expired' and expired_at is null)
    )
);

create trigger payments_set_updated_at
before update on payments
for each row
execute function set_updated_at();