CREATE TABLE IF NOT EXISTS groups (
	id INTEGER PRIMARY KEY,
	name TEXT UNIQUE,
	faculty_id INTEGER,
	spreadsheet_id TEXT,
	admin_id INTEGER
)
CREATE TABLE IF NOT EXISTS lessons (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    group_id        INTEGER,
    subject         TEXT NOT NULL,
	lesson_type     TEXT NOT NULL, 
    subgroup_number INTEGER NOT NULL,
	date 			INTEGER NOT NULL,
	time     		INTEGER NOT NULL
    FOREIGN KEY (group_id) REFERENCES groups(id)
)
CREATE TABLE IF NOT EXISTS lessons_weeks (
    lesson_id INTEGER,
    week_number INTEGER NOT NULL,
    FOREIGN KEY (lesson_id) REFERENCES lessons(id),
)
CREATE INDEX IF NOT EXISTS lessons_weeks_id_idx ON lessons_weeks (lesson_id);
