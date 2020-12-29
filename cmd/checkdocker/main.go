package main

import (
	"flag"
	"fmt"
	"os"
	"path"
	"strings"

	utils "github.com/yametech/cloud-native-tools/pkg/utils"
)

func main() {
	var url string
	var codeType string
	var projectPath string

	flag.StringVar(&url, "url", "./", "-url ./")
	flag.StringVar(&codeType, "codetype", "java-maven", "-codetype java-maven")
	flag.StringVar(&projectPath, "path", "", "-path subdirectory")
	flag.Parse()

	fmt.Printf("url=%v  codeType=%v\n", url, codeType)
	err := CheckDockerFile(url, codeType, projectPath)
	if err != nil {
		panic(err)
	}
}

func CheckDockerFile(url string, codeType string, projectPath string) error {
	url = path.Join(url, "Dockerfile")
	switch codeType {
	case "django":
		err := djangoDocker(url)
		if err != nil {
			return err
		}
	case "java-maven":
		err := javaDocker(url, projectPath)
		if err != nil {
			return err
		}
	case "easyswoole":
		err := easyswooleDocker(url)
		if err != nil {
			return err
		}
	case "web":
		err := webDocker(url)
		if err != nil {
			return err
		}
	}

	return nil
}

func webDocker(url string) error {
	const ngContent = `
# For more information on configuration, see:
#   * Official English Documentation: http://nginx.org/en/docs/
#   * Official Russian Documentation: http://nginx.org/ru/docs/

user nginx;
worker_processes auto;
error_log /var/log/nginx/error.log;
pid /run/nginx.pid;

# Load dynamic modules. See /usr/share/doc/nginx/README.dynamic.
include /usr/share/nginx/modules/*.conf;

events {
    worker_connections 1024;
}

http {
    log_format  main  '$remote_addr - $remote_user [$time_local] "$request" '
                      '$status $body_bytes_sent "$http_referer" '
                      '"$http_user_agent" "$http_x_forwarded_for"';

    access_log  /var/log/nginx/access.log  main;

    sendfile            on;
    tcp_nopush          on;
    tcp_nodelay         on;
    keepalive_timeout   65;
    types_hash_max_size 2048;

    include             /etc/nginx/mime.types;
    default_type        application/octet-stream;

    # Load modular configuration files from the /etc/nginx/conf.d directory.
    # See http://nginx.org/en/docs/ngx_core_module.html#include
    # for more information.

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
        
        location ~ \.(html|ico)$ {
            root /usr/share/nginx/html;
        }

        location /static {
            root /usr/share/nginx/html;
        }

        error_page 404 /404.html;
        location = /404.html {
        }

        error_page 500 502 503 504 /50x.html;
        location = /50x.html {
        }

        location / {
                proxy_pass http://backend;
                proxy_redirect off;
                proxy_set_header Host $host;
                proxy_set_header X-Real-IP $remote_addr;
                proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
                client_max_body_size 10m;
                client_body_buffer_size 128k;
                proxy_connect_timeout 90s;
                proxy_send_timeout 90s;
                proxy_read_timeout 90s;
                proxy_buffer_size 4k;
                proxy_buffers 4 32k;
                proxy_busy_buffers_size 64k;
                proxy_temp_file_write_size 64k;
        }
    
    }
    
    upstream backend {
    	server {{.UPSTREAM}} weight=2 max_fails=3 fail_timeout=3s;
    }

}`
	const content = `
FROM harbor.ym/devops/node:14.3.0 AS builder
WORKDIR /app 

RUN npm install -g cnpm --registry=https://registry.npm.taobao.org
RUN alias cnpm="npm --registry=https://registry.npm.taobao.org \
    --cache=$HOME/.npm/.cache/cnpm \
    --disturl=https://npm.taobao.org/dist \
    --userconfig=$HOME/.cnpmrc"
RUN cnpm install -g webpack

# RUN curl -o- -L https://yarnpkg.com/install.sh | bash
RUN cnpm install yarn -g
RUN  yarn config set registry https://registry.npm.taobao.org \
    && yarn config set sass-binary-site http://npm.taobao.org/mirrors/node-sass

# Install package cache
COPY package.json .
COPY yarn.lock .
RUN yarn install

# Building
COPY . .
RUN yarn run build

FROM harbor.ym/devops/nginx:1.19.0
COPY --from=builder app/dist /usr/share/nginx/html/
COPY --from=builder app/nginx.conf /etc/nginx/nginx.conf

EXPOSE 80
`
	if _, err := os.Stat(url); os.IsNotExist(err) {
		err = utils.GenerateFile(url, content)
		if err != nil {
			return err
		}
	}
	ngFile := strings.Replace(url, "Dockerfile", "nginx.conf", -1)
	if _, err := os.Stat(ngFile); os.IsNotExist(err) {
		err = utils.GenerateFile(ngFile, ngContent)
		if err != nil {
			return err
		}
	}
	return nil
}

func djangoDocker(filename string) error {
	const content = `
FROM harbor.ym/devops/ar:1.1
WORKDIR /workshop
COPY . /workshop/
RUN pip install -r requirements.txt -i 'https://mirrors.aliyun.com/pypi/simple/'
EXPOSE 8000
CMD python manage.py runserver 0.0.0.0:8000`
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		err = utils.GenerateFile(filename, content)
		if err != nil {
			return err
		}
	}
	return nil
}

func javaDocker(filename string, projectPath string) error {
	type Param struct {
		ProjectPath string
	}

	param := &Param{ProjectPath: "*"}
	if len(strings.Trim(projectPath, "")) > 0 {
		param.ProjectPath = projectPath
	}

	var javaMavenDockerfileContentTpl = `
# Step : Test and package
FROM harbor.ym/devops/maven363:latest as builder
WORKDIR /build
ADD settings.xml /root/.m2/settings.xml
ADD pom.xml pom.xml
RUN mvn verify clean --fail-never

# add code build
ADD . .
RUN mvn install

# # Step : Package image
FROM harbor.ym/devops/openjdk8:latest
COPY --from=builder /build/{{.ProjectPath}}/target/* /app/
ENTRYPOINT ["/app/bin/run.sh"]
`

	content, err := Render(param, javaMavenDockerfileContentTpl)
	if err != nil {
		panic(err)
	}

	const xmlContent = `
<?xml version="1.0" encoding="UTF-8"?>

<!--
Licensed to the Apache Software Foundation (ASF) under one
or more contributor license agreements.  See the NOTICE file
distributed with this work for additional information
regarding copyright ownership.  The ASF licenses this file
to you under the Apache License, Version 2.0 (the
"License"); you may not use this file except in compliance
with the License.  You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing,
software distributed under the License is distributed on an
"AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
KIND, either express or implied.  See the License for the
specific language governing permissions and limitations
under the License.
-->

<!--
 | This is the configuration file for Maven. It can be specified at two levels:
 |
 |  1. User Level. This settings.xml file provides configuration for a single user,
 |                 and is normally provided in ${user.home}/.m2/settings.xml.
 |
 |                 NOTE: This location can be overridden with the CLI option:
 |
 |                 -s /path/to/user/settings.xml
 |
 |  2. Global Level. This settings.xml file provides configuration for all Maven
 |                 users on a machine (assuming they're all using the same Maven
 |                 installation). It's normally provided in
 |                 ${maven.home}/conf/settings.xml.
 |
 |                 NOTE: This location can be overridden with the CLI option:
 |
 |                 -gs /path/to/global/settings.xml
 |
 | The sections in this sample file are intended to give you a running start at
 | getting the most out of your Maven installation. Where appropriate, the default
 | values (values used when the setting is not specified) are provided.
 |
 |-->
<settings xmlns="http://maven.apache.org/SETTINGS/1.0.0"
          xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
          xsi:schemaLocation="http://maven.apache.org/SETTINGS/1.0.0 http://maven.apache.org/xsd/settings-1.0.0.xsd">
  <!-- localRepository
   | The path to the local repository maven will use to store artifacts.
   |
   | Default: ${user.home}/.m2/repository
  <localRepository>/path/to/local/repo</localRepository>
  -->
  <localRepository>/Users/xiaopengdeng/java/mavenrepo</localRepository>

  <!-- interactiveMode
   | This will determine whether maven prompts you when it needs input. If set to false,
   | maven will use a sensible default value, perhaps based on some other setting, for
   | the parameter in question.
   |
   | Default: true
  <interactiveMode>true</interactiveMode>
  -->

  <!-- offline
   | Determines whether maven should attempt to connect to the network when executing a build.
   | This will have an effect on artifact downloads, artifact deployment, and others.
   |
   | Default: false
  <offline>false</offline>
  -->

  <!-- pluginGroups
   | This is a list of additional group identifiers that will be searched when resolving plugins by their prefix, i.e.
   | when invoking a command line like "mvn prefix:goal". Maven will automatically add the group identifiers
   | "org.apache.maven.plugins" and "org.codehaus.mojo" if these are not already contained in the list.
   |-->
  <pluginGroups>
    <!-- pluginGroup
     | Specifies a further group identifier to use for plugin lookup.
    <pluginGroup>com.your.plugins</pluginGroup>
    -->
	<pluginGroup>org.sonarsource.scanner.maven</pluginGroup>
  </pluginGroups>

  <!-- proxies
   | This is a list of proxies which can be used on this machine to connect to the network.
   | Unless otherwise specified (by system property or command-line switch), the first proxy
   | specification in this list marked as active will be used.
   |-->
  <proxies>
    <!-- proxy
     | Specification for one proxy, to be used in connecting to the network.
     |
    <proxy>
      <id>optional</id>
      <active>true</active>
      <protocol>http</protocol>
      <username>proxyuser</username>
      <password>proxypass</password>
      <host>proxy.host.net</host>
      <port>80</port>
      <nonProxyHosts>local.net|some.host.com</nonProxyHosts>
    </proxy>
    -->
  </proxies>

  <!-- servers
   | This is a list of authentication profiles, keyed by the server-id used within the system.
   | Authentication profiles can be used whenever maven must make a connection to a remote server.
   |-->
  <servers>
    <!-- server
     | Specifies the authentication information to use when connecting to a particular server, identified by
     | a unique name within the system (referred to by the 'id' attribute below).
     |
     | NOTE: You should either specify username/password OR privateKey/passphrase, since these pairings are
     |       used together.
     |
    <server>
      <id>deploymentRepo</id>
      <username>repouser</username>
      <password>repopwd</password>
    </server>
    -->

    <!-- Another sample, using keys to authenticate.
    <server>
      <id>siteServer</id>
      <privateKey>/path/to/private/key</privateKey>
      <passphrase>optional; leave empty if not used.</passphrase>
    </server>
    -->
	<server>
      <id>releases</id>
      <username>yame_deploy</username>
      <password>yame_deploy666</password>
    </server>
	<server>
      <id>snapshots</id>
      <username>yame_deploy</username>
      <password>yame_deploy666</password>
    </server>
  </servers>

  <!-- mirrors
   | This is a list of mirrors to be used in downloading artifacts from remote repositories.
   |
   | It works like this: a POM may declare a repository to use in resolving certain artifacts.
   | However, this repository may have problems with heavy traffic at times, so people have mirrored
   | it to several places.
   |
   | That repository definition will have a unique id, so we can create a mirror reference for that
   | repository, to be used as an alternate download site. The mirror site will be the preferred
   | server for that repository.
   |-->
  <mirrors>
    <!-- mirror
     | Specifies a repository mirror site to use instead of a given repository. The repository that
     | this mirror serves has an ID that matches the mirrorOf element of this mirror. IDs are used
     | for inheritance and direct lookup purposes, and must be unique across the set of mirrors.
     |
    <mirror>
      <id>mirrorId</id>
      <mirrorOf>repositoryId</mirrorOf>
      <name>Human Readable Name for this Mirror.</name>
      <url>http://my.repository.com/repo/path</url>
    </mirror>
     -->
	<mirror>
      <id>nexus-yame</id>
      <mirrorOf>*</mirrorOf>
      <name>Nexus Yame</name>
      <url>http://repo.hq.in.ecpark.cn:8081/nexus/content/groups/public</url>
    </mirror>

	<mirror>
		<id>aliyunmaven</id>
		<mirrorOf>central</mirrorOf>
		<name>阿里云公共仓库</name>
		<url>https://maven.aliyun.com/repository/public</url>
	</mirror>
  </mirrors>

  <!-- profiles
   | This is a list of profiles which can be activated in a variety of ways, and which can modify
   | the build process. Profiles provided in the settings.xml are intended to provide local machine-
   | specific paths and repository locations which allow the build to work in the local environment.
   |
   | For example, if you have an integration testing plugin - like cactus - that needs to know where
   | your Tomcat instance is installed, you can provide a variable here such that the variable is
   | dereferenced during the build process to configure the cactus plugin.
   |
   | As noted above, profiles can be activated in a variety of ways. One way - the activeProfiles
   | section of this document (settings.xml) - will be discussed later. Another way essentially
   | relies on the detection of a system property, either matching a particular value for the property,
   | or merely testing its existence. Profiles can also be activated by JDK version prefix, where a
   | value of '1.4' might activate a profile when the build is executed on a JDK version of '1.4.2_07'.
   | Finally, the list of active profiles can be specified directly from the command line.
   |
   | NOTE: For profiles defined in the settings.xml, you are restricted to specifying only artifact
   |       repositories, plugin repositories, and free-form properties to be used as configuration
   |       variables for plugins in the POM.
   |
   |-->
  <profiles>
    <!-- profile
     | Specifies a set of introductions to the build process, to be activated using one or more of the
     | mechanisms described above. For inheritance purposes, and to activate profiles via <activatedProfiles/>
     | or the command line, profiles have to have an ID that is unique.
     |
     | An encouraged best practice for profile identification is to use a consistent naming convention
     | for profiles, such as 'env-dev', 'env-test', 'env-production', 'user-jdcasey', 'user-brett', etc.
     | This will make it more intuitive to understand what the set of introduced profiles is attempting
     | to accomplish, particularly when you only have a list of profile id's for debug.
     |
     | This profile example uses the JDK version to trigger activation, and provides a JDK-specific repo.
    <profile>
      <id>jdk-1.4</id>

      <activation>
        <jdk>1.4</jdk>
      </activation>

      <repositories>
        <repository>
          <id>jdk14</id>
          <name>Repository for JDK 1.4 builds</name>
          <url>http://www.myhost.com/maven/jdk14</url>
          <layout>default</layout>
          <snapshotPolicy>always</snapshotPolicy>
        </repository>
      </repositories>
    </profile>
    -->

    <!--
     | Here is another profile, activated by the system property 'target-env' with a value of 'dev',
     | which provides a specific path to the Tomcat instance. To use this, your plugin configuration
     | might hypothetically look like:
     |
     | ...
     | <plugin>
     |   <groupId>org.myco.myplugins</groupId>
     |   <artifactId>myplugin</artifactId>
     |
     |   <configuration>
     |     <tomcatLocation>${tomcatPath}</tomcatLocation>
     |   </configuration>
     | </plugin>
     | ...
     |
     | NOTE: If you just wanted to inject this configuration whenever someone set 'target-env' to
     |       anything, you could just leave off the <value/> inside the activation-property.
     |
    <profile>
      <id>env-dev</id>

      <activation>
        <property>
          <name>target-env</name>
          <value>dev</value>
        </property>
      </activation>

      <properties>
        <tomcatPath>/path/to/tomcat/instance</tomcatPath>
      </properties>
    </profile>
    -->

    <profile>
		<activation><activeByDefault>true</activeByDefault></activation>
		<repositories>
			<repository>
				<id>nexus</id>
				<url>http://repo.hq.in.ecpark.cn:8081/nexus/content/groups/public</url>
				<releases>
					<enabled>true</enabled>
				</releases>
				<snapshots>
					<enabled>true</enabled>
					<updatePolicy>always</updatePolicy>
				</snapshots>
			</repository>
		</repositories>
		<pluginRepositories>
			<pluginRepository>
				<id>nexus</id>
				<url>http://repo.hq.in.ecpark.cn:8081/nexus/content/groups/public</url>
				<releases>
					<enabled>true</enabled>
				</releases>
				<snapshots>
					<enabled>true</enabled>
					<updatePolicy>always</updatePolicy>
				</snapshots>
			</pluginRepository>
		</pluginRepositories>
	</profile>


	<profile>
		<id>sonar</id>
		<activation>
			<activeByDefault>true</activeByDefault>
		</activation>
		<properties>
			<sonar.host.url>
			  http://sonar.ym
			</sonar.host.url>
		</properties>
    </profile>


  </profiles>

  <!-- activeProfiles
   | List of profiles that are active for all builds.
   |
  <activeProfiles>
    <activeProfile>alwaysActiveProfile</activeProfile>
    <activeProfile>anotherAlwaysActiveProfile</activeProfile>
  </activeProfiles>
  -->
</settings>`

	if _, err := os.Stat(filename); os.IsNotExist(err) {
		err = utils.GenerateFile(filename, content)
		if err != nil {
			return err
		}
	}
	// if settings.xml not exists, create it too
	xmlFile := strings.Replace(filename, "Dockerfile", "settings.xml", -1)
	if _, err := os.Stat(xmlFile); os.IsNotExist(err) {
		err = utils.GenerateFile(xmlFile, xmlContent)
		if err != nil {
			return err
		}
	}
	return nil
}

func easyswooleDocker(filename string) error {
	const content = `
FROM centos:8

#version defined
ENV SWOOLE_VERSION 4.4.17
ENV EASYSWOOLE_VERSION 3.x-dev

#install libs
RUN yum install -y curl zip unzip  wget openssl-devel gcc-c++ make autoconf
#install php
RUN yum install -y php-devel php-openssl php-mbstring php-json
# swoole ext
RUN wget https://github.com/swoole/swoole-src/archive/v${SWOOLE_VERSION}.tar.gz -O swoole.tar.gz \
	&& mkdir -p swoole \
	&& tar -xf swoole.tar.gz -C swoole --strip-components=1 \
	&& rm swoole.tar.gz \
	&& ( \
	cd swoole \
	&& phpize \
	&& ./configure --enable-openssl \
	&& make \
	&& make install \
	) \
	&& sed -i "2i extension=swoole.so" /etc/php.ini \
	&& rm -r swoole

# Dir
WORKDIR /easyswoole
# install easyswoole
COPY . /easyswoole
# composer
RUN curl -sS https://getcomposer.org/installer | php \
   && mv composer.phar /usr/bin/composer
# use aliyun composer
RUN composer config -g repo.packagist composer https://mirrors.aliyun.com/composer/
#install app ext by composer.json
RUN composer install
#此端口需根据项目配置文件里的端口做相应开放
EXPOSE 9501
# RUN APP
CMD php easyswoole start
`
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		err = utils.GenerateFile(filename, content)
		if err != nil {
			return err
		}
	}
	return nil
}
