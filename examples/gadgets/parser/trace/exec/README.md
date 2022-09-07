# trace exec example (with filter and parser)

This example shows how to use different packages to create an
application to trace the process creation filtering by containers, using
an event parser to format the events.

This is is a continuation of
[trace/exec/withfilter](../../../withfilter/trace/exec/).

## How to build

```bash
$ go build .
```

## How to run

Start the tracer in a terminal.

```bash
$ sudo ./exec --container name foo
```

Create a `foo` container in another terminal:

```bash
$ sudo docker run --rm --name foo ubuntu bash -c "cat /dev/null && sleep 2"
```

The first terminal will print the processes created inside the new
container. It's important to notice that even the first processes in the
container are traced, bash in this case.

```bash
$ sudo ./exec --containername foo
CONTAINER        PID     PPID    PCOMM            RET  ARGS
foo              135452  135430  bash             0    /usr/bin/bash -c cat /dev/null && sleep 2
foo              135486  135452  cat              0    /usr/bin/cat /dev/null
foo              135452  135430  sleep            0    /usr/bin/sleep 2
```