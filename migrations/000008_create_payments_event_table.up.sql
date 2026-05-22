create table payment_events (
  id bigint generated always as identity primary key,

  provider varchar(50) not null,
  provider_event_id varchar(255) not null,

  payment_id bigint not null references payments (id),

  event_type varchar(100) not null,
  payload jsonb not null,

  created_at timestamptz not null default now(),

  constraint payment_events_provider_not_blank_check
    check (length(trim(provider)) > 0),

  constraint payment_events_provider_event_id_not_blank_check
    check (length(trim(provider_event_id)) > 0),

  constraint payment_events_event_type_not_blank_check
    check (length(trim(event_type)) > 0),

  constraint payment_events_provider_event_unique
    unique (provider, provider_event_id)
);