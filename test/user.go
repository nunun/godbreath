package test

type User struct {
    Id   int    `db:"id" auto:"true"`
    Uuid string `db:"uuid"`
    Name string `db:"name"`
} // user: Insert

