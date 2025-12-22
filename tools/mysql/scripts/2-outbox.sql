CREATE TABLE IF NOT EXISTS outbox(
  id bigint auto_increment,
  aggregate_type varchar(255) not null,
  aggregate_id varchar(255) not null,
  event varchar(255) not null,
  payload json not null,
  retry_at datetime null,
  retry_count int null,
  sent_at datetime null,
  primary key(id)
)
