CREATE TABLE IF NOT EXISTS songs (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL,
    artist_id INTEGER,
    album_id INTEGER NOT NULL,
    path TEXT NOT NULL,
    FOREIGN KEY (artist_id) REFERENCES artists(id)
    FOREIGN KEY (album_id) REFERENCES albums(id)
    UNIQUE(name, artist_id, album_id)
);
/sp
CREATE TABLE IF NOT EXISTS artists (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL,
    UNIQUE(name)
);
/sp
CREATE TABLE IF NOT EXISTS albums (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL,
    artist_id INTEGER,
    FOREIGN KEY (artist_id) REFERENCES artists(id)
    UNIQUE(name, artist_id)
);
/sp
CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL,
    login TEXT NOT NULL,
    password TEXT NOT NULL,
    is_admin INTEGER, 
    UNIQUE(login)
);
/sp
CREATE TABLE IF NOT EXISTS playlists (
    id INTEGER PRIMARY KEY,
    user_id INTEGER NOT NULL,
    name TEXT NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id)
);
/sp
CREATE TABLE IF NOT EXISTS playlist_items (
    id INTEGER PRIMARY KEY,
    song_id INTEGER NOT NULL,
    playlist_id INTEGER NOT NULL,
    FOREIGN KEY (song_id) REFERENCES songs(id),
    FOREIGN KEY (playlist_id) REFERENCES playlists(id),
    UNIQUE(song_id, playlist_id)
);
-- CREATE TRIGGER IF NOT EXISTS before_insert_song_artist
-- BEFORE INSERT ON songs
-- BEGIN
--     INSERT OR IGNORE INTO artists (name) VALUES (NEW.artist);
-- END;
-- /sp

-- CREATE TRIGGER IF NOT EXISTS before_insert_song_album
-- BEFORE INSERT ON songs
-- BEGIN
--     INSERT OR IGNORE INTO albums (name, artist) VALUES (NEW.album, NEW.artist);
-- END;
-- /sp
-- CREATE TRIGGER IF NOT EXISTS after_delete_song_artist
-- AFTER DELETE ON songs
-- BEGIN
--     DELETE FROM artists
--     WHERE name = OLD.artist
--     AND NOT EXISTS (
--         SELECT 1
--         FROM songs
--         WHERE artist = OLD.artist
--     );
-- END;
-- /sp
-- CREATE TRIGGER IF NOT EXISTS after_delete_song_album
-- AFTER DELETE ON songs
-- BEGIN
--     DELETE FROM album
--     WHERE name = OLD.a
--     AND NOT EXISTS (
--         SELECT 1
--         FROM songs
--         WHERE artist = OLD.artist
--     );
-- END;
