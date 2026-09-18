CREATE TABLE inbox (
    consumer_name VARCHAR(100) NOT NULL,
    message_id VARCHAR(255) NOT NULL,
    payload_hash VARCHAR(64) NOT NULL,
    received_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    processed_at TIMESTAMP WITH TIME ZONE NULL,
    PRIMARY KEY (consumer_name, message_id)
);