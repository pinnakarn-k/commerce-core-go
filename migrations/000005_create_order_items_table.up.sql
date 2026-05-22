create table order_items (
  id bigint generated always as identity primary key,

  order_id bigint not null references orders (id),
  product_id bigint not null references products (id),

  product_name varchar(255) not null,
  product_sku citext not null,

  quantity integer not null,

  unit_price_amount integer not null,
  line_total_amount integer not null,
  currency char(3) not null default 'THB',

  created_at timestamptz not null default now(),

  constraint order_items_order_product_unique
    unique (order_id, product_id),

  constraint order_items_quantity_check
    check (quantity > 0),

  constraint order_items_unit_price_amount_check
    check (unit_price_amount > 0),

  constraint order_items_line_total_amount_check
    check (line_total_amount = quantity * unit_price_amount),

  constraint order_items_currency_check
    check (currency = 'THB')
);

create index order_items_order_id_idx
on order_items (order_id);

create index order_items_product_id_idx
on order_items (product_id);