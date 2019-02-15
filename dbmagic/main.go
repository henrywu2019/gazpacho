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
    name string
    file *os.File
}

// implements sql.Driver interface
// Open - 
func (d *DummyDriver) Open(name string) (driver.Conn, error) {
    file, err := os.Open(name)
    if err != nil {
        return nil, err
    }
    return &DummyConn{name, file}, nil
}

// implements sql.Conn interface
// Prepare -
func (c *DummyConn) Prepare(query string) (driver.Stmt, error) {
    return nil, fmt.Errorf("Prepare method not implemented")
}

// Close -
func (c *DummyConn) Close() error {
    return c.file.Close()
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

    r := csv.NewReader(c.file)
    r.FieldsPerRecord = 0       // enforce the same number of columns
    columns, err := r.Read()  // assume first line gives you column names
    if err != nil {
        return nil, err
    }
    res := &results{r, columns}
    return res, nil

}

type results struct {
    reader *csv.Reader
    columns []string
}

// driver.Rows interface
func (r *results) Columns() []string {
    return r.columns
}

func (r *results) Close() error {
    return nil
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

    rows, err := db.Query("SELECT * FROM csv")
    if err != nil {
        panic(err)
    }

    for rows.Next() {
        var f1, f2, f3 string
        err := rows.Scan(&f1, &f2, &f3)
        if err != nil {
            panic(err)
        }
        fmt.Printf("first_name=%s, last_name=%s, username=%s\n", f1, f2, f3)
    }
}
