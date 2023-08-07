#!/usr/bin/sh

APP_NAME=gfa
RUNTIME_PATH=$(cd `dirname $0` && pwd)
BIN_FILE=$RUNTIME_PATH/bin/${APP_NAME}
CFG_PATH=$RUNTIME_PATH/core.conf.toml
TCPDUMP_PROCESS=$(ps -ef | grep "tcpdump" | grep -v "grep")
LOG_PATH=/data/logs/${APP_NAME}
if [ ! -d ${LOG_PATH} ]; then
  mkdir -p ${LOG_PATH}
fi

function create_supervisor_file() {
  rm -f /etc/supervisord.d/${APP_NAME}.ini
	cat > /etc/supervisord.d/${APP_NAME}.ini <<EOF
[program:${APP_NAME}]
directory=${RUNTIME_PATH}
command=$BIN_FILE -conf=${CFG_PATH}
user=root
autostart=true
autorestart=true
startsecs=1
startretries=3
stopsignal=INT
stopwaitsecs=60
stdout_logfile=/data/logs/${APP_NAME}/console.out
stdout_logfile_maxbytes=128MB
stdout_logfile_backups=10
redirect_stderr=true
EOF
}

function stop() {
	sudo supervisorctl stop ${APP_NAME}

	if [ $? -ne 0 ]; then
		echo "Stop ${APP_NAME} failed"
		exit 1
	fi
	if [ -n "$TCPDUMP_PROCESS" ]; then
      pid=$(echo "$TCPDUMP_PROCESS" | awk '{print $2}')
      kill "$pid"
      echo "success kill tcpdump process (PID: $pid)"
  else
      echo "tcpdump process not found"
  fi
}

function start() {
	sudo supervisorctl status ${APP_NAME} |grep RUNNING > /dev/null

	if [ $? -eq 0 ]; then
		echo "${APP_NAME} is running"
		exit 1
	fi

	create_supervisor_file && (sudo supervisorctl update) && (sudo supervisorctl start ${APP_NAME})

	if [ $? -ne 0 ]; then
		echo "Start ${APP_NAME} failed"
		exit 1
	fi
}

if [ $# -lt 1 ]; then
	echo "usage: $0 [start|stop|restart](required) command(required) [config file]"
	exit 1
else
	if [ "$1" == 'stop' ] ; then
		stop && echo "Stop ${APP_NAME}: OK!"
	elif [ "$1" == 'start' ] ; then
		start && echo "Start ${APP_NAME}: OK!"
	elif [ "$1" == 'restart' ] ; then
		stop && start && echo "Restart ${APP_NAME}: OK!"
	else
		echo "usage: $0 [start|stop|restart](required) command(required) [config file]"
		exit 1
	fi
fi

exit 0