package mariadb

const ServiceDockerTemplate = `version: '2'
services:
  {{.Name}}:
    image: '{{.Image}}'
    container_name: '{{.Name}}'
    restart: always
    labels:
      - 'service.catalog={{.Catalog}}'
      - 'service.image={{.Image}}'
      - 'service.confPath={{.Volumes.ConfPath}}'
      - 'service.dataPath={{.Volumes.DataPath}}'
      - 'service.backPath={{.Volumes.BackPath}}'
      - 'service.root={{.Env.Root}}'
      - 'service.password={{.Env.Password}}'
      - 'service.host={{.Svc.Host}}'
      - 'service.port={{.Svc.Port}}'
    volumes:
      - {{.Volumes.ConfPath}}:/etc/mysql/conf.d/
      - {{.Volumes.DataPath}}:/var/lib/mysql
      - {{.Volumes.BackPath}}/data:/var/mariadb/backup
    environment:
      MYSQL_ROOT_PASSWORD: {{.Env.Password}}`

const JobDockerTemplate = `version: '2'
services:
  {{.Name}}Backup:
    image: '{{.Image}}'
    container_name: '{{.Name}}Backup'
    volumes:
      - {{.Volumes.ConfPath}}:/etc/mysql/conf.d/
      - {{.Volumes.DataPath}}:/var/lib/mysql
      - {{.Volumes.BackPath}}/data:/var/mariadb/backup
    environment:
      MYSQL_ROOT_PASSWORD: {{.Env.Password}}
    command: /usr/bin/sh -c {{.Command}}`

const BackupDockerTemplate = `version: '2'
services:
  {{.Name}}Backup:
    image: '{{.Image}}'
    container_name: '{{.Name}}Backup'
    volumes:
      - {{.Volumes.ConfPath}}:/etc/mysql/conf.d/
      - {{.Volumes.DataPath}}:/var/lib/mysql
      - {{.Volumes.BackPath}}/data:/var/mariadb/backup
    environment:
      MYSQL_ROOT_PASSWORD: {{.Env.Password}}
    command: mariabackup --defaults-file=/etc/mysql/conf.d/my.cnf --backup --host={{.Svc.Host}} --port={{.Svc.Port}} --target-dir=/var/mariadb/backup/ --datadir=/var/lib/mysql --user={{.Env.Root}} --password={{.Env.Password}}`

const PrepareDockerTemplate = `version: '2'
services:
  {{.Name}}Prepare:
    image: '{{.Image}}'
    container_name: '{{.Name}}Prepare'
    volumes:
      - {{.Volumes.ConfPath}}:/etc/mysql/conf.d/
      - {{.Volumes.DataPath}}:/var/lib/mysql
      - {{.Volumes.BackPath}}/data:/var/mariadb/backup
    environment:
      MYSQL_ROOT_PASSWORD: {{.Env.Password}}
    command: mariabackup --defaults-file=/etc/mysql/conf.d/my.cnf --prepare --host={{.Svc.Host}} --port={{.Svc.Port}} --target-dir=/var/mariadb/backup/ --datadir=/var/lib/mysql --user={{.Env.Root}} --password={{.Env.Password}}`

const RestoreDockerTemplate = `version: '2'
services:
  {{.Name}}Restore:
    image: '{{.Image}}'
    container_name: '{{.Name}}Restore'
    volumes:
      - {{.Volumes.ConfPath}}:/etc/mysql/conf.d/
      - {{.Volumes.DataPath}}:/var/lib/mysql
      - {{.Volumes.BackPath}}/data:/var/mariadb/backup
    environment:
      MYSQL_ROOT_PASSWORD: {{.Env.Password}}
    command: mariabackup --defaults-file=/etc/mysql/conf.d/my.cnf --copy-back --host={{.Svc.Host}} --port={{.Svc.Port}} --target-dir=/var/mariadb/backup/ --datadir=/var/lib/mysql --user={{.Env.Root}} --password={{.Env.Password}}`

const CleanDockerTemplate = `version: '2'
services:
  {{.Name}}Clean:
    image: '{{.Image}}'
    container_name: '{{.Name}}Clean'
    volumes:
      - {{.Volumes.ConfPath}}:/etc/mysql/conf.d/
      - {{.Volumes.DataPath}}:/var/lib/mysql
      - {{.Volumes.BackPath}}/data:/var/mariadb/backup
    environment:
      MYSQL_ROOT_PASSWORD: {{.Env.Password}}
    command: rm -rf /var/mariadb/backup/* && ls /var/mariadb/backup/`

const ChownDockerTemplate = `version: '2'
services:
  {{.Name}}Chown:
    image: '{{.Image}}'
    container_name: '{{.Name}}Chown'
    volumes:
      - {{.Volumes.ConfPath}}:/etc/mysql/conf.d/
      - {{.Volumes.DataPath}}:/var/lib/mysql
      - {{.Volumes.BackPath}}/data:/var/mariadb/backup
    environment:
      MYSQL_ROOT_PASSWORD: {{.Env.Password}}
    command: chown -R mysql:mysql /var/lib/mysql/*`
