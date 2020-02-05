# goahead_client
Client program that asks the goahead service (https://github.com/xorpaul/goahead) if is allowed to restart

### Workflow

Client reads in its config file:

```
---
service_url: https://goahead-service.domain.tld/
service_url_ca_file: /etc/ssl/certs/optional-ca.pem
requesting_fqdn: foobar-server.domain.tld
restart_condition_script: /etc/goahead/needrestart
restart_condition_script_exit_code_for_reboot: 1
os_restart_hooks_dir: /etc/goahead/restart_hooks.d
```

Then it gets the system's uptime and sends this information to the goahead service:

```
{"fqdn":"foobar-server.domain.tld","uptime":"358h14m48s"}
```


and recieves a response:

```
{
  "timestamp": "2018-11-28T11:44:01.741651181Z",
  "go_ahead": false,
  "unknown_host": true,
  "request_id": "XymongTw",
  "found_cluster": "unknown",
  "requesting_fqdn": "foobar-server.domain.tld",
  "message": "FQDN foobar-server.domain.tld did not match any known cluster",
  "reported_uptime": "358h14m48s"
}
```

In this case the `restart_condition_script` did not exit with the configured `restart_condition_script_exit_code_for_reboot` exit code. 
So the client ask the service if it should restart, because of reasons only the service knows via the URI `/v1/inquire/restart/`

These checks are configured for each cluster via [reboot_goahead_checks](https://github.com/xorpaul/goahead/blob/master/cluster.go#L26)

```
$ goahead_client -debug
2018/11/28 11:44:01 Debug main(): Using as config file: /etc/goahead/client.yml
2018/11/28 11:44:01 Debug readConfigfile(): Trying to read config file: /etc/goahead/client.yml
2018/11/28 11:44:01 Debug setupHttpClient(): Appending certificate /etc/ssl/certs/optional-ca.pem to trusted CAs
2018/11/28 11:44:01 Debug ExecuteCommand(): Executing /etc/goahead/needrestart
2018/11/28 11:44:01 Executing /etc/goahead/needrestart took 0.03878s
2018/11/28 11:44:01 Did not find local reason to restart. Asking if I should restart, because of other reasons.
2018/11/28 11:44:01 Debug doRequest(): sending HTTP request https://goahead-service.domain.tld/v1/inquire/restart/
2018/11/28 11:44:01 Debug getPayload(): Trying to send payload: {"fqdn":"foobar-server.domain.tld","uptime":"358h14m48s"}
2018/11/28 11:44:01 Debug doRequest(): Received response: {"timestamp":"2018-11-28T11:44:01.741651181Z","go_ahead":false,"unknown_host":true,"request_id":"XymongTw","found_cluster":"unknown","requesting_fqdn":"foobar-server.domain.tld","message":"FQDN foobar-server.domain.tld did not match any known cluster","reported_uptime":"358h14m48s"}
```

Here none of the configured `reboot_goahead_checks` did return with the configured `reboot_goahead_checks_exit_code_for_reboot` so no reboot is necessary.


If the configured `restart_condition_script` did exit with the configured `restart_condition_script_exit_code_for_reboot` then the client requests a restart via the URI `/v1/request/restart/os`


```
{
   "found_cluster" : "foobar-servers",
   "go_ahead" : false,
   "reported_uptime" : "358h14m48s",
   "message" : "No previous request file found for fqdn: foobar-server.domain.tld",
   "ask_again_in" : "20s",
   "unknown_host" : false,
   "timestamp" : "2020-02-05T15:33:23.761812213Z",
   "request_id" : "KrXwoDxs",
   "requesting_fqdn" : "foobar-server.domain.tld"
}
```

In this case the the goahead service did not reject the restart request of the client, but tells it to ask again in a few seconds. This waiting time is chosen random to prevent race conditions in cluster nodes asking to reboot at the same time.
The client also need to send the content of the resonse filed `request_id` with this new request, so that the goahead service can verify that it is the same goahead_client process as the previous request.

In case everything worked, then the client recieves the important `"go_ahead" : true` in the response:
```
{
   "requesting_fqdn" : "foobar-server.domain.tld",
   "reported_uptime" : "358h14m48s",
   "request_id" : "uVBEdaBF",
   "go_ahead" : true,
   "found_cluster" : "foobar-servers",
   "ask_again_in" : "20s",
   "unknown_host" : false,
   "timestamp" : "2020-02-05T15:33:43.791804819Z"
}
```

This triggers then the scripts which are found in the configured `os_restart_hooks_dir`.

In this directory you can place different scripts which should be executed after the server recieved the goahead to reboot (notification scripts, silence monitoring, graceful shutdown, etc)
