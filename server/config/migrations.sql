CREATE TABLE IF NOT EXISTS songs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    artist TEXT,
    album TEXT NOT NULL,
    path TEXT NOT NULL
);
жопа
CREATE TABLE IF NOT EXISTS artists (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    UNIQUE(name)
);
жопа
CREATE TABLE IF NOT EXISTS albums (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    artist TEXT,
    UNIQUE(name)

);
жопа

CREATE TRIGGER IF NOT EXISTS after_insert_song_artist
AFTER INSERT ON songs
BEGIN
    INSERT OR IGNORE INTO artists (name) VALUES (NEW.artist);
END;
жопа

CREATE TRIGGER IF NOT EXISTS after_insert_song_album
AFTER INSERT ON songs
BEGIN
    INSERT OR IGNORE INTO albums (name, artist) VALUES (NEW.album, NEW.artist);
END;
жопа

CREATE TRIGGER IF NOT EXISTS after_delete_song_artist
AFTER DELETE ON songs
BEGIN
    DELETE FROM artists
    WHERE name = OLD.artist
    AND NOT EXISTS (
        SELECT 1
        FROM songs
        WHERE artist = OLD.artist
    );
END;
