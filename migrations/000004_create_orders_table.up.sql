create table orders (
  id bigint generated always as identity primary key,

  user_id bigint not null references users (id),

  idempotency_key varchar(128) not null,

  status varchar(20) not null default 'pending',

  total_amount integer not null,
  currency char(3) not null default 'THB',

  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now(),
  paid_at timestamptz null,
  cancelled_at timestamptz null,

  constraint idempotency_key_not_blank_check
    check (length(trim(idempotency_key)) > 0),
  
  constraint orders_user_id_idempotency_key_unique
    unique (user_id, idempotency_key),

  constraint orders_status_check
    check (status in ('pending', 'paid', 'cancelled')),

  constraint orders_total_amount_check
    check (total_amount >= 0),

  constraint orders_currency_check
    check (currency = 'THB'),

  constraint orders_paid_consistency_check
    check (
      (status = 'paid' and paid_at is not null and cancelled_at is null)
      or
      (status <> 'paid' and paid_at is null)
    ),

  constraint orders_cancelled_consistency_check
    check (
      (status = 'cancelled' and cancelled_at is not null and paid_at is null)
      or
      (status <> 'cancelled' and cancelled_at is null)
    )
);

create trigger orders_set_updated_at
before update on orders
for each row
execute function set_updated_at();

create index orders_user_id_idx
on orders (user_id);

create index orders_status_idx
on orders (status);