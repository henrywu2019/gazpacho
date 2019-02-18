// dbmagic - implements database/sql interface on top of csv in 100 lines of Go
// code under MIT license, https://opensource.org/licenses/MIT

package main

import (
    "fmt"
    "os"
    "database/sql"
    "database/sql/driver"
    "encoding/csv"
    "io/ioutil"
)

type DummyDriver struct {
}

type DummyConn struct {
    name string         // file name
}

// implements sql.Driver interface
// Open - 
func (d *DummyDriver) Open(name string) (driver.Conn, error) {
    file, err := os.Open(name)
    if err != nil {
        return nil, err
    }
    defer file.Close()
    return &DummyConn{name: name}, nil
}

// implements sql.Conn interface
// Prepare -
func (c *DummyConn) Prepare(query string) (driver.Stmt, error) {
    return nil, fmt.Errorf("Prepare method not implemented")
}

// Close -
func (c *DummyConn) Close() error {
    return nil
}

// Begin - 
func (c *DummyConn) Begin() (_ driver.Tx, err error) {
    return c, fmt.Errorf("Begin method not implemented")
}

// Tx interface
func (c *DummyConn) Commit() error {
    return fmt.Errorf("Commit method not implemented")
}
func (c *DummyConn) Rollback() error {
    return fmt.Errorf("Rollback method not implemented")
}

// Queryer interface
// Note: this is DEPRECATED, QueryerContext is new guy
func (c *DummyConn) Query(query string, args []driver.Value) (driver.Rows, error) {
    if query != "SELECT * FROM csv" {
        return nil, fmt.Errorf("Only `SELECT * FROM csv` string is implemented!")
    }

    file, err := os.Open(c.name)
    if err != nil {
        return nil, err
    }

    r := csv.NewReader(file)
    r.FieldsPerRecord = 0       // enforce the same number of columns
    columns, err := r.Read()  // assume first line gives you column names
    if err != nil {
        return nil, err
    }
    res := &results{file, r, columns}
    return res, nil
}

type results struct {
    file *os.File
    reader *csv.Reader
    columns []string
}

// driver.Rows interface
func (r *results) Columns() []string {
    return r.columns
}

func (r *results) Close() error {
    return r.file.Close()
}

func (r *results) Next(dest []driver.Value) error {
    d, err := r.reader.Read()
    if err != nil {
        return err
    }
    for i := 0; i != len(r.columns); i++ {
        dest[i] = driver.Value(d[i])
    }
    return nil
}


func main() {

    in := []byte(`first_name,last_name,username
"Rob","Pike",rob
Ken,Thompson,ken
"Robert","Griesemer","gri"`)
    err := ioutil.WriteFile("go.csv", in, 0666)
    if err != nil {
        panic(err)
    }

    // manually register dummy driver
    sql.Register("dummy", &DummyDriver{})

    // use dummy driver via database/sql interfaces
    db, err := sql.Open("dummy", "go.csv")
    if err != nil {
        panic(err)
    }

    rows := make([]*sql.Rows, 2)
    for i := 0; i != 2; i++ {
        rows[i], err = db.Query("SELECT * FROM csv")
        if err != nil {
            panic(err)
        }
    }

    for _, r := range rows {
        for r.Next() {
            var f1, f2, f3 string
            err := r.Scan(&f1, &f2, &f3)
            if err != nil {
                panic(err)
            }
            fmt.Printf("first_name=%s, last_name=%s, username=%s\n", f1, f2, f3)
        }
    }
}
