CREATE TABLE IF NOT EXISTS groups (
	id INTEGER PRIMARY KEY,
	name TEXT UNIQUE,
	faculty_id INTEGER,
	spreadsheet_id TEXT,
	admin_id INTEGER
);
CREATE TABLE IF NOT EXISTS lessons (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    group_id        INTEGER,
    subject         TEXT NOT NULL,
	lesson_type     TEXT NOT NULL, 
    subgroup_number INTEGER NOT NULL,
	date 			INTEGER NOT NULL,
	time     		INTEGER NOT NULL,
    FOREIGN KEY (group_id) REFERENCES groups(id)
);

CREATE INDEX IF NOT EXISTS lessons_group_id_idx ON lessons(group_id);

CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    tg_id INTEGER,
    group_id INTEGER,
    full_name TEXT NOT NULL, 
    CONSTRAINT tg_id_ak UNIQUE (tg_id),
    FOREIGN KEY (group_id) REFERENCES groups(id)
);

CREATE TABLE IF NOT EXISTS lessons_requests (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER, 
    lesson_id INTEGER,
    msg_id INTEGER,
    chat_id INTEGER NOT NULL,
    submit_time INTEGER NOT NULL,
    FOREIGN KEY (lesson_id) REFERENCES lessons(id),
    FOREIGN KEY (user_id) REFERENCES users(id)
);

CREATE INDEX IF NOT EXISTS lessons_requests_lesson_id_idx ON lessons_requests(lesson_id);
CREATE INDEX IF NOT EXISTS lessons_requests_user_id_idx ON lessons_requests(user_id);

CREATE TABLE IF NOT EXISTS users_roles (
    user_id INTEGER,
    role_name TEXT,
    FOREIGN KEY (user_id) REFERENCES users(id)
);

CREATE TABLE IF NOT EXISTS group_requests (
    uuid   TEXT,
    msg_id INTEGER,
    chat_id INTEGER
);

CREATE INDEX IF NOT EXISTS group_requests_msg_id_idx ON group_requests(msg_id);

CREATE TABLE IF NOT EXISTS admin_requests (
    uuid TEXT, 
    msg_id INTEGER,
    chat_id INTEGER
);

CREATE INDEX IF NOT EXISTS admin_requests_msg_id_idx ON admin_requests(msg_id);

CREATE TABLE IF NOT EXISTS states(
    chat_id INTEGER,
    state TEXT
);

CREATE UNIQUE INDEX IF NOT EXISTS states_chat_id_idx ON states(chat_id);

CREATE TABLE IF NOT EXISTS info(
    chat_id INTEGER,
    json TEXT
);

CREATE UNIQUE INDEX IF NOT EXISTS info_chat_id_idx ON info(chat_id);