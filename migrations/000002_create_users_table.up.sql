create table users (
  id bigint generated always as identity primary key,

  name varchar(255) not null,

  email citext not null,
  password_hash text not null,

  role varchar(20) not null default 'user',
  status varchar(20) not null default 'active',

  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now(),
  deleted_at timestamptz null,

  constraint users_name_not_blank_check
    check (length(trim(name)) > 0),
  
  constraint users_email_unique unique (email),

  constraint users_role_check
    check (role in ('root', 'admin', 'user')),

  constraint users_status_check
    check (status in ('active', 'disabled', 'deleted')),

  constraint users_deleted_consistency_check
    check (
      (status = 'deleted' and deleted_at is not null)
      or
      (status <> 'deleted' and deleted_at is null)
    )
);

create trigger users_set_updated_at
before update on users
for each row
execute function set_updated_at();

create index users_status_idx
on users (status);

insert into users (
  name,
  email,
  password_hash,
  role,
  status
) values
('UserA', 'a@example.com', '$2a$10$SYPaNeIpqtDFX/heLmpgZuwsXQkbOsEVhi/YXfyAmo.grO6W3WrOa', 'root', 'active'),
('UserB', 'b@example.com', '$2a$10$SYPaNeIpqtDFX/heLmpgZuwsXQkbOsEVhi/YXfyAmo.grO6W3WrOa', 'admin', 'active'),
('UserC', 'c@example.com', '$2a$10$SYPaNeIpqtDFX/heLmpgZuwsXQkbOsEVhi/YXfyAmo.grO6W3WrOa', 'user', 'active'),
('UserD', 'd@example.com', '$2a$10$SYPaNeIpqtDFX/heLmpgZuwsXQkbOsEVhi/YXfyAmo.grO6W3WrOa', 'user', 'disabled');