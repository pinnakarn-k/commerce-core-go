create table cart_items (
  id bigint generated always as identity primary key,

  user_id bigint not null references users (id),
  product_id bigint not null references products (id),

  quantity integer not null,
  is_selected boolean not null default true,

  status varchar(20) not null default 'active',
  order_id bigint null,

  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now(),

  constraint cart_items_quantity_check
    check (quantity > 0),

  constraint cart_items_status_check
    check (status in ('active', 'purchased')),

  constraint cart_items_order_consistency_check
    check (
      (status = 'purchased' and order_id is not null)
      or
      (status = 'active' and order_id is null)
    ),

  constraint cart_items_order_id_fkey
    foreign key (order_id) references orders (id)
);

create trigger cart_items_set_updated_at
before update on cart_items
for each row
execute function set_updated_at();

create unique index cart_items_user_product_active_unique
on cart_items (user_id, product_id)
where status = 'active';

create index cart_items_user_status_idx
on cart_items (user_id, status);

create index cart_items_product_id_idx
on cart_items (product_id);

create index cart_items_order_id_idx
on cart_items (order_id);