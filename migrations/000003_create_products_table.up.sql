create table products (
  id bigint generated always as identity primary key,

  sku citext not null,
  name varchar(255) not null,
  description text null,

  price_amount integer not null,
  currency char(3) not null default 'THB',

  stock_qty integer not null,

  status varchar(20) not null default 'active',

  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now(),
  deleted_at timestamptz null,

  constraint products_sku_unique unique (sku),

  constraint products_price_amount_check
    check (price_amount > 0),

  constraint products_stock_qty_check
    check (stock_qty >= 0),

  constraint products_currency_check
    check (currency = 'THB'),

  constraint products_status_check
    check (status in ('active', 'disabled', 'deleted')),

  constraint products_deleted_consistency_check
    check (
      (status = 'deleted' and deleted_at is not null)
      or
      (status <> 'deleted' and deleted_at is null)
    )
);

create trigger products_set_updated_at
before update on products
for each row
execute function set_updated_at();

create index products_status_idx
on products (status);

insert into products (
  sku,
  name,
  description,
  price_amount,
  currency,
  stock_qty,
  status
) values
('SKU-001', 'Wireless Mouse', 'Ergonomic wireless mouse', 59000, 'THB', 100, 'active'),
('SKU-002', 'Mechanical Keyboard', 'RGB mechanical keyboard', 249000, 'THB', 50, 'active'),
('SKU-003', 'USB-C Cable', '1m USB-C to USB-C cable', 19900, 'THB', 200, 'active'),
('SKU-004', 'Laptop Stand', 'Adjustable aluminum stand', 89000, 'THB', 80, 'active'),
('SKU-005', 'Noise Cancelling Headphones', 'Over-ear ANC headphones', 359000, 'THB', 30, 'active'),
('SKU-006', 'Webcam 1080p', 'Full HD webcam with mic', 129000, 'THB', 60, 'active'),
('SKU-007', 'Portable SSD 1TB', 'High-speed external SSD', 499000, 'THB', 25, 'active'),
('SKU-008', 'Bluetooth Speaker', 'Compact wireless speaker', 159000, 'THB', 70, 'active'),
('SKU-009', 'Gaming Mouse Pad', 'Large anti-slip surface', 39000, 'THB', 150, 'active'),
('SKU-010', 'USB Hub 6-in-1', 'Multiport USB-C hub', 99000, 'THB', 90, 'active'),
('SKU-011', 'Out of Stock Item', 'No stock available', 100000, 'THB', 0, 'active'),
('SKU-012', 'Disabled Product', 'Not available for sale', 150000, 'THB', 10, 'disabled');