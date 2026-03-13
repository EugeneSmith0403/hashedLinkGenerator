CREATE TABLE IF NOT EXISTS link_clicks (
    link_id    Int64,
    clicked_at DateTime DEFAULT now(),

    -- Network
    ip              String,
    forwarded_for   String,
    real_ip         String,
    remote_addr     String,
    remote_port     String,
    country         String,

    -- Headers
    user_agent      String,
    accept          String,
    accept_language String,
    accept_encoding String,
    origin          String,
    referer         String,

    -- Device
    device_type     String,
    os              String,
    browser         String,

    -- Security
    fingerprint      String,
    request_id       String,
    forwarded_proto  String,
    forwarded_host   String,
    forwarded_port   String,
    scheme           String
) ENGINE = ReplacingMergeTree()
PARTITION BY toYYYYMM(clicked_at)
ORDER BY (link_id, clicked_at, request_id);
