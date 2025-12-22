CREATE TABLE IF NOT EXISTS outbox(
  id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  aggregate_type varchar(255) not null,
  aggregate_id varchar(255) not null,
  event varchar(255) not null,
  payload json not null,
  retry_at timestamp null,
  retry_count int null,
  sent_at timestamp null
)
