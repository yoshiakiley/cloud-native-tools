package main

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

const nginx = `
http {
    server {
        listen       80 default_server;
        listen       [::]:80 default_server;
        server_name  _;
        root         /usr/share/nginx/html;

        # Load configuration files for the default server block.

        location = / {
            root /usr/share/nginx/html;
            index index.html index.htm;
        }

    upstream backend {
    	server {{.UPSTREAM}} weight=2 max_fails=3 fail_timeout=3s;
    }
}
`

const expect_nginx = `
http {
    server {
        listen       80 default_server;
        listen       [::]:80 default_server;
        server_name  _;
        root         /usr/share/nginx/html;

        # Load configuration files for the default server block.

        location = / {
            root /usr/share/nginx/html;
            index index.html index.htm;
        }

    upstream backend {
    	server {{.UPSTREAM}} weight=2 max_fails=3 fail_timeout=3s;
    }
}
`

func Test_renderNginxConfig(t *testing.T) {
	if err := os.Setenv("UPSTREAM", "gateway.service.ym:2000"); err != nil {
		t.Fatal(err)
	}
	dir, err := ioutil.TempDir(".", "tmp")
	if err != nil {
		t.Fatal(err)
	}
	tempConf := strings.Join([]string{dir, "nginx.conf"}, "/")
	if err := ioutil.WriteFile(tempConf, []byte(nginx), 0777); err != nil {
		t.Fatal(err)
	}
	if err := render(tempConf); err != nil {
		t.Fatal(err)
	}
	if strings.Compare(expect_nginx, nginx) > 0 {
		t.Fatal("compare not expected")
	}
}
