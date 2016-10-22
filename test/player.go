package test

type Player struct {
    Id   int    `db:"id" dblabel:"auto"`
    Uuid string `db:"uuid"`
    Name string `db:"name"`
} // player: Insert

