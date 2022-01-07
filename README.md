# neos-thicc-master
A dummy thicc master server implementation - as opposed to Rock's minimaster

[![Dummy Thicc Chungus](https://repository-images.githubusercontent.com/438312945/79b315a0-4518-4577-ad39-2f684cfec9cc)](https://www.youtube.com/watch?v=pY725Ya74VU)

![server-demo](https://user-images.githubusercontent.com/321592/148342937-1999d579-fd73-45d1-b206-53cbe4b0c87d.gif)

### Features

- HTTP Service for Master Administration / Displaying known servers
- Dynamic MOTDs
- IP Banning Client/Servers
- Polling other known masters and merging reported servers
- STUN compatability for reporting external IP address when game servers are run on the same host as the master
- Configurable Logging
- YAML config file

### Usage

```
$ ./mstrsvr --help
Usage of mstrsvr:
  -addadmin string
        add a new admin/password interactively to the admins list
  -passwd string
        updates the password for an existing admin interactively
  -rmadmin string
        remove an existing user from the admin list
```

### Instructions

1. Download the archive applicable to your architecture
2. You may delete the binaries not applicable to your operating system
3. Set up your admin account by running `./mstrsvr -addadmin <username>`
4. Lastly, run the binary `./mstrsvr`

The server is pre-configured with sane defaults see [mstrsvr.yaml.example](./blob/main/mstrsvr.yaml.example) for additional details on configuration options

### Templates

This project is equipped with a templating engine currently used with the HTTPD and Server MOTD

The Server MOTD field is optional and limited to a total (post evaluated) length of 255 characters
Template strings also optional

| Key            | Description                                                                                                                       |
|----------------|-----------------------------------------------------------------------------------------------------------------------------------|
| `{{.NL}}`      | new line (\n)                                                                                                                     |
| `{{.UserNum}}` | number of unique IPs that have requested a server list with in the past calendar day                                              |
| `{{.Time}}`    | local server time, see [the Carbon documentation](https://github.com/golang-module/carbon#format-sign-table) for more information |


The `{{.Time}}` format is defined [in the Carbon documentation](https://github.com/golang-module/carbon#format-sign-table) and defaults to `Y-m-d H:i:s T`


An example MOTD Template can be as followed:

`Welcome to a Testing server for Neo's Dummythiccness{{.NL}}You are currently the {{.UserNum}} user today.{{.NL}}Current local server time is: {{.Time}}`

Which, when evaluated would return:

```
Welcome to a testing server for Neo's Dummythiccness
You are currently the 69th user today.
Current local server time is: 2022-01-05 23:31-16 PST
```

![image](https://user-images.githubusercontent.com/321592/148345842-a0858edd-1264-4b60-a6ab-b42b056bd06f.png)