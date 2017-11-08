package config

import (
	"io/ioutil"
	"os"
	"testing"
)

var configContent = `
[mikrotik]
workers = 8
insecure = true
username = test
password = testpassword
[database]
type     = driver
host     = 127.0.0.1
user     = testuser
password = testpassword
database = testdatabase
`

func TestReadConfig(t *testing.T) {
	tmpfd, err := ioutil.TempFile("", "configtest")
	if err != nil {
		t.Fatalf("Can't create test config file. Err: %s", err)
	}

	defer os.Remove(tmpfd.Name())

	if _, err := tmpfd.Write([]byte(configContent)); err != nil {
		t.Fatalf("Can't write test init settings to file. Err: %s", err)
	}

	c, err := New(tmpfd.Name())
	if err != nil {
		t.Fatalf("ReadConfig return error. Err: %s", err)
	}

	if c.Database.Host != "127.0.0.1" {
		t.Fatalf("Config read error. Database host is %s should be 127.0.0.1.", c.Database.Host)
	}
	if c.Database.Type != "driver" {
		t.Fatalf("Config read error. Database type is %s should be driver", c.Database.Type)
	}
	if c.Database.User != "testuser" {
		t.Fatalf("Config read error. Database user is %s should be testuser", c.Database.User)
	}
	if c.Database.Password != "testpassword" {
		t.Fatalf("Config read error. Database user is %s should be testpassword", c.Database.Password)
	}
	if c.Database.Name != "testdatabase" {
		t.Fatalf("Config read error. Database user is %s should be testdatabase", c.Database.Name)
	}

	if c.Mikrotik.Workers != 8 {
		t.Fatalf("Config read error. Mikrotik workers is %d should be 8", c.Mikrotik.Workers)
	}
	if c.Mikrotik.Insecure != true {
		t.Fatalf("Config read error. Mikrotik insecure is %b should be true", c.Mikrotik.Insecure)
	}
	if c.Mikrotik.Username != "test" {
		t.Fatalf("Config read error. Mikrotik user is %s should be test", c.Mikrotik.Username)
	}
	if c.Mikrotik.Password != "testpassword" {
		t.Fatalf("Config read error. Mikrotik password is %s should be testpassword", c.Mikrotik.Password)
	}
}
