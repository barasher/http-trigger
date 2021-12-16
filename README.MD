# http-trigger

**http-trigger** exposes REST services to execute shell commands.

## Getting `http-trigger`

- Binary releases : see [project releases](https://github.com/barasher/http-trigger/releases)
- Docker images : soon

## Configuration

```json
{
    "port": 8081,
    "commands": {
        "MyCommand": "/home/alfred/command.sh",
        "AnotherCommand": "/home/alfred/otherCommand.sh"
    }
}
```

- `port` : (optional, default : 8080) port that the werserver will be listening
- `commands` : exposed commands, dictionary where
  - Keys will be commands identifiers
  - Values will be the commands

## Execution

```shell
$ ./http-trigger -h
Usage:
  http-trigger [flags]

Flags:
  -c, --conf string   Configuration file
  -h, --help          help for http-trigger

$ ./main -c /tmp/nominal.json 
{"level":"info","time":"2021-12-16T21:46:57+01:00","message":"Server running..."}
```

## Usage

To execute a command : POST request on path `exec/[commandId]`

```shell
$ curl -X POST 127.0.0.1:8081/exec/MyCommand
```

Logging :

```shell
{"level":"debug","cmdId":"MyCommand","time":"2021-12-16T21:46:59+01:00","message":"Invocation..."}
{"level":"info","cmdId":"MyCommand","duration":1,"state":"ok","time":"2021-12-16T21:46:59+01:00","message":"Invocation complete"}
```

Response :
- body : sysout & syserr collected when exeuting command
- status :
  - 200 : command found and executed no matter what happens during execution
  - 404 : command not found
- headers :
  - `http-trigger-exit-code` : return code when command got executed
  - `http-trigger-duration` : duration of the execution (ms)

## Metrics

```shell
$ curl 127.0.0.1:8081/metrics
[...]
# HELP http_trigger_exec_MyCommand_request_duration_seconds Histogram concerning exec request durations (seconds) for command MyCommand
# TYPE http_trigger_exec_MyCommand_request_duration_seconds histogram
http_trigger_exec_MyCommand_request_duration_seconds_bucket{le="0.0025"} 1
http_trigger_exec_MyCommand_request_duration_seconds_bucket{le="0.005"} 1
http_trigger_exec_MyCommand_request_duration_seconds_bucket{le="0.01"} 1
http_trigger_exec_MyCommand_request_duration_seconds_bucket{le="0.025"} 1
http_trigger_exec_MyCommand_request_duration_seconds_bucket{le="0.05"} 1
http_trigger_exec_MyCommand_request_duration_seconds_bucket{le="0.1"} 1
http_trigger_exec_MyCommand_request_duration_seconds_bucket{le="+Inf"} 1
http_trigger_exec_MyCommand_request_duration_seconds_sum 0.001059764
http_trigger_exec_MyCommand_request_duration_seconds_count 1
[...]
```