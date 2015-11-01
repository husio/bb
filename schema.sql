BEGIN;


CREATE TABLE IF NOT EXISTS users (
	user_id  serial PRIMARY KEY,
	name     text NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS topics (
	topic_id   serial PRIMARY KEY,
	title      text NOT NULL UNIQUE,
	author_id  integer NOT NULL REFERENCES users(id),
	created    timestamptz NOT NULL,
	updated    timestamptz NOT NULL,
	replies    integer NOT NULL DEFAULT 0
);


-- Update replies counter by counting all assigned messages and "updated" date
CREATE OR REPLACE FUNCTION update_topic_on_messages_change()
RETURNS TRIGGER AS
$$
BEGIN
    IF (TG_OP = 'INSERT') THEN
        UPDATE topics
            SET
                replies = (SELECT COUNT(*) FROM messages WHERE topic_id = NEW.topic_id) - 1,
                updated = NEW.created
            WHERE topic_id = NEW.topic_id;
        RETURN NEW;
    END IF;

    UPDATE topics
        SET
            replies = (SELECT COUNT(*) FROM messages WHERE topic_id = OLD.topic_id) - 1
        WHERE topic_id = OLD.topic_id;
    RETURN OLD;

END
$$
LANGUAGE plpgsql;


DROP TRIGGER IF EXISTS update_topic_on_messages_change ON messages;
CREATE TRIGGER update_topic_on_messages_change AFTER INSERT OR DELETE ON messages
    FOR EACH ROW EXECUTE PROCEDURE update_topic_on_messages_change();



CREATE TABLE IF NOT EXISTS messages (
	message_id serial PRIMARY KEY,
	topicid    integer NOT NULL REFERENCES topics(id),
	author_id  integer NOT NULL REFERENCES users(id),
	content    text NOT NULL,
	created    timestamptz NOT NULL
);


COMMIT;
