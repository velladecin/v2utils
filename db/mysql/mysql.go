package mysql

import (
    "fmt"
    v2 "vella/v2utils"
    "database/sql"
    _ "github.com/go-sql-driver/mysql-master"
    "reflect"
    "strings"
    "regexp"
)

const (
    // Connection defaults
    myhost = "localhost"
    myport = 3306
    myproto = "tcp" // nomodify
    mytimeout = 3

    // Other defaults
    zero64 int64 = 0
)

type myConf struct {
    uname, passwd, host, db string
    port, timeout int
}

func (mc *myConf) DriverString() string {
    ds := fmt.Sprintf("%s:%s@%s(%s:%d)/%s?parseTime=true", mc.uname, mc.passwd, myproto, mc.host, mc.port, mc.db)

    if mc.timeout > 0 {
        ds = fmt.Sprintf("%s&timeout=%ds", ds, mc.timeout)
    }

    return ds
}

type MysqlModifier func(mc *myConf)

type My struct {
    conf *myConf
    dbh *sql.DB
}

func Connect(uname, db string, mod ...MysqlModifier) *My {
    mc := &myConf{
        uname: uname,
        db: db,
        host: myhost,
        port: myport,
        timeout: mytimeout,
    }

    for _, m := range mod {
        m(mc)
    }

    dbh, err := sql.Open("mysql", mc.DriverString())
    if err != nil {
        panic(err.Error())
    }

    return &My{
        conf: mc,
        dbh: dbh,
    }
}

func (m *My) Close() {
    m.dbh.Close()
}

func (m *My) Alive() bool {
    err := m.dbh.Ping()
    if err != nil {
        return false
    }

    return true
}

type DbColumn struct {
    Name string
    Gotype reflect.Type
    Dbtype string
    Value interface{}
}

type Rowset [][]DbColumn

func (m *My) Select(q string, v ...interface{}) (Rowset, error) {
    rowset := make(Rowset, 0)

    rows, err := m.dbh.Query(q, v...)
    if err != nil {
        return rowset, err
    }
    defer rows.Close()

    ct, err := rows.ColumnTypes()
    if err != nil {
        return rowset, err
    }

    cols, err := rows.Columns()
    if err != nil {
        return rowset, err
    }

    values := make([]interface{}, len(ct))
    for i, t := range ct {
        values[i] = reflect.New(t.ScanType()).Interface()
    }

    for rows.Next() {
        if err := rows.Scan(values...); err != nil {
            return rowset, err
        } 

        row := make([]DbColumn, len(values))
        for i:=0; i<len(values); i++ {
            gotype := reflect.TypeOf(values[i])
            value := reflect.ValueOf(values[i]).Elem().Interface()
            dbtype := ct[i].DatabaseTypeName()
            //scanType := ct[i].ScanType()

            row[i] = DbColumn{cols[i], gotype, dbtype, value}
        }

        rowset = append(rowset, row)
    }

    return rowset, nil
}

func (m *My) Exec(sql string, v ...interface{}) (int64, int64, error) {
    res, err := m.dbh.Exec(sql, v...)
    if err != nil {
        return zero64, zero64, err
    }

    rowsAffected, err := res.RowsAffected()
    if err != nil {
        return zero64, zero64, err
    }

    lastinsertid, err := res.LastInsertId()
    if err != nil {
        return zero64, zero64, err
    }

    return rowsAffected, lastinsertid, nil
}

func (m *My) InsertOnce(sql string) (int64, int64, error) { return m.Exec(sql) }
func (m *My) Insert(table string, cols []string, vals ...[]interface{}) (int64, int64, error) {
    var values []string
    for row:=0; row<len(vals); row++ {
        var rowval []string
        for col:=0; col<len(vals[row]); col++ {
            switch v := reflect.ValueOf(vals[row][col]); v.Kind() {
            case reflect.String:
                s := v.String()
                if ok, _ := regexp.MatchString(`\(\)$`, s); ok {
                    rowval = append(rowval, s)
                } else {
                    rowval = append(rowval, fmt.Sprintf(`"%s"`, s))
                }
            case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
                rowval = append(rowval, fmt.Sprintf("%d", v))
            case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
                rowval = append(rowval, fmt.Sprintf("%d", v))
            case reflect.Float32, reflect.Float64:
                rowval = append(rowval, fmt.Sprintf("%f", v))
            default:
                return zero64, zero64, &UnhandledDbTypeErr{v.Kind().String()}            
            }
        }

        values = append(values, strings.Join(rowval, ","))
    }

    v := strings.Join(values, "),(")

    var q string
    if len(cols) == 0 {
        q = fmt.Sprintf("INSERT INTO %s VALUES(%s)", table, v)
    } else {
        q = fmt.Sprintf("INSERT INTO %s(%s) VALUES(%s)", table, strings.Join(cols, ","), v)
    }
    //fmt.Println(q)

    return m.InsertOnce(q)
}


//
// Modifiers

func MysqlPasswd(passwd string) MysqlModifier {
    return func(m *myConf) {
        m.passwd = passwd
    }
}

func MysqlHost(host string) MysqlModifier {
    return func(m *myConf) {
        m.host = host
    }
}

func MysqlPort(port int) MysqlModifier {
    if err := v2.PortRange(port); err != nil {
        panic(err.Error())
    }

    return func(m *myConf) {
        m.port = port
    }
}

func MysqlTimeout(timeout int) MysqlModifier {
    if timeout < 1 || timeout > 30 { // seconds
        panic(fmt.Sprintf("Allowed timeout(1-30), got(%d) seconds", timeout))
    }

    return func(m *myConf) {
        m.timeout = timeout
    }
}
