#! /bin/bash
test -e /foobar/non-existent/var/run/reboot-required
exit $?
