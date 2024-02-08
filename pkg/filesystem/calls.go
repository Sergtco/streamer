package filesystem

import (
	"fmt"
	"os"
)

/*
Returns pointer to os.File for given path.

If path is not correct, or file does not exist, returns an error.
*/
func FetchFile(path string) (*os.File, error) {
    file, err := os.Open(path)
    if err != nil {
        return nil, fmt.Errorf("Unable to retrieve file: %e", err)
    }
    return file, nil
}

/* 
Deletes file by given path.

If file does not exist, error is returned.
*/
func DeleteFile(path string) (error) {
    err := os.Remove(path)
    if err != nil {
        return fmt.Errorf("Unable to delete file: %e", err)
    }
    return nil
}

/*
Puts file by given path.

If file exists, os.ErrExists is returned.
*/
func PutFile(data []byte, path string) (error) {
    if _, err := os.Stat(path); err == nil {
        return os.ErrExist
    } else {
        err := os.WriteFile(path, data, 0644)
        if err != nil {
            return fmt.Errorf("Unable to write file: %e", err)
        }
    }
    return nil
}
