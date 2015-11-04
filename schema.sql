BEGIN;


CREATE TABLE IF NOT EXISTS users (
	user_id  serial PRIMARY KEY,
	login    text NOT NULL UNIQUE
);


CREATE TABLE IF NOT EXISTS categories (
    category_id  serial PRIMARY KEY,
    name         text NOT NULL,
    description  text NOT NULL,
    topics_count integer NOT NULL DEFAULT 0,
    color        integer DEFAULT 16777215 -- RGB:255,255,255
);

CREATE TABLE IF NOT EXISTS topics (
	topic_id    serial PRIMARY KEY,
	title       text NOT NULL,
	author_id   integer NOT NULL REFERENCES users(user_id),
    category_id integer NOT NULL REFERENCES categories(category_id),
	created     timestamptz NOT NULL,
	updated     timestamptz NOT NULL,
	replies     integer NOT NULL DEFAULT 0
);

CREATE INDEX topics_updated_idx ON topics(updated);

-- Update replies counter by inc/dec-rementing counter
CREATE OR REPLACE FUNCTION update_category_on_topic_change()
RETURNS TRIGGER AS
$$
BEGIN
    IF (TG_OP = 'INSERT') THEN
        UPDATE categories
            SET
                topics_count = (SELECT COUNT(*) FROM topics WHERE category_id = NEW.category_id)
            WHERE category_id = NEW.category_id;
        RETURN NEW;
    END IF;

    UPDATE categories
        SET
            topics_count = (SELECT COUNT(*) FROM topics WHERE category_id = OLD.category_id)
        WHERE category_id = OLD.category_id;
    RETURN OLD;

END
$$
LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS update_category_on_topic_change ON topics;
CREATE TRIGGER update_category_on_topic_change AFTER INSERT OR DELETE ON topics
    FOR EACH ROW EXECUTE PROCEDURE update_category_on_topic_change();




CREATE TABLE IF NOT EXISTS messages (
	message_id serial PRIMARY KEY,
	topic_id   integer NOT NULL REFERENCES topics(topic_id),
	author_id  integer NOT NULL REFERENCES users(user_id),
	content    text NOT NULL,
	created    timestamptz NOT NULL
);

CREATE INDEX messages_created_idx ON messages(created);

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


COMMIT;
