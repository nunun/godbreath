package test

// user: Insert
type User struct {
    Id   int    `db:"id" dblabel:"auto"`
    Uuid string `db:"uuid"`
    Name string `db:"name"`
}

