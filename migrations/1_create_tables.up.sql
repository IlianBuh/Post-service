-- BEGIN;


CREATE TABLE IF NOT EXISTS posts (
    post_id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id INT NOT NULL,
    header TEXT NOT NULL,
    content TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS themes (
    theme_id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    theme_name TEXT
);

CREATE TABLE IF NOT EXISTS post_theme(
    post_id INT REFERENCES posts(post_id) ON DELETE CASCADE,
    theme_id INT REFERENCES themes(theme_id) ON DELETE CASCADE,
    PRIMARY KEY (post_id, theme_id)
);

CREATE TABLE IF NOT EXISTS events(
    event_id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    "uid" TEXT UNIQUE NOT NULL, 
    "type" TEXT NOT NULL CHECK ("type" IN ('created')),
    payload TEXT NOT NULL,
    "status" TEXT NOT NULL DEFAULT 'undone' CHECK ("status" IN ('done', 'undone')),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    reserved_to TIMESTAMPTZ DEFAULT NOW() NOT NULL
);

-- COMMIT;